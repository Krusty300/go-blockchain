package main

import (
	"bytes"
	"fmt"
	"log"
)

type Blockchain struct {
	Blocks  []*Block
	Storage *Storage
}

// CreateGenesisBlock creates the first block in the blockchain
func CreateGenesisBlock(coinbase *Transaction) *Block {
	return &Block{
		Timestamp:     0,
		Transactions:  []*Transaction{coinbase},
		PrevBlockHash: []byte{},
		Hash:          []byte{},
		Nonce:         0,
		Height:        0,
	}
}

func NewBlockchain(address string) *Blockchain {
	storage := NewStorage()

	// Check if blockchain exists in storage
	lastHash, err := storage.GetLastBlockHash()

	var blocks []*Block

	if err == nil && lastHash != nil {
		// Load existing blockchain
		blocks, err = storage.GetAllBlocks()
		if err != nil {
			log.Panicf("Failed to load blockchain: %v", err)
		}
		fmt.Printf("Loaded blockchain with %d blocks\n", len(blocks))
	} else {
		// Create new blockchain
		fmt.Println("No existing blockchain found. Creating genesis block...")
		coinbase := NewCoinbaseTX(address, "Genesis Block")
		genesis := CreateGenesisBlock(coinbase)

		pow := NewProofOfWork(genesis)
		nonce, hash := pow.Mine()
		genesis.Nonce = nonce
		genesis.Hash = hash

		// Save genesis block
		err := storage.SaveBlock(genesis)
		if err != nil {
			log.Panicf("Failed to save genesis block: %v", err)
		}

		blocks = []*Block{genesis}
		fmt.Printf("Genesis block created for wallet: %s\n", address)
	}

	return &Blockchain{
		Blocks:  blocks,
		Storage: storage,
	}
}

func (bc *Blockchain) AddBlock(transactions []*Transaction) {
	prevBlock := bc.Blocks[len(bc.Blocks)-1]
	newBlock := &Block{
		Timestamp:     0,
		Transactions:  transactions,
		PrevBlockHash: prevBlock.Hash,
		Hash:          []byte{},
		Nonce:         0,
		Height:        prevBlock.Height + 1,
	}

	pow := NewProofOfWork(newBlock)
	nonce, hash := pow.Mine()
	newBlock.Nonce = nonce
	newBlock.Hash = hash

	// Save to storage
	err := bc.Storage.SaveBlock(newBlock)
	if err != nil {
		log.Panicf("Failed to save block #%d: %v", newBlock.Height, err)
	}

	bc.Blocks = append(bc.Blocks, newBlock)
	fmt.Printf("Block #%d added to chain\n", newBlock.Height)
}

func (bc *Blockchain) VerifyChain() bool {
	for i := 1; i < len(bc.Blocks); i++ {
		current := bc.Blocks[i]
		previous := bc.Blocks[i-1]

		if !bytes.Equal(current.PrevBlockHash, previous.Hash) {
			log.Printf("Invalid chain: Block #%d prev hash mismatch", i)
			return false
		}

		pow := NewProofOfWork(current)
		if !pow.Validate() {
			log.Printf("Invalid chain: Block #%d invalid PoW", i)
			return false
		}
	}
	return true
}

func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspent []Transaction
	spentTXOs := make(map[string][]int)

	for _, block := range bc.Blocks {
		for _, tx := range block.Transactions {
			txID := fmt.Sprintf("%x", tx.ID)

		Outputs:
			for outIdx, out := range tx.Outputs {
				if spentTXOs[txID] != nil {
					for _, spentIdx := range spentTXOs[txID] {
						if spentIdx == outIdx {
							continue Outputs
						}
					}
				}

				if bytes.Equal(out.PubKeyHash, []byte(address)) {
					unspent = append(unspent, *tx)
				}
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					inTxID := fmt.Sprintf("%x", in.TxID)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.OutputIndex)
				}
			}
		}
	}

	return unspent
}

// GetBalance calculates the balance for an address
func (bc *Blockchain) GetBalance(address string) int {
	utxos := bc.FindUnspentTransactions(address)

	balance := 0
	for _, tx := range utxos {
		for _, output := range tx.Outputs {
			if string(output.PubKeyHash) == address {
				balance += output.Value
			}
		}
	}

	return balance
}

// AddTransaction adds a transaction to the blockchain
// For CLI transactions, we skip signature verification
func (bc *Blockchain) AddTransaction(tx *Transaction) error {
	// Add the transaction to a new block without signature verification
	// This is used for CLI transactions where we don't have the private key
	bc.AddBlock([]*Transaction{tx})
	return nil
}

// AddTransactionWithVerification adds a transaction with signature verification
// Use this for transactions from the web wallet
func (bc *Blockchain) AddTransactionWithVerification(tx *Transaction) error {
	// Verify the transaction signature
	if !VerifyTransactionSignature(tx) {
		return fmt.Errorf("invalid transaction signature")
	}

	// Add the transaction to a new block
	bc.AddBlock([]*Transaction{tx})
	return nil
}

// SendCoins creates a transaction to send coins
func (bc *Blockchain) SendCoins(fromWallet *Wallet, toAddress string, amount int) (*Transaction, error) {
	from := fromWallet.GetAddressString()

	// Get all unspent transactions for the sender
	utxos := bc.FindUnspentTransactions(from)

	if len(utxos) == 0 {
		return nil, fmt.Errorf("insufficient funds: address %s has no unspent transactions", from)
	}

	// Calculate total available balance
	available := 0
	for _, tx := range utxos {
		for _, output := range tx.Outputs {
			if string(output.PubKeyHash) == from {
				available += output.Value
			}
		}
	}

	if available < amount {
		return nil, fmt.Errorf("insufficient funds: have %d, need %d", available, amount)
	}

	// Create transaction inputs
	var inputs []TxInput
	collected := 0

	for _, tx := range utxos {
		for idx, output := range tx.Outputs {
			if string(output.PubKeyHash) == from && collected < amount {
				inputs = append(inputs, TxInput{
					TxID:        tx.ID,
					OutputIndex: idx,
					Signature:   nil,
					PubKey:      fromWallet.PublicKey,
				})
				collected += output.Value
			}
		}
	}

	// Create transaction outputs
	var outputs []TxOutput

	// Output to recipient
	outputs = append(outputs, TxOutput{
		Value:      amount,
		PubKeyHash: []byte(toAddress),
	})

	// Return change to sender
	if collected > amount {
		outputs = append(outputs, TxOutput{
			Value:      collected - amount,
			PubKeyHash: []byte(from),
		})
	}

	// Create the transaction
	tx := Transaction{
		ID:      nil,
		Inputs:  inputs,
		Outputs: outputs,
	}
	tx.ID = tx.Hash()

	return &tx, nil
}
