package spartan_go

import (
	"fmt"
	"time"

	"github.com/holiman/uint256"
)

type Block struct {
	RewardAddr     string
	Proof          uint
	PrevBlock      *Block
	PrevBlockHash  string
	Target         *uint256.Int
	CoinbaseReward uint
	Balances       map[string]uint
	NextNonce      map[string]uint
	Transactions   map[string]*Transaction
	ChainLength    uint
	Timestamp      time.Time
}

func NewBlock(rewardAddr string, prevBlock *Block, target *uint256.Int, coinbaseReward uint) *Block {
	newBlock := &Block{}
	newBlock.Target = target
	newBlock.RewardAddr = rewardAddr
	newBlock.CoinbaseReward = coinbaseReward
	newBlock.Balances = make(map[string]uint)
	newBlock.NextNonce = make(map[string]uint)
	newBlock.Transactions = make(map[string]*Transaction)
	if prevBlock != nil {
		newBlock.PrevBlockHash = prevBlock.HashVal()
		newBlock.ChainLength = prevBlock.ChainLength + 1
		for k, v := range prevBlock.Balances {
			newBlock.Balances[k] = v
		}
		for k, v := range prevBlock.NextNonce {
			newBlock.NextNonce[k] = v
		}
		if len(prevBlock.RewardAddr) != 0 {
			winnerBalance := newBlock.BalanceOf(prevBlock.RewardAddr)
			newBlock.Balances[prevBlock.RewardAddr] = winnerBalance + prevBlock.TotalRewards()
		}
	} else {
		newBlock.ChainLength = 0
	}
	newBlock.Timestamp = time.Now()
	return newBlock
}

func (b *Block) IsGenesisBlock() bool {
	return b.ChainLength == 0
}

func (b *Block) HasValidProof() bool {
	h := Hash(b.Serialize(), "")
	n, _ := uint256.FromHex(h)
	return n.Cmp(b.Target) < 0
}

func (b *Block) Serialize() string {
	return fmt.Sprintf("%+v", b)
}

func (b *Block) HashVal() string {
	return Hash(b.Serialize(), "")
}

func (b *Block) BalanceOf(addr string) uint {
	if balance, ok := b.Balances[addr]; ok {
		return balance
	} else {
		return 0
	}
}

func (b *Block) TotalRewards() uint {
	reward := b.CoinbaseReward
	for _, tx := range b.Transactions {
		reward += tx.Fee
	}
	return reward
}

func (b *Block) Contains(txId string) bool {
	_, ok := b.Transactions[txId]
	return ok
}

func (b *Block) AddTransaction(tx *Transaction, client *Client) bool {

}

func (b *Block) rerun(prevBlock *Block) bool {

}

// toJSON() isn't used for anything so I omitted it
