package spartan_go

import (
	"strconv"
)

type Miner struct {
	Client
	currentBlock *Block
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
		Client:       *NewClient(cfg),
		miningRounds: rounds,
		transactions: make(map[string]*Transaction),
	}
	return miner
}

func (m *Miner) Initialize() {
	m.startNewSearch()
	m.AddListener(START_MINING, m.findProof)
	m.AddListener(POST_TRANSACTION, m.addTransaction)
	go func() {
		m.EmitEvent(START_MINING)
	}()
}

func (m *Miner) startNewSearch(transactions ...map[string]*Transaction) {
	var txMap map[string]*Transaction
	if len(transactions) != 1 {
		txMap = make(map[string]*Transaction)
	} else {
		txMap = transactions[0]
	}

	m.currentBlock = &Block{RewardAddr: m.Address, PrevBlock: m.lastBlock}
	for id, tx := range txMap {
		m.transactions[id] = tx
	}

	for _, tx := range m.transactions {
		m.currentBlock.AddTransaction(tx, &m.Client)
	}
	m.transactions = make(map[string]*Transaction)
	m.currentBlock.Proof = 0
}

func (m *Miner) findProof(oneAndDone ...interface{}) {
	var testing bool
	if oneAndDone != nil {
		testing = true
	}

	pausePoint := m.currentBlock.Proof + m.miningRounds
	for m.currentBlock.Proof < pausePoint {
		if m.currentBlock.HasValidProof() {
			m.log("Found proof for block " + strconv.FormatUint(uint64(m.currentBlock.ChainLength), 10) + ": " + strconv.FormatUint(uint64(m.currentBlock.Proof), 10))
			m.announceProof()
			m.receiveBlock(m.currentBlock)
			break
		}
		m.currentBlock.Proof++
	}

	if !testing {
		go func() {
			m.EmitEvent(START_MINING)
		}()
	}
}

func (m *Miner) announceProof() {
	m.Net.broadcast(PROOF_FOUND, m.currentBlock)
}

func (m *Miner) receiveBlock(block *Block) {
	m.Client.receiveBlock(block)
	if block == nil {
		return
	}

	if m.currentBlock != nil && block.ChainLength >= m.currentBlock.ChainLength {
		m.log("Cutting over to new chain")
		txMap := m.syncTransactions(block)
		m.startNewSearch(txMap)
	}
}

func (m *Miner) syncTransactions(nb *Block) map[string]*Transaction {
	cb := m.currentBlock
	cbTxs := make(map[string]*Transaction)
	nbTxs := make(map[string]*Transaction)

	for nb.ChainLength > cb.ChainLength {
		for id, tx := range nb.Transactions {
			nbTxs[id] = tx
		}
		nb = m.blocks[nb.PrevBlockHash]
	}

	for cb != nil && cb.HashVal() != nb.HashVal() {
		for id, tx := range cb.Transactions {
			cbTxs[id] = tx
		}
		for id, tx := range nb.Transactions {
			nbTxs[id] = tx
		}
		cb = m.blocks[cb.PrevBlockHash]
		nb = m.blocks[nb.PrevBlockHash]
	}

	for id := range nbTxs {
		delete(cbTxs, id)
	}

	return cbTxs
}

func (m *Miner) addTransaction(txs ...interface{}) {
	if len(txs) != 1 {
		m.log("addTransaction(...) requires 1 transaction parameter")
		return
	}
	tx := txs[0].(*Transaction)
	tx = NewTransaction(tx.From, tx.Nonce, tx.PubKey, tx.sig, tx.Fee, tx.Outputs)
	m.transactions[tx.Id()] = tx
}

func (m *Miner) PostTransaction(outputs []TxOuput, fee ...uint) {
	tx, err := m.Client.PostTransaction(outputs, fee...)
	if err != nil {
		m.log(err.Error())
		return
	}
	m.addTransaction(tx)
}
