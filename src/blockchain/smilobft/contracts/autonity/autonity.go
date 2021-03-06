package autonity

import (
	"errors"
	"math/big"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	"go-smilo/src/blockchain/smilobft/accounts/abi"
	"go-smilo/src/blockchain/smilobft/cmn"
	"go-smilo/src/blockchain/smilobft/consensus"
	"go-smilo/src/blockchain/smilobft/core/state"
	"go-smilo/src/blockchain/smilobft/core/types"
	"go-smilo/src/blockchain/smilobft/core/vm"
	"go-smilo/src/blockchain/smilobft/params"
)

func NewAutonityContract(
	bc Blockchainer,
	canTransfer func(db vm.StateDB, addr common.Address, amount *big.Int) bool,
	transfer func(db vm.StateDB, sender, recipient common.Address, amount, blockNumber *big.Int),
	GetHashFn func(ref *types.Header, chain ChainContext) func(n uint64) common.Hash,
) *Contract {
	return &Contract{
		bc:          bc,
		canTransfer: canTransfer,
		transfer:    transfer,
		GetHashFn:   GetHashFn,
		//SavedValidatorsRetriever: SavedValidatorsRetriever,
	}
}

type ChainContext interface {
	// Engine retrieves the chain's consensus engine.
	Engine() consensus.Engine

	// GetHeader returns the hash corresponding to their hash.
	GetHeader(common.Hash, uint64) *types.Header
}
type Blockchainer interface {
	ChainContext
	GetVMConfig() *vm.Config
	Config() *params.ChainConfig

	UpdateEnodeWhitelist(newWhitelist *types.Nodes)
	ReadEnodeWhitelist(EnableNodePermissionFlag bool) *types.Nodes

	UpdateBlacklist(newBlacklist *types.Nodes)
	ReadBlacklist(TxPoolBlacklistFlag bool) *types.Nodes
}

type Contract struct {
	address                  common.Address
	contractABI              *abi.ABI
	bc                       Blockchainer
	SavedValidatorsRetriever func(i uint64) ([]common.Address, error)
	metrics                  EconomicMetrics

	canTransfer func(db vm.StateDB, addr common.Address, amount *big.Int) bool
	transfer    func(db vm.StateDB, sender, recipient common.Address, amount, blockNumber *big.Int)
	GetHashFn   func(ref *types.Header, chain ChainContext) func(n uint64) common.Hash
	sync.RWMutex
}

// measure metrics of user's meta data by regarding of network economic.
func (ac *Contract) MeasureMetricsOfNetworkEconomic(header *types.Header, stateDB *state.StateDB) {
	if header == nil || stateDB == nil || header.Number.Uint64() < 1 {
		return
	}

	// prepare abi and evm context
	deployer := ac.bc.Config().AutonityContractConfig.Deployer
	sender := vm.AccountRef(deployer)
	gas := uint64(0xFFFFFFFF)
	evm := ac.getEVM(header, deployer, stateDB)

	ABI, err := ac.abi()
	if err != nil {
		return
	}

	// pack the function which dump the data from contract.
	input, err := ABI.Pack("dumpEconomicsMetricData")
	if err != nil {
		log.Warn("cannot pack the method: ", err.Error())
		return
	}

	// call evm.
	value := new(big.Int).SetUint64(0x00)
	ret, _, vmerr := evm.Call(sender, ac.Address(), input, gas, value, false)
	log.Debug("bytes return from contract: ", ret)
	if vmerr != nil {
		log.Warn("Error Autonity Contract dumpNetworkEconomics")
		return
	}

	// marshal the data from bytes arrays into specified structure.
	v := EconomicMetaData{make([]common.Address, 32), make([]uint8, 32), make([]*big.Int, 32),
		make([]*big.Int, 32), new(big.Int), new(big.Int)}

	if err := ABI.Unpack(&v, "dumpEconomicsMetricData", ret); err != nil { // can't work with aliased types
		log.Warn("Could not unpack dumpNetworkEconomicsData returned value", "err", err, "header.num",
			header.Number.Uint64())
		return
	}

	ac.metrics.SubmitEconomicMetrics(&v, stateDB, header.Number.Uint64(), ac.bc.Config().AutonityContractConfig.Operator)
}

