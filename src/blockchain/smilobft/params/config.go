// Copyright 2019 The go-smilo Authors
// Copyright 2016 The go-ethereum Authors
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

package params

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/common"
)

// Genesis hashes to enforce below configs on.
var (
	MainnetGenesisHash = common.HexToHash("0x60a23d7e3337a9e3694af178e629cbe07cf160af3eef711e7f1c9405f38c19ab")
	TestnetGenesisHash = common.HexToHash("0x853e0fbdc73a57c3f74b5716ce778fe0dd45025d00c3f016b9cc33bac7b0d92e")
	RinkebyGenesisHash = common.HexToHash("0x6341fd3daf94b748c72ced5a5b26028f2474f5f00d824504e4fa37a75767e177")
	GoerliGenesisHash  = common.HexToHash("0xbf7e331f7f7c1dd2e05159666b3bf8bc7a8a3a9eb1d518969eab529dd9b88c1a")

	// Glienicke Default config
	GlienickeDefaultABI      = `[{"constant":false,"inputs":[{"name":"_enode","type":"string"}],"name":"RemoveEnode","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"","type":"uint256"}],"name":"enodes","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_enode","type":"string"}],"name":"AddEnode","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"getWhitelist","outputs":[{"name":"","type":"string[]"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"s1","type":"string"},{"name":"s2","type":"string"}],"name":"compareStringsbyBytes","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"pure","type":"function"},{"inputs":[{"name":"_genesisEnodes","type":"string[]"}],"payable":false,"stateMutability":"nonpayable","type":"constructor"}]`
	GlienickeDefaultBytecode = "60806040523480156200001157600080fd5b5060405162000c8738038062000c87833981018060405262000037919081019062000209565b60005b81518110156200009157600082828151811015156200005557fe5b6020908102909101810151825460018101808555600094855293839020825162000086949190920192019062000099565b50506001016200003a565b5050620002ec565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f10620000dc57805160ff19168380011785556200010c565b828001600101855582156200010c579182015b828111156200010c578251825591602001919060010190620000ef565b506200011a9291506200011e565b5090565b6200013b91905b808211156200011a576000815560010162000125565b90565b6000601f820183136200015057600080fd5b815162000167620001618262000270565b62000249565b81815260209384019390925082018360005b83811015620001a95781518601620001928882620001b3565b845250602092830192919091019060010162000179565b5050505092915050565b6000601f82018313620001c557600080fd5b8151620001d6620001618262000291565b91508082526020830160208301858383011115620001f357600080fd5b62000200838284620002b9565b50505092915050565b6000602082840312156200021c57600080fd5b81516001604060020a038111156200023357600080fd5b62000241848285016200013e565b949350505050565b6040518181016001604060020a03811182821017156200026857600080fd5b604052919050565b60006001604060020a038211156200028757600080fd5b5060209081020190565b60006001604060020a03821115620002a857600080fd5b506020601f91909101601f19160190565b60005b83811015620002d6578181015183820152602001620002bc565b83811115620002e6576000848401525b50505050565b61098b80620002fc6000396000f30060806040526004361061006c5763ffffffff7c0100000000000000000000000000000000000000000000000000000000600035041663189da37281146100715780639890b32914610093578063ba1de5da146100c9578063d01f63f5146100e9578063e9a734ff1461010b575b600080fd5b34801561007d57600080fd5b5061009161008c36600461070e565b610138565b005b34801561009f57600080fd5b506100b36100ae3660046107b4565b610280565b6040516100c0919061089c565b60405180910390f35b3480156100d557600080fd5b506100916100e436600461070e565b610327565b3480156100f557600080fd5b506100fe61036d565b6040516100c0919061087d565b34801561011757600080fd5b5061012b61012636600461074b565b610446565b6040516100c0919061088e565b6000805460011061014857600080fd5b5060005b60005481101561027c5761020160008281548110151561016857fe5b600091825260209182902001805460408051601f60026000196101006001871615020190941693909304928301859004850281018501909152818152928301828280156101f65780601f106101cb576101008083540402835291602001916101f6565b820191906000526020600020905b8154815290600101906020018083116101d957829003601f168201915b505050505083610446565b156102745760008054600019810190811061021857fe5b9060005260206000200160008281548110151561023157fe5b90600052602060002001908054600181600116156101000203166002900461025a92919061050c565b50600080549061026e906000198301610591565b5061027c565b60010161014c565b5050565b600080548290811061028e57fe5b600091825260209182902001805460408051601f600260001961010060018716150201909416939093049283018590048502810185019091528181529350909183018282801561031f5780601f106102f45761010080835404028352916020019161031f565b820191906000526020600020905b81548152906001019060200180831161030257829003601f168201915b505050505081565b60008054600181018083559180528251610368917f290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e563019060208501906105b5565b505050565b60606000805480602002602001604051908101604052809291908181526020016000905b8282101561043c5760008481526020908190208301805460408051601f60026000196101006001871615020190941693909304928301859004850281018501909152818152928301828280156104285780601f106103fd57610100808354040283529160200191610428565b820191906000526020600020905b81548152906001019060200180831161040b57829003601f168201915b505050505081526020019060010190610391565b5050505090505b90565b6000816040518082805190602001908083835b602083106104785780518252601f199092019160209182019101610459565b51815160209384036101000a6000190180199092169116179052604051919093018190038120885190955088945090928392508401908083835b602083106104d15780518252601f1990920191602091820191016104b2565b6001836020036101000a0380198251168184511680821785525050505050509050019150506040518091039020600019161490505b92915050565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f106105455780548555610581565b8280016001018555821561058157600052602060002091601f016020900482015b82811115610581578254825591600101919060010190610566565b5061058d929150610623565b5090565b8154818355818111156103685760008381526020902061036891810190830161063d565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f106105f657805160ff1916838001178555610581565b82800160010185558215610581579182015b82811115610581578251825591602001919060010190610608565b61044391905b8082111561058d5760008155600101610629565b61044391905b8082111561058d5760006106578282610660565b50600101610643565b50805460018160011615610100020316600290046000825580601f1061068657506106a4565b601f0160209004906000526020600020908101906106a49190610623565b50565b6000601f820183136106b857600080fd5b81356106cb6106c6826108d4565b6108ad565b915080825260208301602083018583830111156106e757600080fd5b6106f283828461090b565b50505092915050565b60006107078235610443565b9392505050565b60006020828403121561072057600080fd5b813567ffffffffffffffff81111561073757600080fd5b610743848285016106a7565b949350505050565b6000806040838503121561075e57600080fd5b823567ffffffffffffffff81111561077557600080fd5b610781858286016106a7565b925050602083013567ffffffffffffffff81111561079e57600080fd5b6107aa858286016106a7565b9150509250929050565b6000602082840312156107c657600080fd5b600061074384846106fb565b60006107dd82610902565b808452602084019350836020820285016107f6856108fc565b60005b8481101561082d578383038852610811838351610848565b925061081c826108fc565b6020989098019791506001016107f9565b50909695505050505050565b61084281610906565b82525050565b600061085382610902565b808452610867816020860160208601610917565b61087081610947565b9093016020019392505050565b6020808252810161070781846107d2565b602081016105068284610839565b602080825281016107078184610848565b60405181810167ffffffffffffffff811182821017156108cc57600080fd5b604052919050565b600067ffffffffffffffff8211156108eb57600080fd5b506020601f91909101601f19160190565b60200190565b5190565b151590565b82818337506000910152565b60005b8381101561093257818101518382015260200161091a565b83811115610941576000848401525b50505050565b601f01601f1916905600a265627a7a72305820964f5cc22f1190ee37bff2da4fb93a5ec323c47cf5f68a08c1483c225b9fb26d6c6578706572696d656e74616cf50037"
	GlienickeDefaultDeployer = common.BytesToAddress([]byte{13, 37})
)

