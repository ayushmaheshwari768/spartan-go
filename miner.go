package spartan_go

import (
	"strconv"

	. "github.com/vansante/go-event-emitter"
)

type Miner struct {
	Emitter
	Client       *Client
	CurrentBlock *Block
	miningRounds uint
	transactions map[string]*Transaction
}

func NewMiner(cfg *Client, miningRounds ...uint) *Miner {
	var rounds uint
	if len(miningRounds) == 1 {
		rounds = miningRounds[0]
	} else {
		rounds = NUM_ROUNDS_MINING
	}
	miner := &Miner{
		Emitter:      *NewEmitter(true),
		Client:       NewClient(cfg),
		miningRounds: rounds,
		transactions: make(map[string]*Transaction),
	}
	miner.AddListener(START_MINING, miner.findProof)
	miner.AddListener(POST_TRANSACTION, miner.addTransaction)
	miner.AddListener(MISSING_BLOCK, miner.provideMissingBlock)
	miner.AddListener(PROOF_FOUND, miner.receiveBlock)
	return miner
}

func (m *Miner) Initialize() {
	m.startNewSearch()
	// go func() {
	m.EmitEvent(START_MINING)
	// }()
}

func (m *Miner) startNewSearch(transactions ...map[string]*Transaction) {
	var txMap map[string]*Transaction
	if len(transactions) != 1 {
		txMap = make(map[string]*Transaction)
	} else {
		txMap = transactions[0]
	}

	m.CurrentBlock = NewBlock(m.Client.Address, m.Client.LastBlock, nil)
	for id, tx := range txMap {
		m.transactions[id] = tx
	}

	for _, tx := range m.transactions {
		m.CurrentBlock.AddTransaction(tx, m.Client)
	}
	m.transactions = make(map[string]*Transaction)
	m.CurrentBlock.Proof = 0
}

func (m *Miner) findProof(oneAndDone ...interface{}) {
	var testing bool
	if oneAndDone != nil {
		testing = true
	}

	pausePoint := m.CurrentBlock.Proof + m.miningRounds
	for m.CurrentBlock.Proof < pausePoint {
		if m.CurrentBlock.HasValidProof() {
			m.Client.log("Found proof for block " + strconv.FormatUint(uint64(m.CurrentBlock.ChainLength), 10) + ": " + strconv.FormatUint(uint64(m.CurrentBlock.Proof), 10))
			m.announceProof()
			m.receiveBlock(m.CurrentBlock)
			break
		}
		m.CurrentBlock.Proof++
	}

	if !testing {
		// m.Client.log("mining again")
		// go func() {
		m.EmitEvent(START_MINING)
		// }()
	}
}

func (m *Miner) announceProof() {
	m.Client.Net.Broadcast(PROOF_FOUND, m.CurrentBlock)
}

func (m *Miner) receiveBlock(block ...interface{}) {
	b := block[0].(*Block)

	b = m.Client.receiveBlockHelper(b)
	if b == nil {
		return
	}

	if m.CurrentBlock != nil && b.ChainLength >= m.CurrentBlock.ChainLength {
		m.Client.log("Cutting over to new chain")
		txMap := m.syncTransactions(b)
		m.startNewSearch(txMap)
	}
}

func (m *Miner) syncTransactions(nb *Block) map[string]*Transaction {
	m.Client.blocksLock.Lock()
	defer m.Client.blocksLock.Unlock()

	cb := m.CurrentBlock
	cbTxs := make(map[string]*Transaction)
	nbTxs := make(map[string]*Transaction)

	for nb.ChainLength > cb.ChainLength {
		for id, tx := range nb.Transactions {
			nbTxs[id] = tx
		}
		nb = m.Client.blocks[nb.PrevBlockHash]
	}

	for cb != nil && cb.HashVal() != nb.HashVal() {
		for id, tx := range cb.Transactions {
			cbTxs[id] = tx
		}
		for id, tx := range nb.Transactions {
			nbTxs[id] = tx
		}
		cb = m.Client.blocks[cb.PrevBlockHash]
		nb = m.Client.blocks[nb.PrevBlockHash]
	}

	for id := range nbTxs {
		delete(cbTxs, id)
	}

	return cbTxs
}

func (m *Miner) addTransaction(txs ...interface{}) {
	if len(txs) != 1 {
		m.Client.log("addTransaction(...) requires 1 transaction parameter")
		return
	}
	tx := txs[0].(*Transaction)
	newTx := NewTransaction(tx.From, tx.Nonce, tx.PubKey, tx.sig, tx.Fee, tx.Outputs)
	m.transactions[newTx.Id()] = newTx
}

func (m *Miner) provideMissingBlock(o ...interface{}) {
	m.Client.provideMissingBlock(o...)
}

func (m *Miner) PostTransaction(outputs []TxOuput, fee ...uint) {
	tx, err := m.Client.PostTransaction(outputs, fee...)
	if err != nil {
		m.Client.log(err.Error())
		return
	}
	m.addTransaction(tx)
}