//// Instantiates a new EVM object which is required when creating or calling a deployed contract
func (ac *Contract) getEVM(header *types.Header, origin common.Address, statedb *state.StateDB) *vm.EVM {
	coinbase, _ := types.Ecrecover(header)
	evmContext := vm.Context{
		CanTransfer: ac.canTransfer,
		Transfer:    ac.transfer,
		GetHash:     ac.GetHashFn(header, ac.bc),
		Origin:      origin,
		Coinbase:    coinbase,
		BlockNumber: header.Number,
		Time:        new(big.Int).SetUint64(header.Time),
		GasLimit:    header.GasLimit,
		Difficulty:  header.Difficulty,
		GasPrice:    new(big.Int).SetUint64(0x0),
	}
	vmConfig := *ac.bc.GetVMConfig()
	evm := vm.NewEVM(evmContext, statedb, statedb, ac.bc.Config(), vmConfig)
	return evm
}

// deployContract deploys the contract contained within the genesis field bytecode
func (ac *Contract) DeployAutonityContract(chain consensus.ChainReader, header *types.Header, statedb *state.StateDB) (common.Address, error) {
	// Convert the contract bytecode from hex into bytes
	contractBytecode := common.Hex2Bytes(chain.Config().AutonityContractConfig.Bytecode)
	evm := ac.getEVM(header, chain.Config().AutonityContractConfig.Deployer, statedb)
	sender := vm.AccountRef(chain.Config().AutonityContractConfig.Deployer)

	contractABI, err := ac.abi()
	if err != nil {
		log.Error("abi.JSON returns err", "err", err)
		return common.Address{}, err
	}

	ln := len(chain.Config().AutonityContractConfig.Users)
	users := make(cmn.Addresses, 0, ln)
	enodes := make([]string, 0, ln)
	accTypes := make([]*big.Int, 0, ln)
	participantStake := make([]*big.Int, 0, ln)
	for _, v := range chain.Config().AutonityContractConfig.Users {
		users = append(users, v.Address)
		enodes = append(enodes, v.Enode)
		accTypes = append(accTypes, big.NewInt(int64(v.Type.GetID())))
		participantStake = append(participantStake, big.NewInt(int64(v.Stake)))
	}

	//"" means contructor
	constructorParams, err := contractABI.Pack("",
		users,
		enodes,
		accTypes,
		participantStake,
		chain.Config().AutonityContractConfig.Operator,
		new(big.Int).SetUint64(chain.Config().AutonityContractConfig.MinGasPrice))
	if err != nil {
		log.Error("contractABI.Pack returns err", "err", err)
		return common.Address{}, err
	}

	//We need to append to data the constructor's parameters
	data := append(contractBytecode, constructorParams...)
	gas := uint64(0xFFFFFFFF)
	value := new(big.Int).SetUint64(0x00)

	// Deploy the Autonity contract
	_, contractAddress, _, vmerr := evm.Create(sender, data, gas, value, false)
	if vmerr != nil {
		log.Error("evm.Create returns err", "err", vmerr)
		return contractAddress, vmerr
	}
	ac.Lock()
	ac.address = contractAddress
	ac.Unlock()
	log.Info("Deployed Autonity Contract", "Address", contractAddress.String())

	return contractAddress, nil
}

func (ac *Contract) ContractGetValidators(chain consensus.ChainReader, header *types.Header, statedb *state.StateDB) ([]common.Address, error) {
	if header.Number.Cmp(big.NewInt(1)) == 0 && ac.SavedValidatorsRetriever != nil {
		return ac.SavedValidatorsRetriever(1)
	}
	sender := vm.AccountRef(chain.Config().AutonityContractConfig.Deployer)
	gas := uint64(0xFFFFFFFF)
	evm := ac.getEVM(header, chain.Config().AutonityContractConfig.Deployer, statedb)
	contractABI, err := ac.abi()
	if err != nil {
		return nil, err
	}

	input, err := contractABI.Pack("getValidators")
	if err != nil {
		return nil, err
	}

	value := new(big.Int).SetUint64(0x00)

	log.Debug("Autonity Contract GetValidators",
		"header Number", header.Number.Int64(),
		"Address", chain.Config().AutonityContractConfig.Deployer,
		"sender", sender,
		"gas", gas)

	//A standard call is issued - we leave the possibility to modify the state
	ret, _, vmerr := evm.Call(sender, ac.Address(), input, gas, value, false)
	if vmerr != nil {
		return nil, vmerr
	}

	var addresses []common.Address
	if err := contractABI.Unpack(&addresses, "getValidators", ret); err != nil { // can't work with aliased types
		msg := "Could not unpack getValidators returned value"
		log.Error(msg, "err", err)
		//panic(msg)
		return nil, err
	}

	sortableAddresses := cmn.Addresses(addresses)
	sort.Sort(sortableAddresses)
	return sortableAddresses, nil
}