// TrustedCheckpoints associates each known checkpoint with the genesis hash of
// the chain it belongs to.
var TrustedCheckpoints = map[common.Hash]*TrustedCheckpoint{
	MainnetGenesisHash: nil,
	TestnetGenesisHash: nil,
	RinkebyGenesisHash: RinkebyTrustedCheckpoint,
	GoerliGenesisHash:  GoerliTrustedCheckpoint,
}

// CheckpointOracles associates each known checkpoint oracles with the genesis hash of
// the chain it belongs to.
var CheckpointOracles = map[common.Hash]*CheckpointOracleConfig{
	MainnetGenesisHash: nil,
	TestnetGenesisHash: nil,
	RinkebyGenesisHash: RinkebyCheckpointOracle,
	GoerliGenesisHash:  GoerliCheckpointOracle,
}

var (
	// MainnetChainConfig is the chain parameters to run a node on the main network.  THIS IS VALID ONLY FOR ETH MAINNET
	MainnetChainConfig = &ChainConfig{
		ChainID:             big.NewInt(1),
		HomesteadBlock:      big.NewInt(1150000),
		DAOForkBlock:        big.NewInt(1920000),
		DAOForkSupport:      true,
		EIP150Block:         big.NewInt(2463000),
		EIP150Hash:          common.HexToHash("0x2086799aeebeae135c246c65021c82b4e15a2c451340993aacfd2751886514f0"),
		EIP155Block:         big.NewInt(2675000),
		EIP158Block:         big.NewInt(2675000),
		ByzantiumBlock:      big.NewInt(4370000),
		ConstantinopleBlock: big.NewInt(7280000),
		PetersburgBlock:     big.NewInt(7280000),
		IstanbulBlock:       nil,
		Ethash:              new(EthashConfig),
	}

	// MainnetTrustedCheckpoint contains the light client trusted checkpoint for the main network.
	//MainnetTrustedCheckpoint = &TrustedCheckpoint{
	//	SectionIndex: 246,
	//	SectionHead:  common.HexToHash("0xb86fbe8a2b1f3c576d06fe1721cd976f98ac1cbf1823da16ef74811e85fd44ac"),
	//	CHTRoot:      common.HexToHash("0xe99b397f908a391d0d6bd41d1c19cea4bf5051a9695c94d58de44c538d7a1037"),
	//	BloomRoot:    common.HexToHash("0xa1c1e064ccc16690c5fbabf600c4c7ebb2d8e8fcc674e59365087a77fb391a47"),
	//}

	// MainnetCheckpointOracle contains a set of configs for the main network oracle.
	//MainnetCheckpointOracle = &CheckpointOracleConfig{
	//	Address: common.HexToAddress("0x9a9070028361F7AAbeB3f2F2Dc07F82C4a98A02a"),
	//	Signers: []common.Address{
	//		common.HexToAddress("0x1b2C260efc720BE89101890E4Db589b44E950527"), // Peter
	//		common.HexToAddress("0x78d1aD571A1A09D60D9BBf25894b44e4C8859595"), // Martin
	//		common.HexToAddress("0x286834935f4A8Cfb4FF4C77D5770C2775aE2b0E7"), // Zsolt
	//		common.HexToAddress("0xb86e2B0Ab5A4B1373e40c51A7C712c70Ba2f9f8E"), // Gary
	//		common.HexToAddress("0x0DF8fa387C602AE62559cC4aFa4972A7045d6707"), // Guillaume
	//	},
	//	Threshold: 2,
	//}

	// TestnetChainConfig contains the chain parameters to run a node on the Ropsten test network.
	TestnetChainConfig = &ChainConfig{
		ByzantiumBlock:       big.NewInt(1),
		EIP150Block:          big.NewInt(2),
		EIP150Hash:           common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
		EIP155Block:          big.NewInt(0),
		EIP158Block:          big.NewInt(3),
		PetersburgBlock:      big.NewInt(4),
		ConstantinopleBlock:  big.NewInt(5),
		SixtySixPercentBlock: big.NewInt(310000),

		Sport: &SportConfig{
			Epoch:         30000,
			SpeakerPolicy: 0,
			MinFunds:      10000,
		},

		IsSmilo:       true,
		IsGas:         true,
		IsGasRefunded: true,

		ChainID: big.NewInt(10),

		RequiredMinFunds: 1,
		IstanbulBlock:    nil,
		Ethash:           new(EthashConfig),
	}

	// TestnetTrustedCheckpoint contains the light client trusted checkpoint for the Ropsten test network.
	//TestnetTrustedCheckpoint = &TrustedCheckpoint{
	//Name: "testnet",
	//SectionIndex: 148,
	//SectionHead:  common.HexToHash("0x4d3181bedb6aa96a6f3efa866c71f7802400d0fb4a6906946c453630d850efc0"),
	//CHTRoot:      common.HexToHash("0x25df2f9d63a5f84b2852988f0f0f7af5a7877da061c11b85c812780b5a27a5ec"),
	//BloomRoot:    common.HexToHash("0x0584834e5222471a06c669d210e302ca602780eaaddd04634fd65471c2a91419"),
	//}

	// TestnetCheckpointOracle contains a set of configs for the Ropsten test network oracle.
	//TestnetCheckpointOracle = &CheckpointOracleConfig{
	//	Address: common.HexToAddress("0xEF79475013f154E6A65b54cB2742867791bf0B84"),
	//	Signers: []common.Address{
	//		common.HexToAddress("0x32162F3581E88a5f62e8A61892B42C46E2c18f7b"), // Peter
	//		common.HexToAddress("0x78d1aD571A1A09D60D9BBf25894b44e4C8859595"), // Martin
	//		common.HexToAddress("0x286834935f4A8Cfb4FF4C77D5770C2775aE2b0E7"), // Zsolt
	//		common.HexToAddress("0xb86e2B0Ab5A4B1373e40c51A7C712c70Ba2f9f8E"), // Gary
	//		common.HexToAddress("0x0DF8fa387C602AE62559cC4aFa4972A7045d6707"), // Guillaume
	//	},
	//	Threshold: 2,
	//}

	// RinkebyChainConfig contains the chain parameters to run a node on the Rinkeby test network.
	RinkebyChainConfig = &ChainConfig{
		ChainID:             big.NewInt(4),
		HomesteadBlock:      big.NewInt(1),
		DAOForkBlock:        nil,
		DAOForkSupport:      true,
		EIP150Block:         big.NewInt(2),
		EIP150Hash:          common.HexToHash("0x9b095b36c15eaf13044373aef8ee0bd3a382a5abb92e402afa44b8249c3a90e9"),
		EIP155Block:         big.NewInt(3),
		EIP158Block:         big.NewInt(3),
		ByzantiumBlock:      big.NewInt(1035301),
		ConstantinopleBlock: big.NewInt(3660663),
		PetersburgBlock:     big.NewInt(4321234),
		IstanbulBlock:       nil,
		Clique: &CliqueConfig{
			Period: 15,
			Epoch:  30000,
		},
	}

	// RinkebyTrustedCheckpoint contains the light client trusted checkpoint for the Rinkeby test network.
	RinkebyTrustedCheckpoint = &TrustedCheckpoint{
		SectionIndex: 142,
		SectionHead:  common.HexToHash("0xf7e3946d54c3040d391edd61a855fec7293f9d0b51445ede88562f2dc2edce3f"),
		CHTRoot:      common.HexToHash("0xb2beee185e3ecada83eb69f72cbcca3e0978dbc8da5cdb3e34a71b3d597815d0"),
		BloomRoot:    common.HexToHash("0x3970039fee31eb0542090030d1567cc99b8051572d51899db4d91619ca26f0cb"),
	}

	// RinkebyCheckpointOracle contains a set of configs for the Rinkeby test network oracle.
	RinkebyCheckpointOracle = &CheckpointOracleConfig{
		Address: common.HexToAddress("0xebe8eFA441B9302A0d7eaECc277c09d20D684540"),
		Signers: []common.Address{
			common.HexToAddress("0xd9c9cd5f6779558b6e0ed4e6acf6b1947e7fa1f3"), // Peter
			common.HexToAddress("0x78d1aD571A1A09D60D9BBf25894b44e4C8859595"), // Martin
			common.HexToAddress("0x286834935f4A8Cfb4FF4C77D5770C2775aE2b0E7"), // Zsolt
			common.HexToAddress("0xb86e2B0Ab5A4B1373e40c51A7C712c70Ba2f9f8E"), // Gary
		},
		Threshold: 2,
	}

	// GoerliChainConfig contains the chain parameters to run a node on the Görli test network.
	GoerliChainConfig = &ChainConfig{
		ChainID:             big.NewInt(5),
		HomesteadBlock:      big.NewInt(0),
		DAOForkBlock:        nil,
		DAOForkSupport:      true,
		EIP150Block:         big.NewInt(0),
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:       nil,
		Clique: &CliqueConfig{
			Period: 15,
			Epoch:  30000,
		},
	}

	// GoerliTrustedCheckpoint contains the light client trusted checkpoint for the Görli test network.
	GoerliTrustedCheckpoint = &TrustedCheckpoint{
		SectionIndex: 26,
		SectionHead:  common.HexToHash("0xd0c206e064c8efea930d97e56786af95354ea481b35294a20e5a340937e4c2c9"),
		CHTRoot:      common.HexToHash("0xce7235999aa8d73c4493b8f397474dafc627652a790dec60c4a968e2dfa5d7be"),
		BloomRoot:    common.HexToHash("0xc1ac19553473ebb07325b5092a09277d62e9ffe159166a1c6fbec614c4dfd885"),
	}

	// GoerliCheckpointOracle contains a set of configs for the Goerli test network oracle.
	GoerliCheckpointOracle = &CheckpointOracleConfig{
		Address: common.HexToAddress("0x18CA0E045F0D772a851BC7e48357Bcaab0a0795D"),
		Signers: []common.Address{
			common.HexToAddress("0x4769bcaD07e3b938B7f43EB7D278Bc7Cb9efFb38"), // Peter
			common.HexToAddress("0x78d1aD571A1A09D60D9BBf25894b44e4C8859595"), // Martin
			common.HexToAddress("0x286834935f4A8Cfb4FF4C77D5770C2775aE2b0E7"), // Zsolt
			common.HexToAddress("0xb86e2B0Ab5A4B1373e40c51A7C712c70Ba2f9f8E"), // Gary
			common.HexToAddress("0x0DF8fa387C602AE62559cC4aFa4972A7045d6707"), // Guillaume
		},
		Threshold: 2,
	}

	// SportChainConfig contains the chain parameters to run a node on the Sport test network.
	SportChainConfig = &ChainConfig{

		ByzantiumBlock:      big.NewInt(1),
		EIP150Block:         big.NewInt(2),
		EIP150Hash:          common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(3),
		PetersburgBlock:     big.NewInt(4),
		ConstantinopleBlock: big.NewInt(5),

		SixtySixPercentBlock: big.NewInt(2000000),

		Sport: &SportConfig{
			Epoch:         30000,
			SpeakerPolicy: 0,
			MinFunds:      20000,
		},

		IsSmilo:       true,
		IsGas:         true,
		IsGasRefunded: true,

		ChainID: big.NewInt(20080914),

		RequiredMinFunds: 1,
	}

	//// SportTrustedCheckpoint contains the light client trusted checkpoint for the Smilo test network.
	//SportTrustedCheckpoint = &TrustedCheckpoint{
	//	Name:         "smilo",
	//	SectionIndex: 0,
	//	SectionHead:  common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
	//	CHTRoot:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
	//	BloomRoot:    common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
	//}

	// AllEthashProtocolChanges contains every protocol change (EIPs) introduced
	// and accepted by the Ethereum core developers into the Ethash consensus.
	//
	// This configuration is intentionally not using keyed fields to force anyone
	// adding flags to the config to also have to set these fields.
	AllEthashProtocolChanges = &ChainConfig{big.NewInt(20080914), big.NewInt(0), nil, false, big.NewInt(0), common.Hash{}, big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), nil, nil, new(EthashConfig), nil, nil, false, true, false, 0, 32, nil, nil, nil, nil, nil, big.NewInt(0)}

	// AllCliqueProtocolChanges contains every protocol change (EIPs) introduced
	// and accepted by the Ethereum core developers into the Clique consensus.
	//
	// This configuration is intentionally not using keyed fields to force anyone
	// adding flags to the config to also have to set these fields.
	AllCliqueProtocolChanges = &ChainConfig{big.NewInt(1337), big.NewInt(0), nil, false, big.NewInt(0), common.Hash{}, big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), nil, nil, nil, &CliqueConfig{Period: 0, Epoch: 30000}, nil, false, false, false, 0, 32, nil, nil, nil, nil, nil, big.NewInt(0)}

	TestChainConfig = &ChainConfig{big.NewInt(10), big.NewInt(0), nil, false, big.NewInt(0), common.Hash{}, big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), nil, nil, new(EthashConfig), nil, nil, false, true, false, 0, 32, nil, nil, nil, nil, nil, big.NewInt(0)}
	TestRules       = TestChainConfig.Rules(new(big.Int))

	SmiloTestChainConfig = &ChainConfig{big.NewInt(10), big.NewInt(0), nil, false, nil, common.Hash{}, nil, nil, big.NewInt(300000), nil, nil, big.NewInt(0), nil, nil, new(EthashConfig), nil, nil, true, true, false, 0, 32, nil, nil, nil, nil, nil, big.NewInt(0)}
)

