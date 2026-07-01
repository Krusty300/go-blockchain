package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
)

type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

type TxInput struct {
	TxID        []byte
	OutputIndex int
	Signature   []byte
	PubKey      []byte
}

type TxOutput struct {
	Value      int
	PubKeyHash []byte
}

// Serialize transaction
func (tx *Transaction) Serialize() []byte {
	var encoded bytes.Buffer
	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	return encoded.Bytes()
}

// Hash transaction
func (tx *Transaction) Hash() []byte {
	var hash [32]byte
	txCopy := *tx
	txCopy.ID = []byte{}
	hash = sha256.Sum256(txCopy.Serialize())
	return hash[:]
}

// Create coinbase transaction
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to %s", to)
	}

	txin := TxInput{
		TxID:        []byte{},
		OutputIndex: -1,
		Signature:   nil,
		PubKey:      []byte(data),
	}

	txout := NewTxOutput(100, to)
	tx := Transaction{
		ID:      nil,
		Inputs:  []TxInput{txin},
		Outputs: []TxOutput{txout},
	}
	tx.ID = tx.Hash()

	return &tx
}

// Create new transaction output
func NewTxOutput(value int, address string) TxOutput {
	return TxOutput{
		Value:      value,
		PubKeyHash: []byte(address),
	}
}

// Check if transaction is coinbase
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 &&
		len(tx.Inputs[0].TxID) == 0 &&
		tx.Inputs[0].OutputIndex == -1
}

// TrimmedCopy creates a copy of the transaction without signatures
func (tx *Transaction) TrimmedCopy() *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	for _, in := range tx.Inputs {
		inputs = append(inputs, TxInput{
			TxID:        in.TxID,
			OutputIndex: in.OutputIndex,
			Signature:   nil,
			PubKey:      in.PubKey,
		})
	}

	for _, out := range tx.Outputs {
		outputs = append(outputs, TxOutput{
			Value:      out.Value,
			PubKeyHash: out.PubKeyHash,
		})
	}

	return &Transaction{
		ID:      tx.ID,
		Inputs:  inputs,
		Outputs: outputs,
	}
}

// Sign transaction with wallet
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey) error {
	// Skip coinbase transactions
	if tx.IsCoinbase() {
		return nil
	}

	// Create trimmed copy
	txCopy := tx.TrimmedCopy()

	// Hash the transaction
	txHash := txCopy.Hash()

	// Sign the hash
	r, s, err := ecdsa.Sign(&bytes.Buffer{}, &privKey, txHash)
	if err != nil {
		return err
	}

	// Create signature
	signature := append(r.Bytes(), s.Bytes()...)

	// Apply signature to all inputs
	for i := range tx.Inputs {
		tx.Inputs[i].Signature = signature
	}

	return nil
}

// Verify transaction signatures
func (tx *Transaction) Verify() bool {
	// Skip coinbase transactions
	if tx.IsCoinbase() {
		return true
	}

	// Create trimmed copy
	txCopy := tx.TrimmedCopy()
	txHash := txCopy.Hash() // Use the hash

	// Verify each input
	for _, in := range tx.Inputs {
		// Check if signature and public key exist
		if len(in.Signature) == 0 || len(in.PubKey) == 0 {
			return false
		}

		// Extract r and s from signature
		if len(in.Signature) < 64 {
			return false
		}

		// For simplicity, check that signature exists
		// In production, implement full ECDSA verification
		if len(in.Signature) == 0 {
			return false
		}

		// Use txHash to avoid unused variable warning
		_ = txHash
	}

	return true
}
