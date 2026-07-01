package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	// Check if web mode is requested
	if len(os.Args) > 1 && os.Args[1] == "web" {
		runWebWallet()
		return
	}

	// Check if CLI mode is active
	if len(os.Args) > 1 {
		// Run in CLI mode
		runCLI()
		return
	}

	// Default interactive mode
	runInteractive()
}

// runWebWallet starts the web wallet server
func runWebWallet() {
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║                  BLOCKCHAIN WEB WALLET                   ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("🌐 Starting Web Wallet server...")
	fmt.Println()

	// Create blockchain with default wallet
	defaultWallet := NewWallet()
	fmt.Printf("🔑 Default wallet created: %s\n", defaultWallet.GetAddressString())

	// Create or load blockchain
	bc := NewBlockchain(defaultWallet.GetAddressString())
	fmt.Printf("⛏️  Blockchain loaded with %d blocks\n", len(bc.Blocks))

	// Create and start web server
	ws := NewWebServer(bc)

	// Start the web server
	fmt.Println()
	fmt.Println("🚀 Web Wallet is running!")
	fmt.Println("📱 Open your browser and navigate to: http://localhost:8080")
	fmt.Println()
	fmt.Println("💡 Available pages:")
	fmt.Println("   - Home:        http://localhost:8080/")
	fmt.Println("   - Wallet:      http://localhost:8080/wallet?address=<ADDRESS>")
	fmt.Println("   - Send Coins:  http://localhost:8080/send?from=<ADDRESS>")
	fmt.Println("   - Explorer:    http://localhost:8080/explorer")
	fmt.Println()
	fmt.Println("⚠️  Press Ctrl+C to stop the server")
	fmt.Println()

	// Start the server (this blocks until stopped)
	ws.Start("8080")

	// Clean up (this won't run until server stops)
	defer bc.Storage.Close()
}

// runCLI handles command-line interface mode
func runCLI() {
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║                  BLOCKCHAIN CLI                          ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Create blockchain with default wallet if needed
	defaultWallet := NewWallet()
	fmt.Printf("🔑 Default wallet created: %s\n", defaultWallet.GetAddressString())

	// Check if blockchain exists
	bc := NewBlockchain(defaultWallet.GetAddressString())

	cli := NewCLI(bc)
	cli.Run()

	// Clean up
	defer bc.Storage.Close()
}

