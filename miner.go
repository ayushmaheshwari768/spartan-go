package spartan_go

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

}

func (m *Miner) addTransaction(tx ...interface{}) {

}
