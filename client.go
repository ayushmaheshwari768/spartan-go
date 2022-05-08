package spartan_go

import (
	"crypto/rsa"
	"errors"
	"log"
	"strconv"
	"sync"

	. "github.com/vansante/go-event-emitter"
)

type Client struct {
	Emitter
	Name                        string
	key                         *rsa.PrivateKey
	Net                         *FakeNet
	nonce                       uint
	pendingOutgoingTransactions map[string]*Transaction
	pendingReceivedTransactions map[string]*Transaction
	blocks                      map[string]*Block
	pendingBlocks               map[string][]*Block
	StartingBlock               *Block
	LastBlock                   *Block
	LastConfirmedBlock          *Block
	Address                     string
	lock                        sync.Mutex
}

func NewClient(cfg *Client) *Client {
	client := &Client{
		Emitter:                     *NewEmitter(true),
		Name:                        cfg.Name,
		Net:                         cfg.Net,
		nonce:                       0,
		pendingOutgoingTransactions: make(map[string]*Transaction),
		pendingReceivedTransactions: make(map[string]*Transaction),
		blocks:                      make(map[string]*Block),
		pendingBlocks:               make(map[string][]*Block),
	}

	if cfg.key == nil {
		client.key = GenerateKey()
	} else {
		client.key = cfg.key
	}
	client.Address = CalcAddress(client.key.PublicKey)

	if cfg.StartingBlock != nil {
		client.setGenesisBlock(cfg.StartingBlock)
	}

	client.AddListener(PROOF_FOUND, client.receiveBlock)
	client.AddListener(MISSING_BLOCK, client.provideMissingBlock)

	return client
}

func (c *Client) setGenesisBlock(startingBlock *Block) error {
	if c.LastBlock != nil {
		return errors.New("Cannot set genesis block for existing blockchain")
	}
	c.LastConfirmedBlock = startingBlock
	c.LastBlock = startingBlock
	c.blocks[startingBlock.HashVal()] = startingBlock
	return nil
}

func (c *Client) ConfirmedBalance() uint {
	return c.LastConfirmedBlock.BalanceOf(c.Address)
}

func (c *Client) AvailableGold() uint {
	pendingSpent := uint(0)
	for _, tx := range c.pendingOutgoingTransactions {
		pendingSpent += tx.TotalOutput()
	}
	return c.ConfirmedBalance() - pendingSpent
}

func (c *Client) PostTransaction(outputs []TxOuput, fee ...uint) (*Transaction, error) {
	txFee := DEFAULT_TX_FEE
	if len(fee) == 1 {
		txFee = fee[0]
	}
	totalPayments := txFee
	for _, output := range outputs {
		totalPayments += output.Amount
	}
	if totalPayments > c.AvailableGold() {
		return nil, errors.New("Requested " + strconv.FormatUint(uint64(totalPayments), 10) + ", but account only has " + strconv.FormatUint(uint64(c.AvailableGold()), 10))
	}
	return c.postGenericTransaction(
		&Transaction{
			Outputs: outputs,
			Fee:     txFee,
			From:    c.Address,
			Nonce:   c.nonce,
			PubKey:  c.key.PublicKey,
		},
	), nil
}

func (c *Client) postGenericTransaction(tx *Transaction) *Transaction {
	tx.Sign(c.key)
	c.pendingOutgoingTransactions[tx.Id()] = tx
	c.nonce++
	c.Net.Broadcast(POST_TRANSACTION, tx)
	return tx
}