// TrustedCheckpoint represents a set of post-processed trie roots (CHT and
// BloomTrie) associated with the appropriate section index and head hash. It is
// used to start light syncing from this checkpoint and avoid downloading the
// entire header chain while still being able to securely access old headers/logs.
type TrustedCheckpoint struct {
	SectionIndex uint64      `json:"sectionIndex"`
	SectionHead  common.Hash `json:"sectionHead"`
	CHTRoot      common.Hash `json:"chtRoot"`
	BloomRoot    common.Hash `json:"bloomRoot"`
}

// HashEqual returns an indicator comparing the itself hash with given one.
func (c *TrustedCheckpoint) HashEqual(hash common.Hash) bool {
	if c.Empty() {
		return hash == common.Hash{}
	}
	return c.Hash() == hash
}

// Hash returns the hash of checkpoint's four key fields(index, sectionHead, chtRoot and bloomTrieRoot).
func (c *TrustedCheckpoint) Hash() common.Hash {
	buf := make([]byte, 8+3*common.HashLength)
	binary.BigEndian.PutUint64(buf, c.SectionIndex)
	copy(buf[8:], c.SectionHead.Bytes())
	copy(buf[8+common.HashLength:], c.CHTRoot.Bytes())
	copy(buf[8+2*common.HashLength:], c.BloomRoot.Bytes())
	return crypto.Keccak256Hash(buf)
}

