package spartan_go

import (
	"strconv"
	"sync"
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
	lock           sync.Mutex
}

func NewBlock(rewardAddr string, prevBlock *Block, target *uint256.Int, coinbaseReward ...uint) *Block {
	if prevBlock != nil {
		prevBlock.lock.Lock()
		defer prevBlock.lock.Unlock()
	}

	newBlock := &Block{}
	newBlock.RewardAddr = rewardAddr
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

	if target != nil {
		newBlock.Target = target
	} else {
		newBlock.Target = POW_TARGET
	}

	if coinbaseReward != nil {
		newBlock.CoinbaseReward = coinbaseReward[0]
	} else {
		newBlock.CoinbaseReward = COINBASE_AMT_ALLOWED
	}

	newBlock.Timestamp = time.Now()
	return newBlock
}

func (b *Block) IsGenesisBlock() bool {
	return b.ChainLength == 0
}

func (b *Block) HasValidProof() bool {
	h := b.HashVal()

	// remove leading zeroes because uint256.FromHex doesn't like those
	for h[0] == '0' {
		h = h[1:]
	}

	n, _ := uint256.FromHex("0x" + h)
	return n.Cmp(b.Target) < 0
}

func (b *Block) Serialize() string {
	// it took 2 whole days of debugging to find out reading the maps in this block (part of
	// Sprintf used below) while editing the block in another thread was causing concurrency issues
	// b.lock.Lock()
	// defer b.lock.Unlock()
	// return fmt.Sprintf("%+v", b)
	// return Jsonify(b)
	// return b.RewardAddr + "ffff"
	return b.RewardAddr + b.PrevBlockHash + strconv.FormatUint(uint64(b.Proof), 10)
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
	if _, ok := b.Transactions[tx.Id()]; ok {
		if client != nil {
			client.log("Duplicate transaction " + tx.Id())
		}
		return false
	} else if len(tx.sig) == 0 {
		if client != nil {
			client.log("Unsigned transaction " + tx.Id())
		}
		return false
	} else if !tx.ValidSignature() {
		if client != nil {
			client.log("Invalid signature for transaction " + tx.Id())
		}
		return false
	} else if !tx.SufficientFunds(b) {
		if client != nil {
			client.log("Insufficient gold for transaction " + tx.Id())
		}
		return false
	}

	nonce, ok := b.NextNonce[tx.From]
	if !ok {
		nonce = 0
	}
	if tx.Nonce < nonce {
		if client != nil {
			client.log("Replayed transaction " + tx.Id())
		}
		return false
	} else if tx.Nonce > nonce {
		if client != nil {
			client.log("Out of order transaction " + tx.Id())
		}
		return false
	}
	b.NextNonce[tx.From] = nonce + 1

	b.Transactions[tx.Id()] = tx
	senderBalance := b.BalanceOf(tx.From)
	b.Balances[tx.From] = senderBalance - tx.TotalOutput()

	for _, output := range tx.Outputs {
		oldBalance := b.BalanceOf(output.Address)
		b.Balances[output.Address] = output.Amount + oldBalance
	}

	return true
}

func (b *Block) rerun(prevBlock *Block) bool {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.Balances = make(map[string]uint)
	b.NextNonce = make(map[string]uint)
	for key, val := range prevBlock.Balances {
		b.Balances[key] = val
	}
	for key, val := range prevBlock.NextNonce {
		b.NextNonce[key] = val
	}

	winnerBalance := b.BalanceOf(prevBlock.RewardAddr)
	if len(prevBlock.RewardAddr) != 0 {
		b.Balances[prevBlock.RewardAddr] = winnerBalance + prevBlock.TotalRewards()
	}

	txs := b.Transactions
	b.Transactions = make(map[string]*Transaction)
	for _, tx := range txs {
		success := b.AddTransaction(tx, nil)
		if !success {
			return false
		}
	}
	return true
}

// toJSON() isn't used for anything so I omitted it
