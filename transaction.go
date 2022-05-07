package spartan_go

import (
	"crypto/rsa"
	"fmt"
)

type TxOuput struct {
	amount  uint
	address string
}

type Transaction struct {
	Fee     uint
	From    string
	Nonce   uint
	PubKey  rsa.PublicKey
	sig     string
	Outputs []TxOuput
}

const TX_CONST = "TX"

func NewTransaction(from string, nonce uint, pubKey rsa.PublicKey, sig string, fee uint, outputs []TxOuput) *Transaction {
	newTx := &Transaction{
		Fee:     fee,
		From:    from,
		Nonce:   nonce,
		PubKey:  pubKey,
		sig:     sig,
		Outputs: make([]TxOuput, 0),
	}
	for _, output := range outputs {
		newTx.Outputs = append(newTx.Outputs, output)
	}
	return newTx
}

func (t *Transaction) Id() string {
	txWithoutSig := &Transaction{
		Fee:     t.Fee,
		From:    t.From,
		Nonce:   t.Nonce,
		PubKey:  t.PubKey,
		Outputs: t.Outputs,
	}
	return Hash(TX_CONST+fmt.Sprintf("%+v", txWithoutSig), "")
}

func (t *Transaction) Sign(privKey *rsa.PrivateKey) {
	if sig, err := Sign(privKey, t.Id()); err != nil {
		t.sig = sig
	}
}

func (t *Transaction) ValidSignature() bool {
	return len(t.sig) > 0 &&
		AddressMatchesKey(t.From, t.PubKey) &&
		VerifySignature(t.PubKey, t.Id(), t.sig)
}

func (t *Transaction) SufficientFunds(block *Block) bool {
	return t.TotalOutput() <= block.Balances[t.From]
}

func (t *Transaction) TotalOutput() uint {
	totalOutput := t.Fee
	for _, output := range t.Outputs {
		totalOutput += output.amount
	}
	return totalOutput
}