// Empty returns an indicator whether the checkpoint is regarded as empty.
func (c *TrustedCheckpoint) Empty() bool {
	return c.SectionHead == (common.Hash{}) || c.CHTRoot == (common.Hash{}) || c.BloomRoot == (common.Hash{})
}

// CheckpointOracleConfig represents a set of checkpoint contract(which acts as an oracle)
// config which used for light client checkpoint syncing.
type CheckpointOracleConfig struct {
	Address   common.Address   `json:"address"`
	Signers   []common.Address `json:"signers"`
	Threshold uint64           `json:"threshold"`
}

// ChainConfig is the core config which determines the blockchain settings.
//
// ChainConfig is stored in the database on a per block basis. This means
// that any network, identified by its genesis block, can have its own
// set of configuration options.
type ChainConfig struct {
	ChainID *big.Int `json:"chainId"` // chainId identifies the current chain and is used for replay protection

	HomesteadBlock *big.Int `json:"homesteadBlock,omitempty"` // Homestead switch block (nil = no fork, 0 = already homestead)

	DAOForkBlock   *big.Int `json:"daoForkBlock,omitempty"`   // TheDAO hard-fork switch block (nil = no fork)
	DAOForkSupport bool     `json:"daoForkSupport,omitempty"` // Whether the nodes supports or opposes the DAO hard-fork

	// EIP150 implements the Gas price changes (https://github.com/ethereum/EIPs/issues/150)
	EIP150Block *big.Int    `json:"eip150Block,omitempty"` // EIP150 HF block (nil = no fork)
	EIP150Hash  common.Hash `json:"eip150Hash,omitempty"`  // EIP150 HF hash (needed for header only clients as only gas pricing changed)

	EIP155Block *big.Int `json:"eip155Block,omitempty"` // EIP155 HF block
	EIP158Block *big.Int `json:"eip158Block,omitempty"` // EIP158 HF block

	SixtySixPercentBlock *big.Int `json:"sixtySixPercentBlock"` // The 66% switch block (nil = no fork)

	ByzantiumBlock      *big.Int `json:"byzantiumBlock,omitempty"`      // Byzantium switch block (nil = no fork, 0 = already on byzantium)
	ConstantinopleBlock *big.Int `json:"constantinopleBlock,omitempty"` // Constantinople switch block (nil = no fork, 0 = already activated)
	PetersburgBlock     *big.Int `json:"petersburgBlock,omitempty"`     // Petersburg switch block (nil = same as Constantinople)
	IstanbulBlock       *big.Int `json:"istanbulBlock,omitempty"`       // Istanbul switch block (nil = no fork, 0 = already on istanbul)
	EWASMBlock          *big.Int `json:"ewasmBlock,omitempty"`          // EWASM switch block (nil = no fork, 0 = already activated)

	// Various consensus engines
	Ethash *EthashConfig `json:"ethash,omitempty"`
	Clique *CliqueConfig `json:"clique,omitempty"`
	Sport  *SportConfig  `json:"sport,omitempty"`

	IsSmilo bool `json:"isSmilo"`

	IsGas         bool `json:"isGas"`         //true for when using gas, false when not using any
	IsGasRefunded bool `json:"isGasRefunded"` //true for when using gas and refund is enabled, false when not refundable

	RequiredMinFunds           int64  `json:"required_min_funds"` // 1e16 -> 1 -> 1e16
	CustomTransactionSizeLimit uint64 `json:"custom_transaction_size_limit"`

	Tendermint             *TendermintConfig        `json:"tendermint,omitempty"`
	AutonityContractConfig *AutonityContractGenesis `json:"autonityContract,omitempty"`
	Istanbul               *IstanbulConfig          `json:"istanbul,omitempty"`
	SportDAO               *SportDAOConfig          `json:"sportdao,omitempty"`

	// Quorum
	//
	// QIP714Block implements the permissions related changes
	QIP714Block            *big.Int `json:"qip714Block,omitempty"`
	MaxCodeSizeChangeBlock *big.Int `json:"maxCodeSizeChangeBlock,omitempty"`
}