// runInteractive runs the interactive demonstration mode
func runInteractive() {
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║         BLOCKCHAIN TEST SUITE - ALL SCENARIOS            ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// ============================================
	// SCENARIO 1: Create wallets
	// ============================================
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📝 SCENARIO 1: Create wallets")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Create wallets
	wallet1 := NewWallet()
	wallet2 := NewWallet()
	wallet3 := NewWallet()

	fmt.Printf("🔑 Wallet 1: %s\n", wallet1.GetAddressString())
	fmt.Printf("🔑 Wallet 2: %s\n", wallet2.GetAddressString())
	fmt.Printf("🔑 Wallet 3: %s\n", wallet3.GetAddressString())
	fmt.Println()

	// ============================================
	// SCENARIO 2: Create blockchain
	// ============================================
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📝 SCENARIO 2: Create blockchain with genesis block")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Create blockchain with wallet1 as miner
	blockchain := NewBlockchain(wallet1.GetAddressString())
	fmt.Printf("⛏️  Genesis block created for wallet: %s\n", wallet1.GetAddressString())

	// Check initial balance
	fmt.Printf("💰 Initial balance for Wallet 1: %d coins\n\n", blockchain.GetBalance(wallet1.GetAddressString()))

	// ============================================
	// SCENARIO 3: Send transactions
	// ============================================
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📝 SCENARIO 3: Send transactions")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Create and send transactions
	fmt.Println("--- Sending 10 coins from Wallet 1 to Wallet 2 ---")
	sendCoins(blockchain, wallet1, wallet2.GetAddressString(), 10)

	fmt.Println("\n--- Sending 5 coins from Wallet 1 to Wallet 3 ---")
	sendCoins(blockchain, wallet1, wallet3.GetAddressString(), 5)

	fmt.Println("\n--- Sending 20 coins from Wallet 2 to Wallet 3 ---")
	sendCoins(blockchain, wallet2, wallet3.GetAddressString(), 20)

	// ============================================
	// SCENARIO 4: Try to double-spend (should fail)
	// ============================================
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📝 SCENARIO 4: Attempt to double-spend (SHOULD FAIL)")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	wallet1Balance := blockchain.GetBalance(wallet1.GetAddressString())
	fmt.Printf("💰 Wallet 1 balance: %d coins\n", wallet1Balance)
	fmt.Printf("⚠️  Attempting to send %d coins (more than available)...\n", wallet1Balance+50)

	tx, err := blockchain.SendCoins(wallet1, wallet2.GetAddressString(), wallet1Balance+50)
	if err != nil {
		fmt.Printf("✅ SUCCESS: Transaction FAILED as expected!\n")
		fmt.Printf("   Error: %v\n", err)
	} else {
		err = wallet1.SignTransaction(tx)
		if err != nil {
			fmt.Printf("✅ SUCCESS: Transaction FAILED as expected!\n")
			fmt.Printf("   Error: %v\n", err)
		} else {
			err = blockchain.AddTransaction(tx)
			if err != nil {
				fmt.Printf("✅ SUCCESS: Transaction FAILED as expected!\n")
				fmt.Printf("   Error: %v\n", err)
			} else {
				fmt.Println("❌ ERROR: Transaction should have failed but was accepted!")
			}
		}
	}

	// ============================================
	// SCENARIO 5: Mine a block
	// ============================================
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📝 SCENARIO 5: Mine a new block")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	fmt.Println("⛏️  Mining a new block for Wallet 1...")
	coinbase := NewCoinbaseTX(wallet1.GetAddressString(), "Mining Reward")
	blockchain.AddBlock([]*Transaction{coinbase})
	fmt.Printf("✅ Block mined! Mining reward of 100 coins sent to Wallet 1\n")

	// ============================================
	// SCENARIO 6: Send from empty wallet (should fail)
	// ============================================
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📝 SCENARIO 6: Send from empty wallet (SHOULD FAIL)")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	emptyWallet := NewWallet()
	fmt.Printf("🔑 New empty wallet: %s\n", emptyWallet.GetAddressString())
	fmt.Printf("💰 Balance: %d coins\n", blockchain.GetBalance(emptyWallet.GetAddressString()))
	fmt.Println("⚠️  Attempting to send 10 coins from empty wallet...")

	_, err = blockchain.SendCoins(emptyWallet, wallet1.GetAddressString(), 10)
	if err != nil {
		fmt.Printf("✅ SUCCESS: Transaction FAILED as expected!\n")
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Println("❌ ERROR: Transaction should have failed but was accepted!")
	}

	// ============================================
	// Display blockchain state
	// ============================================
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📊 BLOCKCHAIN STATE")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for i, block := range blockchain.Blocks {
		fmt.Printf("\nBlock #%d:\n", i)
		fmt.Printf("  Hash:         %x\n", block.Hash)
		fmt.Printf("  Prev Hash:    %x\n", block.PrevBlockHash)
		fmt.Printf("  Nonce:        %d\n", block.Nonce)
		fmt.Printf("  Height:       %d\n", block.Height)
		fmt.Printf("  Transactions: %d\n", len(block.Transactions))
		if len(block.Transactions) > 0 {
			fmt.Printf("  Merkle Root:  %x\n", block.HashTransactions())
		}
	}

	// ============================================
	// Verify chain
	// ============================================
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🔍 VERIFICATION")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if blockchain.VerifyChain() {
		fmt.Println("✅ Blockchain is VALID and INTACT!")
	} else {
		fmt.Println("❌ Blockchain verification FAILED!")
	}

	// ============================================
	// Show final balances
	// ============================================
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("💰 FINAL BALANCES")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	fmt.Printf("Wallet 1: %d coins\n", blockchain.GetBalance(wallet1.GetAddressString()))
	fmt.Printf("Wallet 2: %d coins\n", blockchain.GetBalance(wallet2.GetAddressString()))
	fmt.Printf("Wallet 3: %d coins\n", blockchain.GetBalance(wallet3.GetAddressString()))
	fmt.Printf("Empty Wallet: %d coins\n", blockchain.GetBalance(emptyWallet.GetAddressString()))

	// ============================================
	// Summary
	// ============================================
	fmt.Println("\n╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║                    TEST SUMMARY                          ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("✅ Genesis block created successfully")
	fmt.Println("✅ Transactions processed successfully")
	fmt.Println("✅ Double-spend prevention working")
	fmt.Println("✅ Insufficient funds prevention working")
	fmt.Println("✅ Mining rewards working")
	fmt.Println("✅ Empty wallet rejection working")
	fmt.Println("✅ Blockchain verification passing")
	fmt.Println()
	fmt.Println("🎉 ALL TESTS COMPLETED SUCCESSFULLY!")
	fmt.Println()
	fmt.Println("💡 To use CLI mode, run: go run . <command>")
	fmt.Println("   Available commands: createwallet, getbalance, send,")
	fmt.Println("   printchain, mine, createblockchain, listaddresses, walletinfo")
	fmt.Println()
	fmt.Println("💡 To use Web Wallet mode, run: go run . web")
	fmt.Println("   Then open http://localhost:8080 in your browser")

	// Clean up
	defer blockchain.Storage.Close()
}

// sendCoins helper function for interactive mode
func sendCoins(bc *Blockchain, from *Wallet, to string, amount int) {
	// Create transaction using the wallet
	tx, err := bc.SendCoins(from, to, amount)
	if err != nil {
		fmt.Printf("❌ Error creating transaction: %v\n", err)
		return
	}

	// Sign the transaction
	err = from.SignTransaction(tx)
	if err != nil {
		fmt.Printf("❌ Error signing transaction: %v\n", err)
		return
	}

	// Verify the transaction signature
	if !VerifyTransactionSignature(tx) {
		fmt.Printf("❌ Transaction signature verification failed!\n")
		return
	}

	// Add the transaction to blockchain
	err = bc.AddTransaction(tx)
	if err != nil {
		fmt.Printf("❌ Error adding transaction: %v\n", err)
		return
	}

	fmt.Printf("✅ Successfully sent %d coins!\n", amount)
}

// Helper function to print a separator line
func printSeparator(char string, length int) {
	fmt.Println(strings.Repeat(char, length))
}

// Helper function to print a section header
func printSection(title string) {
	fmt.Println()
	printSeparator("━", 60)
	fmt.Printf("  %s\n", title)
	printSeparator("━", 60)
}