var ErrAutonityContract = errors.New("could not call Autonity contract")

func (ac *Contract) UpdateEnodesWhitelist(state, vaultstate *state.StateDB, block *types.Block) error {
	newWhitelist, err := ac.GetWhitelist(block, state, vaultstate)
	if err != nil {
		log.Error("could not call contract", "err", err)
		return ErrAutonityContract
	}

	ac.bc.UpdateEnodeWhitelist(newWhitelist)
	return nil
}

func (ac *Contract) GetWhitelist(block *types.Block, db, vaultstate *state.StateDB) (*types.Nodes, error) {
	var (
		newWhitelist *types.Nodes
		err          error
	)

	if block.Number().Uint64() == 1 {
		// use genesis block whitelist
		newWhitelist = ac.bc.ReadEnodeWhitelist(false)
	} else {
		// call retrieveWhitelist contract function
		newWhitelist, err = ac.callGetWhitelist(db, vaultstate, block.Header())
	}

	return newWhitelist, err
}

//blockchain

func (ac *Contract) callGetWhitelist(state, vaultstate *state.StateDB, header *types.Header) (*types.Nodes, error) {
	// Needs to be refactored somehow
	deployer := ac.bc.Config().AutonityContractConfig.Deployer
	sender := vm.AccountRef(deployer)
	gas := uint64(0xFFFFFFFF)
	evm := ac.getEVM(header, deployer, state)

	ABI, err := ac.abi()
	if err != nil {
		return nil, err
	}

	input, err := ABI.Pack("getWhitelist")
	if err != nil {
		return nil, err
	}

	ret, _, vmerr := evm.StaticCall(sender, ac.Address(), input, gas, false)
	if vmerr != nil {
		log.Error("Error Autonity Contract getWhitelist()")
		return nil, vmerr
	}

	var returnedEnodes []string
	if err := ABI.Unpack(&returnedEnodes, "getWhitelist", ret); err != nil { // can't work with aliased types
		log.Error("Could not unpack getWhitelist returned value")
		return nil, err
	}

	return types.NewNodes(returnedEnodes, false), nil
}

func (ac *Contract) GetMinimumGasPrice(block *types.Block, db, vaultstate *state.StateDB) (uint64, error) {
	if block.Number().Uint64() <= 1 {
		return ac.bc.Config().AutonityContractConfig.MinGasPrice, nil
	}

	return ac.callGetMinimumGasPrice(db, vaultstate, block.Header())
}

func (ac *Contract) SetMinimumGasPrice(block *types.Block, db, vaultstate *state.StateDB, price *big.Int) error {
	if block.Number().Uint64() <= 1 {
		return nil
	}

	return ac.callSetMinimumGasPrice(db, vaultstate, block.Header(), price)
}

func (ac *Contract) callGetMinimumGasPrice(state, vaultstate *state.StateDB, header *types.Header) (uint64, error) {
	// Needs to be refactored somehow
	deployer := ac.bc.Config().AutonityContractConfig.Deployer
	sender := vm.AccountRef(deployer)
	gas := uint64(0xFFFFFFFF)
	evm := ac.getEVM(header, deployer, state)

	ABI, err := ac.abi()
	if err != nil {
		return 0, err
	}

	input, err := ABI.Pack("getMinimumGasPrice")
	if err != nil {
		return 0, err
	}

	value := new(big.Int).SetUint64(0x00)
	ret, _, vmerr := evm.Call(sender, ac.Address(), input, gas, value, false)
	if vmerr != nil {
		log.Error("Error Autonity Contract getMinimumGasPrice()")
		return 0, vmerr
	}

	minGasPrice := new(big.Int)
	if err := ABI.Unpack(&minGasPrice, "getMinimumGasPrice", ret); err != nil { // can't work with aliased types
		log.Error("Could not unpack minGasPrice returned value", "err", err, "header.num", header.Number.Uint64())
		return 0, err
	}

	return minGasPrice.Uint64(), nil
}