// EthashConfig is the consensus engine configs for proof-of-work based sealing.
type EthashConfig struct{}

// String implements the stringer interface, returning the consensus engine details.
func (c *EthashConfig) String() string {
	return "ethash"
}

// CliqueConfig is the consensus engine configs for proof-of-authority based sealing.
type CliqueConfig struct {
	Period uint64 `json:"period"` // Number of seconds between blocks to enforce
	Epoch  uint64 `json:"epoch"`  // Epoch length to reset votes and checkpoint
}

// String implements the stringer interface, returning the consensus engine details.
func (c *CliqueConfig) String() string {
	return "clique"
}

// SportConfig is the consensus engine configs for Sport based sealing.
type SportConfig struct {
	Epoch         uint64 `json:"epoch"`    // Epoch length to reset votes and checkpoint
	SpeakerPolicy uint64 `json:"policy"`   // The policy for speaker selection
	MinFunds      int64  `json:"minfunds"` // The policy for speaker selection
}

// String implements the stringer interface, returning the consensus engine details.
func (c *SportConfig) String() string {
	return "smilobft"
}

// IstanbulConfig is the consensus engine configs for Istanbul based sealing.
type SportDAOConfig struct {
	Epoch         uint64 `json:"epoch"`    // Epoch length to reset votes and checkpoint
	SpeakerPolicy uint64 `json:"policy"`   // The policy for speaker selection
	MinFunds      int64  `json:"minfunds"` // The policy for speaker selection
}

//String implements the stringer interface, returning the consensus engine details.
func (c *SportDAOConfig) String() string {
	return "smilobftdao"
}

