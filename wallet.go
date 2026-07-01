package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
)

// Wallet represents a cryptocurrency wallet with private/public keys
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// NewWallet creates and returns a new wallet
func NewWallet() *Wallet {
	privateKey, publicKey := generateKeyPair()
	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}
}

// generateKeyPair creates a new ECDSA key pair using P-256 curve
func generateKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic("Failed to generate private key:", err)
	}

	// Pad X and Y coordinates to 32 bytes each
	xBytes := privateKey.PublicKey.X.Bytes()
	yBytes := privateKey.PublicKey.Y.Bytes()

	// Pad to 32 bytes
	xPadded := make([]byte, 32)
	yPadded := make([]byte, 32)

	copy(xPadded[32-len(xBytes):], xBytes)
	copy(yPadded[32-len(yBytes):], yBytes)

	// Combine X and Y coordinates to form the public key
	publicKey := append(xPadded, yPadded...)

	return *privateKey, publicKey
}

// GetAddress returns the wallet address
func (w Wallet) GetAddress() []byte {
	// Check if PublicKey is a hex string address (for CLI usage)
	// The address is 64 hex characters = 32 bytes when decoded
	if len(w.PublicKey) == 32 {
		// Try to decode as hex to see if it's a valid address
		hexStr := string(w.PublicKey)
		// Check if it's a valid hex string of length 64
		if len(hexStr) == 64 {
			_, err := hex.DecodeString(hexStr)
			if err == nil {
				// It's a valid hex address, use it directly
				return w.PublicKey
			}
		}
	}

	// Also check if PublicKey is the raw hex string (64 bytes)
	if len(w.PublicKey) == 64 {
		// Check if it's a valid hex string
		hexStr := string(w.PublicKey)
		_, err := hex.DecodeString(hexStr)
		if err == nil {
			// It's a valid hex string, use it as address
			return w.PublicKey
		}
	}

	// Normal case: SHA-256 hash of the public key
	hash := sha256.Sum256(w.PublicKey)
	return hash[:]
}

// GetAddressString returns the wallet address as a hex string
func (w Wallet) GetAddressString() string {
	addr := w.GetAddress()

	// If the address is already a hex string (64 characters), return it as-is
	if len(addr) == 64 {
		// Check if it's a valid hex string
		_, err := hex.DecodeString(string(addr))
		if err == nil {
			return string(addr)
		}
	}

	// If the address is 32 bytes, it might be a hex string encoded
	if len(addr) == 32 {
		// Try to interpret as hex string
		hexStr := string(addr)
		if len(hexStr) == 64 {
			_, err := hex.DecodeString(hexStr)
			if err == nil {
				return hexStr
			}
		}
	}

	// Otherwise, return hex representation
	return fmt.Sprintf("%x", addr)
}

// SignTransaction signs a transaction using the wallet's private key
func (w *Wallet) SignTransaction(tx *Transaction) error {
	// Skip coinbase transactions (they don't need signing)
	if tx.IsCoinbase() {
		return nil
	}

	// Create a trimmed copy of the transaction (without signatures)
	txCopy := tx.TrimmedCopy()

	// Hash the transaction
	txHash := txCopy.Hash()

	// Sign the hash using ECDSA
	r, s, err := ecdsa.Sign(rand.Reader, &w.PrivateKey, txHash)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %v", err)
	}

	// Create signature by combining r and s (each is 32 bytes)
	signature := append(r.Bytes(), s.Bytes()...)

	// Apply signature and public key to all inputs
	for i := range tx.Inputs {
		tx.Inputs[i].Signature = signature
		tx.Inputs[i].PubKey = w.PublicKey
	}

	return nil
}