func (ac *Contract) callSetMinimumGasPrice(state, vaultstate *state.StateDB, header *types.Header, price *big.Int) error {
	// Needs to be refactored somehow
	deployer := ac.bc.Config().AutonityContractConfig.Deployer
	sender := vm.AccountRef(deployer)
	gas := uint64(0xFFFFFFFF)
	evm := ac.getEVM(header, deployer, state)

	ABI, err := ac.abi()
	if err != nil {
		return err
	}

	input, err := ABI.Pack("setMinimumGasPrice")
	if err != nil {
		return err
	}

	_, _, vmerr := evm.Call(sender, ac.Address(), input, gas, price, false)
	if vmerr != nil {
		log.Error("Error Autonity Contract getMinimumGasPrice()")
		return vmerr
	}
	return nil
}

func (ac *Contract) PerformRedistribution(header *types.Header, db *state.StateDB, gasUsed *big.Int) error {
	if header.Number.Uint64() <= 1 {
		return nil
	}
	return ac.callPerformRedistribution(db, header, gasUsed)
}

func (ac *Contract) callPerformRedistribution(state *state.StateDB, header *types.Header, blockGas *big.Int) error {
	// Needs to be refactored somehow
	deployer := ac.bc.Config().AutonityContractConfig.Deployer

	sender := vm.AccountRef(deployer)
	gas := uint64(0xFFFFFFFF)
	evm := ac.getEVM(header, deployer, state)

	ABI, err := ac.abi()
	if err != nil {
		return err
	}

	input, err := ABI.Pack("performRedistribution", blockGas)
	if err != nil {
		log.Error("Error Autonity Contract callPerformRedistribution()", "err", err)
		return err
	}

	value := new(big.Int).SetUint64(0x00)

	ret, _, vmerr := evm.Call(sender, ac.Address(), input, gas, value, false)
	if vmerr != nil {
		log.Error("Error Autonity Contract callPerformRedistribution()", "err", err)
		return vmerr
	}

	// after reward distribution, update metrics with the return values.
	//v := RewardDistributionMetaData {true, make([]common.Address, 32), make([]*big.Int, 32), new(big.Int)}
	v := RewardDistributionMetaData{}
	v.Result = true
	v.Holders = make([]common.Address, 32)
	v.Rewardfractions = make([]*big.Int, 32)
	v.Amount = new(big.Int)

	if err := ABI.Unpack(&v, "performRedistribution", ret); err != nil { // can't work with aliased types
		log.Error("Could not unpack performRedistribution returned value", "err", err, "header.num", header.Number.Uint64())
		return nil
	}

	ac.metrics.SubmitRewardDistributionMetrics(&v, header.Number.Uint64())
	return nil
}

func (ac *Contract) ApplyPerformRedistribution(transactions types.Transactions, receipts types.Receipts, header *types.Header, statedb *state.StateDB) error {
	log.Info("ApplyPerformRedistribution", "header", header.Number.Uint64())
	if header.Number.Cmp(big.NewInt(1)) < 1 {
		return nil
	}
	blockGas := new(big.Int)
	for i, tx := range transactions {
		blockGas.Add(blockGas, new(big.Int).Mul(tx.GasPrice(), new(big.Int).SetUint64(receipts[i].GasUsed)))
	}

	log.Info("execution start ApplyPerformRedistribution", "balance", statedb.GetBalance(ac.Address()), "block", header.Number.Uint64(), "gas", blockGas.Uint64())
	if blockGas.Cmp(new(big.Int)) == 0 {
		log.Info("execution start ApplyPerformRedistribution with 0 gas", "balance", statedb.GetBalance(ac.Address()), "block", header.Number.Uint64())
		return nil
	}
	return ac.PerformRedistribution(header, statedb, blockGas)
}

func (ac *Contract) Address() common.Address {
	if reflect.DeepEqual(ac.address, common.Address{}) {
		addr, err := ac.bc.Config().AutonityContractConfig.GetContractAddress()
		if err != nil {
			log.Error("Cant get contract address", "err", err)
		}
		return addr
	}
	return ac.address
}

func (ac *Contract) abi() (*abi.ABI, error) {
	ac.Lock()
	defer ac.Unlock()
	if ac.contractABI != nil {
		return ac.contractABI, nil
	}
	ABI, err := abi.JSON(strings.NewReader(ac.bc.Config().AutonityContractConfig.ABI))
	if err != nil {
		return nil, err
	}
	ac.contractABI = &ABI
	return ac.contractABI, nil

}