// IstanbulConfig is the consensus engine configs for Istanbul based sealing.
type IstanbulConfig struct {
	Epoch          uint64   `json:"epoch"`  // Epoch length to reset votes and checkpoint
	ProposerPolicy uint64   `json:"policy"` // The policy for proposer selection
	BlockPeriod    uint64   `json:"block-period"`
	RequestTimeout uint64   `json:"request-timeout"`
	Ceil2Nby3Block *big.Int `json:"ceil2Nby3Block,omitempty"` // Number of confirmations required to move from one state to next [2F + 1 to Ceil(2N/3)]
}

// String implements the stringer interface, returning the consensus engine details.
func (c *IstanbulConfig) String() string {
	return "istanbul"
}

// TendermintConfig is the consensus engine configs for Tendermint based sealing.
type TendermintConfig struct {
	Epoch          uint64 `json:"epoch"`  // Epoch length to reset votes and checkpoint
	ProposerPolicy uint64 `json:"policy"` // The policy for proposer selection
	BlockPeriod    uint64 `json:"block-period"`
	RequestTimeout uint64 `json:"request-timeout"`
}

// String implements the stringer interface, returning the consensus engine details.
func (c *TendermintConfig) String() string {
	return "tendermint"
}

// String implements the fmt.Stringer interface.
func (c *ChainConfig) String() string {
	var engine interface{}
	switch {
	case c.Sport != nil:
		engine = c.Sport
	case c.Ethash != nil:
		engine = c.Ethash
	case c.Clique != nil:
		engine = c.Clique
	case c.Istanbul != nil:
		engine = c.Istanbul
	case c.SportDAO != nil:
		engine = c.SportDAO
	case c.Tendermint != nil:
		engine = c.Tendermint
	default:
		engine = "unknown"
	}
	return fmt.Sprintf("{ChainID: %v Homestead: %v DAO: %v DAOSupport: %v EIP150: %v EIP155: %v EIP158: %v Byzantium: %v Constantinople: %v Petersburg: %v Istanbul: %v IsSmilo: %v, IsGas: %v, IsGasRefunded: %v, MinFunds: %v, Engine: %v}",
		c.ChainID,
		c.HomesteadBlock,
		c.DAOForkBlock,
		c.DAOForkSupport,
		c.EIP150Block,
		c.EIP155Block,
		c.EIP158Block,
		c.ByzantiumBlock,
		c.ConstantinopleBlock,
		c.PetersburgBlock,
		c.IstanbulBlock,
		c.IsSmilo,
		c.IsGas,
		c.IsGasRefunded,
		c.RequiredMinFunds,
		engine,
	)
}

// IsHomestead returns whether num is either equal to the homestead block or greater.
func (c *ChainConfig) IsHomestead(num *big.Int) bool {
	return isForked(c.HomesteadBlock, num)
}

// IsDAOFork returns whether num is either equal to the DAO fork block or greater.
func (c *ChainConfig) IsDAOFork(num *big.Int) bool {
	return isForked(c.DAOForkBlock, num)
}

// IsEIP150 returns whether num is either equal to the EIP150 fork block or greater.
func (c *ChainConfig) IsEIP150(num *big.Int) bool {
	return isForked(c.EIP150Block, num)
}

// IsEIP155 returns whether num is either equal to the EIP155 fork block or greater.
func (c *ChainConfig) IsEIP155(num *big.Int) bool {
	return isForked(c.EIP155Block, num)
}

// IsEIP158 returns whether num is either equal to the EIP158 fork block or greater.
func (c *ChainConfig) IsEIP158(num *big.Int) bool {
	return isForked(c.EIP158Block, num)
}

// IsByzantium returns whether num is either equal to the Byzantium fork block or greater.
func (c *ChainConfig) IsByzantium(num *big.Int) bool {
	return isForked(c.ByzantiumBlock, num)
}

// IsConstantinople returns whether num is either equal to the Constantinople fork block or greater.
func (c *ChainConfig) IsConstantinople(num *big.Int) bool {
	return isForked(c.ConstantinopleBlock, num)
}

// IsPetersburg returns whether num is either
// - equal to or greater than the PetersburgBlock fork block,
// - OR is nil, and Constantinople is active
func (c *ChainConfig) IsPetersburg(num *big.Int) bool {
	return isForked(c.PetersburgBlock, num) || c.PetersburgBlock == nil && isForked(c.ConstantinopleBlock, num)
}

// IsIstanbul returns whether num is either equal to the Istanbul fork block or greater.
func (c *ChainConfig) IsIstanbul(num *big.Int) bool {
	return isForked(c.IstanbulBlock, num)
}

// IsEWASM returns whether num represents a block number after the EWASM fork
func (c *ChainConfig) IsEWASM(num *big.Int) bool {
	return isForked(c.EWASMBlock, num)
}

// Quorum
//
// IsQIP714 returns whether num represents a block number where permissions is enabled
func (c *ChainConfig) IsQIP714(num *big.Int) bool {
	return isForked(c.QIP714Block, num)
}

// Quorum
//
// IsMaxCodeSizeChangeBlock returns whether num represents a block number max code size
// was changed from default 24K to new value
func (c *ChainConfig) IsMaxCodeSizeChangeBlock(num *big.Int) bool {
	return isForked(c.MaxCodeSizeChangeBlock, num)
}

// GasTable returns the gas table corresponding to the current phase (homestead or homestead reprice).
//
// The returned GasTable's fields shouldn't, under any circumstances, be changed.
func (c *ChainConfig) GasTable(num *big.Int) GasTable {
	if num == nil {
		return GasTableHomestead
	}
	switch {
	case c.IsConstantinople(num):
		return GasTableConstantinople
	case c.IsEIP158(num):
		return GasTableEIP158
	case c.IsEIP150(num):
		return GasTableEIP150
	default:
		return GasTableHomestead
	}
}