func (c *Client) receiveBlockHelper(b *Block) *Block {
	if b == nil {
		return nil
	}
	if _, ok := c.blocks[b.HashVal()]; ok {
		return nil
	}

	if !b.HasValidProof() && !b.IsGenesisBlock() {
		c.log("Block " + b.HashVal() + " does not have a valid proof.")
		return nil
	}

	prevBlock, ok := c.blocks[b.PrevBlockHash]
	if !ok && !b.IsGenesisBlock() {
		stuckBlocks, ok := c.pendingBlocks[b.PrevBlockHash]
		if !ok {
			c.requestMissingBlock(b)
			stuckBlocks = make([]*Block, 0)
		}
		alreadyPending := false
		for _, stuckBlock := range stuckBlocks {
			if stuckBlock.HashVal() == b.HashVal() {
				alreadyPending = true
				break
			}
		}
		if !alreadyPending {
			stuckBlocks = append(stuckBlocks, b)
		}
		c.pendingBlocks[b.PrevBlockHash] = stuckBlocks
		return nil
	}

	if !b.IsGenesisBlock() {
		success := b.rerun(prevBlock)
		if !success {
			return nil
		}
	}

	c.blocks[b.HashVal()] = b

	if c.LastBlock.ChainLength < b.ChainLength {
		c.LastBlock = b
		c.setLastConfirmed()
	}

	unstuckBlocks := make([]*Block, 0)
	if pBlocks, ok := c.pendingBlocks[b.HashVal()]; ok {
		unstuckBlocks = append(unstuckBlocks, pBlocks...)
	}
	delete(c.pendingBlocks, b.HashVal())

	for _, unstuckBlock := range unstuckBlocks {
		c.log("Processing unstuck block " + unstuckBlock.HashVal())
		c.receiveBlockHelper(unstuckBlock)
	}

	return b
}

func (c *Client) receiveBlock(block ...interface{}) {
	if len(block) != 1 {
		c.log("receiveBlock(...) is only supposed to receive 1 block")
		return
	}

	b := block[0].(*Block)
	c.receiveBlockHelper(b)
}

func (c *Client) requestMissingBlock(block *Block) {
	c.log("Asking for missing block " + block.HashVal())
	c.Net.Broadcast(MISSING_BLOCK, c.Address, block.PrevBlockHash)
}

// msg[0] = from address
// msg[1] = missing block hash
func (c *Client) provideMissingBlock(msg ...interface{}) {
	if len(msg) != 2 {
		return
	}
	if block, ok := c.blocks[msg[1].(string)]; ok {
		c.log("Providing missing block " + msg[1].(string))
		c.Net.SendMessage(msg[0].(string), PROOF_FOUND, block)
	}
}

func (c *Client) resendPendingTransactions() {
	for _, tx := range c.pendingOutgoingTransactions {
		c.Net.Broadcast(POST_TRANSACTION, tx)
	}
}

func (c *Client) setLastConfirmed() {
	block := c.LastBlock
	confirmedBlockHeight := block.ChainLength - CONFIRMED_DEPTH
	if confirmedBlockHeight < 0 {
		confirmedBlockHeight = 0
	}
	for block.ChainLength > confirmedBlockHeight {
		block = c.blocks[block.PrevBlockHash]
	}
	c.LastConfirmedBlock = block

	toDelete := make([]string, 0)
	for txId := range c.pendingOutgoingTransactions {
		if c.LastConfirmedBlock.Contains(txId) {
			toDelete = append(toDelete, txId)
		}
	}
	for _, txId := range toDelete {
		delete(c.pendingOutgoingTransactions, txId)
	}
}

func (c *Client) log(msg string) {
	name := c.Address[:10]
	if len(c.Name) != 0 {
		name = c.Name
	}
	log.Println(name + ": " + msg)
}

func (c *Client) ShowAllBalances() {
	c.log("Showing balances:")
	for id, balance := range c.LastConfirmedBlock.Balances {
		c.log("	" + id + ":" + strconv.FormatUint(uint64(balance), 10))
	}
}

func (c *Client) ShowBlockchain() {
	block := c.LastBlock
	log.Println("BLOCKCHAIN:")
	for block != nil {
		log.Println(block.HashVal())
		block = c.blocks[block.PrevBlockHash]
	}
}
