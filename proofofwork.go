package main

import (
    "bytes"
    "crypto/sha256"
    "fmt"
    "math"
    "math/big"
)

const (
    targetBits = 24 // Difficulty level (lower = harder)
    maxNonce   = math.MaxInt64
)

type ProofOfWork struct {
    Block  *Block
    Target *big.Int
}

func NewProofOfWork(b *Block) *ProofOfWork {
    target := big.NewInt(1)
    target.Lsh(target, uint(256-targetBits))
    
    return &ProofOfWork{
        Block:  b,
        Target: target,
    }
}

// Prepare block data for hashing
func (pow *ProofOfWork) prepareData(nonce int) []byte {
    data := bytes.Join(
        [][]byte{
            pow.Block.PrevBlockHash,
            pow.Block.HashTransactions(),
            IntToHex(pow.Block.Timestamp),
            IntToHex(int64(targetBits)),
            IntToHex(int64(nonce)),
        },
        []byte{},
    )
    return data
}

// Mine block - find valid nonce
func (pow *ProofOfWork) Mine() (int, []byte) {
    var hashInt big.Int
    var hash [32]byte
    nonce := 0
    
    fmt.Printf("Mining block with transactions...\n")
    
    for nonce < maxNonce {
        data := pow.prepareData(nonce)
        hash = sha256.Sum256(data)
        hashInt.SetBytes(hash[:])
        
        if hashInt.Cmp(pow.Target) == -1 {
            fmt.Printf("Block mined! Hash: %x\n", hash)
            break
        } else {
            nonce++
        }
    }
    
    return nonce, hash[:]
}

// Verify block's proof of work
func (pow *ProofOfWork) Validate() bool {
    var hashInt big.Int
    
    data := pow.prepareData(pow.Block.Nonce)
    hash := sha256.Sum256(data)
    hashInt.SetBytes(hash[:])
    
    isValid := hashInt.Cmp(pow.Target) == -1
    return isValid
}

// Helper: Convert int64 to bytes
func IntToHex(n int64) []byte {
    return []byte(fmt.Sprintf("%x", n))
}