// CheckCompatible checks whether scheduled fork transitions have been imported
// with a mismatching chain configuration.
func (c *ChainConfig) CheckCompatible(newcfg *ChainConfig, height uint64, isSmiloEIP155Activated bool) *ConfigCompatError {
	bhead := new(big.Int).SetUint64(height)

	// Iterate checkCompatible to find the lowest conflict.
	var lasterr *ConfigCompatError
	for {
		err := c.checkCompatible(newcfg, bhead, isSmiloEIP155Activated)
		if err == nil || (lasterr != nil && err.RewindTo == lasterr.RewindTo) {
			break
		}
		lasterr = err
		bhead.SetUint64(err.RewindTo)
	}
	return lasterr
}

func (c *ChainConfig) checkCompatible(newcfg *ChainConfig, head *big.Int, isSmiloEIP155Activated bool) *ConfigCompatError {
	if isForkIncompatible(c.HomesteadBlock, newcfg.HomesteadBlock, head) {
		return newCompatError("Homestead fork block", c.HomesteadBlock, newcfg.HomesteadBlock)
	}
	if isForkIncompatible(c.DAOForkBlock, newcfg.DAOForkBlock, head) {
		return newCompatError("DAO fork block", c.DAOForkBlock, newcfg.DAOForkBlock)
	}
	if c.IsDAOFork(head) && c.DAOForkSupport != newcfg.DAOForkSupport {
		return newCompatError("DAO fork support flag", c.DAOForkBlock, newcfg.DAOForkBlock)
	}
	if isForkIncompatible(c.EIP150Block, newcfg.EIP150Block, head) {
		return newCompatError("EIP150 fork block", c.EIP150Block, newcfg.EIP150Block)
	}
	if isSmiloEIP155Activated && c.ChainID != nil && isForkIncompatible(c.EIP155Block, newcfg.EIP155Block, head) {
		return newCompatError("EIP155 fork block", c.EIP155Block, newcfg.EIP155Block)
	}
	if isSmiloEIP155Activated && c.ChainID != nil && c.IsEIP155(head) && !configNumEqual(c.ChainID, newcfg.ChainID) {
		return newCompatError("EIP155 chain ID", c.ChainID, newcfg.ChainID)
	}
	if isForkIncompatible(c.EIP158Block, newcfg.EIP158Block, head) {
		return newCompatError("EIP158 fork block", c.EIP158Block, newcfg.EIP158Block)
	}
	if c.IsEIP158(head) && !configNumEqual(c.ChainID, newcfg.ChainID) {
		return newCompatError("EIP158 chain ID", c.EIP158Block, newcfg.EIP158Block)
	}
	if isForkIncompatible(c.ByzantiumBlock, newcfg.ByzantiumBlock, head) {
		return newCompatError("Byzantium fork block", c.ByzantiumBlock, newcfg.ByzantiumBlock)
	}
	if isForkIncompatible(c.ConstantinopleBlock, newcfg.ConstantinopleBlock, head) {
		return newCompatError("Constantinople fork block", c.ConstantinopleBlock, newcfg.ConstantinopleBlock)
	}
	if isForkIncompatible(c.PetersburgBlock, newcfg.PetersburgBlock, head) {
		return newCompatError("Petersburg fork block", c.PetersburgBlock, newcfg.PetersburgBlock)
	}
	if isForkIncompatible(c.IstanbulBlock, newcfg.IstanbulBlock, head) {
		return newCompatError("Istanbul fork block", c.IstanbulBlock, newcfg.IstanbulBlock)
	}
	if isForkIncompatible(c.EWASMBlock, newcfg.EWASMBlock, head) {
		return newCompatError("ewasm fork block", c.EWASMBlock, newcfg.EWASMBlock)
	}
	if c.Istanbul != nil && newcfg.Istanbul != nil && isForkIncompatible(c.Istanbul.Ceil2Nby3Block, newcfg.Istanbul.Ceil2Nby3Block, head) {
		return newCompatError("Ceil 2N/3 fork block", c.Istanbul.Ceil2Nby3Block, newcfg.Istanbul.Ceil2Nby3Block)
	}
	if isForkIncompatible(c.QIP714Block, newcfg.QIP714Block, head) {
		return newCompatError("permissions fork block", c.QIP714Block, newcfg.QIP714Block)
	}
	if isForkIncompatible(c.MaxCodeSizeChangeBlock, newcfg.MaxCodeSizeChangeBlock, head) {
		return newCompatError("max code size change fork block", c.MaxCodeSizeChangeBlock, newcfg.MaxCodeSizeChangeBlock)
	}
	return nil
}

//func (c *ChainConfig) GetEnodeWhitelist() []string {
//	c.mu.RLock()
//	defer c.mu.RUnlock()
//
//	return c.EnodeWhitelist
//}
//func (c *ChainConfig) SetEnodeWhitelist(l []string) {
//	c.mu.Lock()
//	defer c.mu.Unlock()
//
//	c.EnodeWhitelist = l
//}
//func (c *ChainConfig) SortEnodeWhitelist() {
//	c.mu.Lock()
//	defer c.mu.Unlock()
//
//	sort.Strings(c.EnodeWhitelist)
//}
//
//func (c *ChainConfig) GetGlienickeDeployer() common.Address {
//	c.mu.RLock()
//	defer c.mu.RUnlock()
//
//	return c.GlienickeDeployer
//}
//func (c *ChainConfig) SetGlienickeDeployer(a common.Address) {
//	c.mu.Lock()
//	defer c.mu.Unlock()
//
//	c.GlienickeDeployer = a
//}
//
//func (c *ChainConfig) GetGlienickeBytecode() string {
//	c.mu.RLock()
//	defer c.mu.RUnlock()
//
//	return c.GlienickeBytecode
//}
//func (c *ChainConfig) SetGlienickeBytecode(s string) {
//	c.mu.Lock()
//	defer c.mu.Unlock()
//
//	c.GlienickeBytecode = s
//}
//
//func (c *ChainConfig) GetGlienickeABI() string {
//	c.mu.RLock()
//	defer c.mu.RUnlock()
//
//	return c.GlienickeABI
//}
//func (c *ChainConfig) SetGlienickeABI(s string) {
//	c.GlienickeABI = s
//}

