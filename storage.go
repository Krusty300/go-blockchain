package main

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/boltdb/bolt"
)

const (
	dbFile       = "blockchain.db"
	blocksBucket = "blocks"
	utxoBucket   = "utxos"
	lastBlockKey = "l"
)

// Storage handles all database operations
type Storage struct {
	db   *bolt.DB
	lock sync.RWMutex
}

// NewStorage creates a new storage instance
func NewStorage() *Storage {
	// Ensure directory exists
	dir := filepath.Dir(dbFile)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panicf("Failed to open database: %v", err)
	}

	// Create buckets if they don't exist
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(blocksBucket))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(utxoBucket))
		return err
	})

	if err != nil {
		log.Panicf("Failed to create buckets: %v", err)
	}

	return &Storage{db: db}
}

// SaveBlock saves a block to the database
func (s *Storage) SaveBlock(block *Block) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		if bucket == nil {
			return nil
		}

		serialized := block.Serialize()
		err := bucket.Put(block.Hash, serialized)
		if err != nil {
			return err
		}

		// Save last block hash
		err = bucket.Put([]byte(lastBlockKey), block.Hash)
		return err
	})
}

// GetBlock retrieves a block by its hash
func (s *Storage) GetBlock(hash []byte) (*Block, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var block *Block
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		if bucket == nil {
			return nil
		}

		data := bucket.Get(hash)
		if data == nil {
			return nil
		}

		block = DeserializeBlock(data)
		return nil
	})

	return block, err
}

// GetLastBlockHash returns the hash of the last block
func (s *Storage) GetLastBlockHash() ([]byte, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var lastHash []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		if bucket == nil {
			return nil
		}

		lastHash = bucket.Get([]byte(lastBlockKey))
		return nil
	})

	return lastHash, err
}

// GetAllBlocks returns all blocks in the chain
func (s *Storage) GetAllBlocks() ([]*Block, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var blocks []*Block
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		if bucket == nil {
			return nil
		}

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			// Skip the last block key
			if string(k) == lastBlockKey {
				continue
			}
			block := DeserializeBlock(v)
			blocks = append(blocks, block)
		}
		return nil
	})

	return blocks, err
}

// SaveUTXO saves UTXO set for an address
func (s *Storage) SaveUTXO(address string, utxos []UTXO) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket))
		if bucket == nil {
			return nil
		}

		var data []byte
		if len(utxos) > 0 {
			var err error
			data, err = serializeUTXOs(utxos)
			if err != nil {
				return err
			}
		}

		return bucket.Put([]byte(address), data)
	})
}

// GetUTXOs retrieves UTXOs for an address
func (s *Storage) GetUTXOs(address string) ([]UTXO, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var utxos []UTXO
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket))
		if bucket == nil {
			return nil
		}

		data := bucket.Get([]byte(address))
		if data == nil {
			return nil
		}

		var err error
		utxos, err = deserializeUTXOs(data)
		return err
	})

	return utxos, err
}

// Close closes the database connection
func (s *Storage) Close() error {
	return s.db.Close()
}

// UTXO represents an unspent transaction output
type UTXO struct {
	TxID       []byte
	Index      int
	Value      int
	PubKeyHash []byte
}

// serializeUTXOs serializes a slice of UTXOs
func serializeUTXOs(utxos []UTXO) ([]byte, error) {
	var result []byte
	for _, utxo := range utxos {
		result = append(result, utxo.TxID...)
		result = append(result, byte(utxo.Index))
		result = append(result, byte(utxo.Value>>24), byte(utxo.Value>>16), byte(utxo.Value>>8), byte(utxo.Value))
		result = append(result, utxo.PubKeyHash...)
	}
	return result, nil
}

// deserializeUTXOs deserializes UTXOs from bytes
func deserializeUTXOs(data []byte) ([]UTXO, error) {
	var utxos []UTXO
	i := 0
	for i < len(data) {
		if i+32+1+4 > len(data) {
			break
		}

		utxo := UTXO{
			TxID:  data[i : i+32],
			Index: int(data[i+32]),
			Value: int(data[i+33])<<24 | int(data[i+34])<<16 | int(data[i+35])<<8 | int(data[i+36]),
		}

		i += 37

		if i+32 <= len(data) {
			utxo.PubKeyHash = data[i : i+32]
			i += 32
		}

		utxos = append(utxos, utxo)
	}
	return utxos, nil
}
