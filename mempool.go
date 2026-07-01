package main

import (
	"fmt"
	"sync"
	"time"
)

// TransactionStatus represents the status of a transaction
type TransactionStatus int

const (
	StatusPending TransactionStatus = iota
	StatusConfirmed
	StatusFailed
	StatusInvalid
)

// TransactionPool manages pending transactions
type TransactionPool struct {
	transactions map[string]*PoolTransaction
	mu           sync.RWMutex
	maxSize      int
	blockchain   *Blockchain
}

// PoolTransaction represents a transaction in the pool
type PoolTransaction struct {
	Transaction *Transaction
	Status      TransactionStatus
	AddedAt     time.Time
	Attempts    int
	Error       string
}

// NewTransactionPool creates a new transaction pool
func NewTransactionPool(maxSize int, bc *Blockchain) *TransactionPool {
	return &TransactionPool{
		transactions: make(map[string]*PoolTransaction),
		maxSize:      maxSize,
		blockchain:   bc,
	}
}

// AddTransaction adds a transaction to the pool
func (tp *TransactionPool) AddTransaction(tx *Transaction) error {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	// Check if pool is full
	if len(tp.transactions) >= tp.maxSize {
		return fmt.Errorf("transaction pool is full (max %d)", tp.maxSize)
	}

	txID := fmt.Sprintf("%x", tx.ID)

	// Check if transaction already exists
	if _, exists := tp.transactions[txID]; exists {
		return fmt.Errorf("transaction already in pool")
	}

	// Verify transaction
	if !VerifyTransactionSignature(tx) {
		return fmt.Errorf("invalid transaction signature")
	}

	// Add to pool
	tp.transactions[txID] = &PoolTransaction{
		Transaction: tx,
		Status:      StatusPending,
		AddedAt:     time.Now(),
		Attempts:    0,
	}

	return nil
}

// GetPendingTransactions returns all pending transactions
func (tp *TransactionPool) GetPendingTransactions() []*Transaction {
	tp.mu.RLock()
	defer tp.mu.RUnlock()

	var txs []*Transaction
	for _, pt := range tp.transactions {
		if pt.Status == StatusPending {
			txs = append(txs, pt.Transaction)
		}
	}
	return txs
}

// ConfirmTransaction marks a transaction as confirmed
func (tp *TransactionPool) ConfirmTransaction(txID []byte) error {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	key := fmt.Sprintf("%x", txID)
	pt, exists := tp.transactions[key]
	if !exists {
		return fmt.Errorf("transaction not found in pool")
	}

	pt.Status = StatusConfirmed
	return nil
}

// FailTransaction marks a transaction as failed
func (tp *TransactionPool) FailTransaction(txID []byte, errMsg string) error {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	key := fmt.Sprintf("%x", txID)
	pt, exists := tp.transactions[key]
	if !exists {
		return fmt.Errorf("transaction not found in pool")
	}

	pt.Status = StatusFailed
	pt.Error = errMsg
	pt.Attempts++
	return nil
}

// Cleanup removes old or invalid transactions
func (tp *TransactionPool) Cleanup(maxAge time.Duration) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	now := time.Now()
	for key, pt := range tp.transactions {
		// Remove old transactions
		if now.Sub(pt.AddedAt) > maxAge {
			delete(tp.transactions, key)
			continue
		}

		// Remove transactions with too many attempts
		if pt.Attempts > 3 {
			delete(tp.transactions, key)
			continue
		}
	}
}

// Size returns the current size of the pool
func (tp *TransactionPool) Size() int {
	tp.mu.RLock()
	defer tp.mu.RUnlock()
	return len(tp.transactions)
}

// IsFull checks if the pool is full
func (tp *TransactionPool) IsFull() bool {
	return tp.Size() >= tp.maxSize
}

// GetTransaction returns a transaction from the pool
func (tp *TransactionPool) GetTransaction(txID []byte) (*PoolTransaction, bool) {
	tp.mu.RLock()
	defer tp.mu.RUnlock()

	pt, exists := tp.transactions[fmt.Sprintf("%x", txID)]
	return pt, exists
}

// Clear removes all transactions from the pool
func (tp *TransactionPool) Clear() {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	tp.transactions = make(map[string]*PoolTransaction)
}

// GetStats returns statistics about the pool
func (tp *TransactionPool) GetStats() map[string]interface{} {
	tp.mu.RLock()
	defer tp.mu.RUnlock()

	stats := map[string]interface{}{
		"total":     len(tp.transactions),
		"pending":   0,
		"confirmed": 0,
		"failed":    0,
		"invalid":   0,
		"max_size":  tp.maxSize,
	}

	for _, pt := range tp.transactions {
		switch pt.Status {
		case StatusPending:
			stats["pending"] = stats["pending"].(int) + 1
		case StatusConfirmed:
			stats["confirmed"] = stats["confirmed"].(int) + 1
		case StatusFailed:
			stats["failed"] = stats["failed"].(int) + 1
		case StatusInvalid:
			stats["invalid"] = stats["invalid"].(int) + 1
		}
	}

	return stats
}

// MinePendingTransactions mines all pending transactions in a block
func (tp *TransactionPool) MinePendingTransactions(minerAddress string) error {
	pending := tp.GetPendingTransactions()
	if len(pending) == 0 {
		return fmt.Errorf("no pending transactions to mine")
	}

	// Add all pending transactions to a block
	tp.blockchain.AddBlock(pending)

	// Mark all as confirmed
	for _, tx := range pending {
		tp.ConfirmTransaction(tx.ID)
	}

	return nil
}
