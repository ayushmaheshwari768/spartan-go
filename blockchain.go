package spartan_go

import (
	"errors"

	"github.com/holiman/uint256"
)

// Go doesn't allow to pass types to structs/functions so I am leaving
// out blockClass/transactionClass and assuming that it is Block/Transaction
// from this package
type Blockchain struct {
	ClientBalanceMap map[*Client]uint
	StartingBalances map[string]uint
	powTarget        *uint256.Int
	powLeadingZeroes uint
	coinbaseAmount   uint
	defaultTxFee     uint
	confirmedDepth   uint
}

const (
	MISSING_BLOCK    = "MISSING_BLOCK"
	POST_TRANSACTION = "POST_TRANSACTION"
	PROOF_FOUND      = "PROOF_FOUND"
	START_MINING     = "START_MINING"

	NUM_ROUNDS_MINING = uint(2000)

	POW_LEADING_ZEROES = uint(20)

	COINBASE_AMT_ALLOWED = uint(25)
	DEFAULT_TX_FEE       = uint(1)

	CONFIRMED_DEPTH = uint(6)
)

var blockchain = &Blockchain{}
var POW_TARGET, _ = uint256.FromHex("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")

func MakeGenesis(cfg *Blockchain) (*Block, error) {
	if cfg.ClientBalanceMap == nil && cfg.StartingBalances == nil {
		return &Block{}, errors.New("Must initialize clientBalanceMap XOR startingBalances")
	}
	if cfg.ClientBalanceMap != nil && cfg.StartingBalances != nil {
		return &Block{}, errors.New("You may set clientBalanceMap XOR set startingBalances, but not both")
	}

	blockchain = cfg
	blockchain.powTarget = POW_TARGET.Rsh(POW_TARGET, POW_LEADING_ZEROES)

	var balances map[string]uint
	if cfg.ClientBalanceMap != nil {
		balances = make(map[string]uint)
		for client, balance := range cfg.ClientBalanceMap {
			balances[client.Address] = balance
		}
	} else {
		balances = cfg.StartingBalances
	}

	g := NewBlock("", nil, nil)
	g.Balances = make(map[string]uint)
	for addr, balance := range balances {
		g.Balances[addr] = balance
	}

	if cfg.ClientBalanceMap != nil {
		for client := range cfg.ClientBalanceMap {
			client.setGenesisBlock(g)
		}
	}

	return g, nil
}

// DeserializeBlock, MakeBlock and MakeTransaction are not implemented because
// taking object-like interface{} as parameter is discouraged in GoLang. All
// objects will take defined structs instead of the all encompassing interface{}

// func DeserializeBlock(o interface{}) *Block {
// 	if b, ok := o.(Block); ok {
// 		return &b
// 	}
// }

// func MakeBlock(cfg *BlockConfig) *Block {
// 	return &Block{cfg: cfg}
// }