// VerifyTransactionSignature verifies all signatures in a transaction
func VerifyTransactionSignature(tx *Transaction) bool {
	// Coinbase transactions don't need verification
	if tx.IsCoinbase() {
		return true
	}

	// Create a copy without signatures for verification
	txCopy := tx.TrimmedCopy()
	txHash := txCopy.Hash()

	// Verify each input
	for _, input := range tx.Inputs {
		// Check if signature and public key exist
		if len(input.Signature) == 0 {
			log.Printf("Missing signature for input")
			return false
		}
		if len(input.PubKey) == 0 {
			log.Printf("Missing public key for input")
			return false
		}

		// Check if public key is valid (64 bytes for ECDSA P-256)
		if len(input.PubKey) != 64 {
			log.Printf("Public key wrong length: %d bytes (expected 64)", len(input.PubKey))
			return false
		}

		// Extract r and s from signature
		sigLen := len(input.Signature)
		if sigLen < 64 {
			log.Printf("Signature too short: %d bytes", sigLen)
			return false
		}

		// For P-256, r and s are each 32 bytes
		sigOffset := 0
		if sigLen > 64 {
			sigOffset = sigLen - 64
		}

		r := big.Int{}
		s := big.Int{}
		r.SetBytes(input.Signature[sigOffset : sigOffset+32])
		s.SetBytes(input.Signature[sigOffset+32:])

		// Extract public key (first 32 bytes = X, last 32 bytes = Y)
		x := big.Int{}
		y := big.Int{}
		x.SetBytes(input.PubKey[:32])
		y.SetBytes(input.PubKey[32:])

		// Create the public key
		pubKey := ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     &x,
			Y:     &y,
		}

		// Verify the signature
		if !ecdsa.Verify(&pubKey, txHash, &r, &s) {
			log.Printf("Signature verification failed for input")
			return false
		}
	}

	return true
}

// ValidateAddress checks if a given address is valid
func ValidateAddress(address string) bool {
	if len(address) == 0 {
		return false
	}
	// Check if it's a valid hex string
	if len(address) == 64 {
		_, err := hex.DecodeString(address)
		return err == nil
	}
	return false
}

// String returns a string representation of the wallet
func (w Wallet) String() string {
	addr := w.GetAddressString()
	pubKeyStr := ""
	if len(w.PublicKey) > 10 {
		pubKeyStr = fmt.Sprintf("%x...", w.PublicKey[:10])
	} else {
		pubKeyStr = fmt.Sprintf("%x", w.PublicKey)
	}
	return fmt.Sprintf("Wallet{Address: %s, PublicKey: %s}", addr, pubKeyStr)
}

// SerializeWallet serializes a wallet to bytes
func (w *Wallet) Serialize() []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	type WalletData struct {
		PublicKey []byte
	}

	data := WalletData{
		PublicKey: w.PublicKey,
	}

	err := enc.Encode(data)
	if err != nil {
		log.Printf("Failed to serialize wallet: %v", err)
		return []byte{}
	}

	return buf.Bytes()
}

// DeserializeWallet creates a wallet from bytes
func DeserializeWallet(data []byte) *Wallet {
	if len(data) == 0 {
		return nil
	}

	buf := bytes.NewReader(data)
	dec := gob.NewDecoder(buf)

	type WalletData struct {
		PublicKey []byte
	}

	var wd WalletData
	err := dec.Decode(&wd)
	if err != nil {
		log.Printf("Failed to deserialize wallet: %v", err)
		return nil
	}

	// Try to create a wallet from the deserialized data
	wallet := NewWallet()
	wallet.PublicKey = wd.PublicKey

	return wallet
}

// VerifyTransaction verifies a transaction (public wrapper)
func VerifyTransaction(tx *Transaction) bool {
	return VerifyTransactionSignature(tx)
}

// IsValidAddress checks if an address is valid
func IsValidAddress(address string) bool {
	return ValidateAddress(address)
}

// NewWalletFromAddress creates a wallet from an address string (for CLI use)
func NewWalletFromAddress(address string) *Wallet {
	// Check if address is a valid hex string
	if len(address) == 64 {
		_, err := hex.DecodeString(address)
		if err == nil {
			// Create a wallet with the address as the public key
			return &Wallet{
				PublicKey: []byte(address),
			}
		}
	}
	// If not a valid hex address, create a new wallet
	return NewWallet()
}
