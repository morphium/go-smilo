// Copyright 2019 The go-smilo Authors
// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package backend

import (
	"bytes"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"

	"go-smilo/src/blockchain/smilobft/consensus"
	"go-smilo/src/blockchain/smilobft/consensus/sport"
	"go-smilo/src/blockchain/smilobft/consensus/sport/smilobftcore"
	"go-smilo/src/blockchain/smilobft/core/state"
	"go-smilo/src/blockchain/smilobft/core/types"
)

// verifySigner checks whether the signer is in parent's fullnode set
func (sb *backend) verifySigner(chain consensus.ChainReader, header *types.Header, parents []*types.Header) error {
	// Verifying the genesis block is not supported
	number := header.Number.Uint64()
	if number == 0 {
		return errUnknownBlock
	}

	// Retrieve the snapshot needed to verify this header and cache it
	snap, err := sb.snapshot(chain, number-1, header.ParentHash, parents)
	if err != nil {
		return err
	}

	// resolve the authorization key and check against signers
	signer, err := ecrecover(header)
	if err != nil {
		return err
	}

	// Signer should be in the fullnode set of previous block's extraData.
	if _, v := snap.FullnodeSet.GetByAddress(signer); v == nil {
		return errUnauthorized
	}
	return nil
}

// verifyCommittedSeals checks whether every committed seal is signed by one of the parent's fullnodes
func (sb *backend) verifyCommittedSeals(chain consensus.ChainReader, header *types.Header, parents []*types.Header) error {
	number := header.Number.Uint64()
	// We don't need to verify committed seals in the genesis block
	if number == 0 {
		return nil
	}

	// Retrieve the snapshot needed to verify this header and cache it
	snap, err := sb.snapshot(chain, number-1, header.ParentHash, parents)
	if err != nil {
		return err
	}

	extra, err := types.ExtractSportExtra(header)
	if err != nil {
		return err
	}
	// The length of Committed seals should be larger than 0
	if len(extra.CommittedSeal) == 0 {
		return errEmptyCommittedSeals
	}

	fullnodes := snap.FullnodeSet.Copy()
	// Check whether the committed seals are generated by parent's fullnodes
	validSeal := 0.0
	proposalSeal := smilobftcore.PrepareCommittedSeal(header.Hash())
	// 1. Get committed seals from current header
	for _, seal := range extra.CommittedSeal {
		// 2. Get the original address by seal and parent block hash
		addr, err := sport.GetSignatureAddress(proposalSeal, seal)
		if err != nil {
			sb.logger.Error("not a valid address", "err", err)
			return errInvalidSignature
		}
		// Every fullnode can have only one seal. If more than one seals are signed by a
		// fullnode, the fullnode cannot be found and errInvalidCommittedSeals is returned.
		if fullnodes.RemoveFullnode(addr) {
			validSeal += 1
		} else {
			return errInvalidCommittedSeals
		}
	}

	// The length of validSeal should be larger than number of faulty node + 1
	if validSeal < 2*snap.FullnodeSet.F() {
		sb.logger.Error("The length of validSeal should be larger or eq than number of 2x faulty nodes", "validSeal", "2*snap.FullnodeSet.F()", 2*snap.FullnodeSet.F())
		return errInvalidCommittedSeals
	}

	return nil
}

// AccumulateRewards (override from ethash) credits the coinbase of the given block with the mining reward.
// The total reward consists of the static block reward and rewards for  the community.
func AccumulateRewards(communityAddress string, state *state.StateDB, header *types.Header) {
	// add reward based on chain progression
	blockReward := getSmiloBlockReward(header.Number)

	// Accumulate the rewards
	reward := new(big.Int).Set(blockReward)
	emptryAddress := common.Address{}
	if header.Coinbase != emptryAddress {

		log.Info("$$$$$$$$$$$$$$$$$$$$$ AccumulateRewards, block: ", "blockNum", header.Number.Int64(), "BlockReward", blockReward.Int64(), "Coinbase", header.Coinbase.Hex())

		// Accumulate the rewards to community
		if communityAddress != "" {
			rewardForCommunity := new(big.Int).Div(blockReward, big.NewInt(4))
			state.AddBalance(common.HexToAddress(communityAddress), rewardForCommunity, header.Number)
			log.Info("$$$$$$$$$$$$$$$$$$$$$ AccumulateRewards, adding reward to community ", "rewardForCommunity", rewardForCommunity, "communityAddress", communityAddress)
		}
		state.AddBalance(header.Coinbase, reward, header.Number)
	}
}