func (c *ChainConfig) Copy() *ChainConfig {
	cfg := &ChainConfig{
		DAOForkSupport: c.DAOForkSupport,
		EIP150Hash:     c.EIP150Hash,
	}
	if c.Ethash != nil {
		cfg.Ethash = &(*c.Ethash)
	}
	if c.Clique != nil {
		cfg.Clique = &(*c.Clique)
	}
	if c.Istanbul != nil {
		cfg.Istanbul = &(*c.Istanbul)
	}
	if c.Tendermint != nil {
		cfg.Tendermint = &(*c.Tendermint)
	}
	if c.ChainID != nil {
		cfg.ChainID = big.NewInt(0).Set(c.ChainID)
	}
	if c.HomesteadBlock != nil {
		cfg.HomesteadBlock = big.NewInt(0).Set(c.HomesteadBlock)
	}
	if c.DAOForkBlock != nil {
		cfg.DAOForkBlock = big.NewInt(0).Set(c.DAOForkBlock)
	}
	if c.EIP150Block != nil {
		cfg.EIP150Block = big.NewInt(0).Set(c.EIP150Block)
	}
	if c.EIP155Block != nil {
		cfg.EIP155Block = big.NewInt(0).Set(c.EIP155Block)
	}
	if c.EIP158Block != nil {
		cfg.EIP158Block = big.NewInt(0).Set(c.EIP158Block)
	}
	if c.ByzantiumBlock != nil {
		cfg.ByzantiumBlock = big.NewInt(0).Set(c.ByzantiumBlock)
	}
	if c.ConstantinopleBlock != nil {
		cfg.ConstantinopleBlock = big.NewInt(0).Set(c.ConstantinopleBlock)
	}
	if c.PetersburgBlock != nil {
		cfg.PetersburgBlock = big.NewInt(0).Set(c.PetersburgBlock)
	}
	if c.EWASMBlock != nil {
		cfg.EWASMBlock = big.NewInt(0).Set(c.EWASMBlock)
	}

	return cfg
}

// isForkIncompatible returns true if a fork scheduled at s1 cannot be rescheduled to
// block s2 because head is already past the fork.
func isForkIncompatible(s1, s2, head *big.Int) bool {
	return (isForked(s1, head) || isForked(s2, head)) && !configNumEqual(s1, s2)
}

// isForked returns whether a fork scheduled at block s is active at the given head block.
func isForked(s, head *big.Int) bool {
	if s == nil || head == nil {
		return false
	}
	return s.Cmp(head) <= 0
}

func configNumEqual(x, y *big.Int) bool {
	if x == nil {
		return y == nil
	}
	if y == nil {
		return x == nil
	}
	return x.Cmp(y) == 0
}

// ConfigCompatError is raised if the locally-stored blockchain is initialised with a
// ChainConfig that would alter the past.
type ConfigCompatError struct {
	What string
	// block numbers of the stored and new configurations
	StoredConfig, NewConfig *big.Int
	// the block number to which the local chain must be rewound to correct the error
	RewindTo uint64
}

func newCompatError(what string, storedblock, newblock *big.Int) *ConfigCompatError {
	var rew *big.Int
	switch {
	case storedblock == nil:
		rew = newblock
	case newblock == nil || storedblock.Cmp(newblock) < 0:
		rew = storedblock
	default:
		rew = newblock
	}
	err := &ConfigCompatError{what, storedblock, newblock, 0}
	if rew != nil && rew.Sign() > 0 {
		err.RewindTo = rew.Uint64() - 1
	}
	return err
}

func (err *ConfigCompatError) Error() string {
	return fmt.Sprintf("mismatching %s in database (have %d, want %d, rewindto %d)", err.What, err.StoredConfig, err.NewConfig, err.RewindTo)
}

// Rules wraps ChainConfig and is merely syntactic sugar or can be used for functions
// that do not have or require information about the block.
//
// Rules is a one time interface meaning that it shouldn't be used in between transition
// phases.
type Rules struct {
	ChainID                                                 *big.Int
	IsHomestead, IsEIP150, IsEIP155, IsEIP158               bool
	IsByzantium, IsConstantinople, IsPetersburg, IsIstanbul bool
}

// Rules ensures c's ChainID is not nil.
func (c *ChainConfig) Rules(num *big.Int) Rules {
	chainID := c.ChainID
	if chainID == nil {
		chainID = new(big.Int)
	}
	return Rules{
		ChainID:          new(big.Int).Set(chainID),
		IsHomestead:      c.IsHomestead(num),
		IsEIP150:         c.IsEIP150(num),
		IsEIP155:         c.IsEIP155(num),
		IsEIP158:         c.IsEIP158(num),
		IsByzantium:      c.IsByzantium(num),
		IsConstantinople: c.IsConstantinople(num),
		IsPetersburg:     c.IsPetersburg(num),
		IsIstanbul:       c.IsIstanbul(num),
	}
}

// Rules for CustomTransactionSizeLimit if set
func (c *ChainConfig) IsValid() error {
	if c.CustomTransactionSizeLimit < 32 || c.CustomTransactionSizeLimit > 128 {
		return errors.New("custom transaction size limit must be bigger than 32 and lower than 128")
	}
	return nil
}