func getSmiloBlockReward(blockNum *big.Int) (blockReward *big.Int) {
	blockReward = new(big.Int)
	for maxBlockRange, reward := range smiloTokenMetricsTable {
		if blockNum.Cmp(maxBlockRange) == -1 {
			if reward.Cmp(blockReward) > 0 {
				blockReward = reward
			}
		}
	}
	return blockReward
}

// update timestamp and signature of the block based on its number of transactions
func (sb *backend) updateBlock(parent *types.Header, block *types.Block) (*types.Block, error) {
	header := block.Header()
	// sign the hash
	seal, err := sb.Sign(sigHash(header).Bytes())
	if err != nil {
		return nil, err
	}

	err = writeSeal(header, seal)
	if err != nil {
		return nil, err
	}

	return block.WithSeal(header), nil
}

// Start implements consensus.Sport.Start
func (sb *backend) Start(chain consensus.ChainReader, currentBlock func() *types.Block, hasBadBlock func(hash common.Hash) bool) error {
	sb.coreMu.Lock()
	defer sb.coreMu.Unlock()
	if sb.coreStarted {
		return consensus.ErrStartedEngine
	}

	// clear previous data
	sb.proposedBlockHash = common.Hash{}
	if sb.commitCh != nil {
		close(sb.commitCh)
	}
	sb.commitCh = make(chan *types.Block, 1)

	sb.chain = chain
	sb.currentBlock = currentBlock
	sb.hasBadBlock = hasBadBlock

	if err := sb.core.Start(); err != nil {
		return err
	}

	sb.coreStarted = true
	return nil
}

// Stop implements consensus.Sport.Stop
func (sb *backend) Stop() error {
	sb.coreMu.Lock()
	defer sb.coreMu.Unlock()
	if !sb.coreStarted {
		return sport.ErrStoppedEngine
	}
	if err := sb.core.Stop(); err != nil {
		return err
	}
	sb.coreStarted = false
	return nil
}

// ecrecover extracts the Ethereum account address from a signed header.
func ecrecover(header *types.Header) (common.Address, error) {
	hash := header.Hash()
	if addr, ok := recentAddresses.Get(hash); ok {
		return addr.(common.Address), nil
	}

	// Retrieve the signature from the header extra-data
	sportExtra, err := types.ExtractSportExtra(header)
	if err != nil {
		return common.Address{}, err
	}

	addr, err := sport.GetSignatureAddress(sigHash(header).Bytes(), sportExtra.Seal)
	if err != nil {
		return addr, err
	}
	recentAddresses.Add(hash, addr)
	return addr, nil
}

// prepareExtra returns a extra-data of the given header and fullnodes
func prepareExtra(header *types.Header, vals []common.Address) ([]byte, error) {
	var buf bytes.Buffer

	// compensate the lack bytes if header.Extra is not enough SportExtraVanity bytes.
	if len(header.Extra) < types.SportExtraVanity {
		header.Extra = append(header.Extra, bytes.Repeat([]byte{0x00}, types.SportExtraVanity-len(header.Extra))...)
	}
	buf.Write(header.Extra[:types.SportExtraVanity])

	ist := &types.SportExtra{
		Fullnodes:     vals,
		Seal:          []byte{},
		CommittedSeal: [][]byte{},
	}

	payload, err := rlp.EncodeToBytes(&ist)
	if err != nil {
		return nil, err
	}

	return append(buf.Bytes(), payload...), nil
}

// writeSeal writes the extra-data field of the given header with the given seals.
// suggest to rename to writeSeal.
func writeSeal(h *types.Header, seal []byte) error {
	if len(seal)%types.SportExtraSeal != 0 {
		return errInvalidSignature
	}

	sportExtra, err := types.ExtractSportExtra(h)
	if err != nil {
		return err
	}

	sportExtra.Seal = seal
	payload, err := rlp.EncodeToBytes(&sportExtra)
	if err != nil {
		return err
	}

	h.Extra = append(h.Extra[:types.SportExtraVanity], payload...)
	return nil
}

// writeCommittedSeals writes the extra-data field of a block header with given committed seals.
func writeCommittedSeals(h *types.Header, committedSeals [][]byte) error {
	if len(committedSeals) == 0 {
		return errInvalidCommittedSeals
	}

	for _, seal := range committedSeals {
		if len(seal) != types.SportExtraSeal {
			return errInvalidCommittedSeals
		}
	}

	sportExtra, err := types.ExtractSportExtra(h)
	if err != nil {
		return err
	}

	sportExtra.CommittedSeal = make([][]byte, len(committedSeals))
	copy(sportExtra.CommittedSeal, committedSeals)

	payload, err := rlp.EncodeToBytes(&sportExtra)
	if err != nil {
		return err
	}

	h.Extra = append(h.Extra[:types.SportExtraVanity], payload...)
	return nil
}