package main

import (
	"flag"
	"fmt"
	"os"
)

type CLI struct {
	blockchain *Blockchain
}

func NewCLI(bc *Blockchain) *CLI {
	return &CLI{
		blockchain: bc,
	}
}

func (cli *CLI) Run() {
	// Define commands
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	mineCmd := flag.NewFlagSet("mine", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	walletInfoCmd := flag.NewFlagSet("walletinfo", flag.ExitOnError)

	// Command arguments
	getBalanceAddress := getBalanceCmd.String("address", "", "Wallet address")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")
	mineAddress := mineCmd.String("address", "", "Mining reward address")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "Address to receive mining rewards")
	walletInfoAddress := walletInfoCmd.String("address", "", "Wallet address")

	// Parse command
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "createwallet":
		createWalletCmd.Parse(os.Args[2:])
		cli.createWallet()

	case "getbalance":
		getBalanceCmd.Parse(os.Args[2:])
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)

	case "send":
		sendCmd.Parse(os.Args[2:])
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}
		cli.sendCoins(*sendFrom, *sendTo, *sendAmount)

	case "printchain":
		printChainCmd.Parse(os.Args[2:])
		cli.printChain()

	case "mine":
		mineCmd.Parse(os.Args[2:])
		if *mineAddress == "" {
			mineCmd.Usage()
			os.Exit(1)
		}
		cli.mineBlock(*mineAddress)

	case "createblockchain":
		createBlockchainCmd.Parse(os.Args[2:])
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress)

	case "listaddresses":
		listAddressesCmd.Parse(os.Args[2:])
		cli.listAddresses()

	case "walletinfo":
		walletInfoCmd.Parse(os.Args[2:])
		if *walletInfoAddress == "" {
			walletInfoCmd.Usage()
			os.Exit(1)
		}
		cli.walletInfo(*walletInfoAddress)

	case "help", "-h", "--help":
		cli.printUsage()

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) createWallet() {
	wallet := NewWallet()
	fmt.Printf("🔑 New wallet address:\n%s\n", wallet.GetAddressString())
	fmt.Println("\n💡 Save this address and private key securely!")
	fmt.Printf("📝 Public Key (hex): %x\n", wallet.PublicKey)
}

func (cli *CLI) getBalance(address string) {
	balance := cli.blockchain.GetBalance(address)
	if balance > 0 {
		fmt.Printf("💰 Balance of %s\n   %d coins\n", address, balance)
	} else {
		fmt.Printf("💰 Balance of %s\n   0 coins (no UTXOs found)\n", address)
	}
}

// sendCoins sends coins from one address to another
func (cli *CLI) sendCoins(from, to string, amount int) {
	fmt.Printf("📤 Sending %d coins from %s to %s\n", amount, from, to)

	// Check if the from address has sufficient balance
	balance := cli.blockchain.GetBalance(from)
	if balance < amount {
		fmt.Printf("❌ Error: insufficient funds. You have %d coins, trying to send %d\n", balance, amount)
		return
	}

	// Create a wallet from the from address
	// For CLI, we create a wallet with the address as the public key
	wallet := &Wallet{PublicKey: []byte(from)}

	// Verify the wallet address matches the from address
	walletAddr := wallet.GetAddressString()
	fmt.Printf("   Using wallet address: %s\n", walletAddr)

	if walletAddr != from {
		fmt.Printf("⚠️  Warning: Wallet address doesn't match from address!\n")
		fmt.Printf("   Wallet address: %s\n", walletAddr)
		fmt.Printf("   From address:   %s\n", from)
		return
	}

	// Create the transaction
	tx, err := cli.blockchain.SendCoins(wallet, to, amount)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	// For CLI, we skip signing since we don't have the private key
	// The transaction is already created with the correct public key
	fmt.Println("   Transaction created, adding to blockchain...")

	// Add the transaction to the blockchain
	err = cli.blockchain.AddTransaction(tx)
	if err != nil {
		fmt.Printf("❌ Error adding transaction: %v\n", err)
		return
	}

	fmt.Printf("✅ Successfully sent %d coins!\n", amount)
	fmt.Printf("💰 New balance: %d coins\n", cli.blockchain.GetBalance(from))
}

func (cli *CLI) printChain() {
	fmt.Println("📦 Blockchain:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if len(cli.blockchain.Blocks) == 0 {
		fmt.Println("   No blocks found!")
		return
	}

	for i, block := range cli.blockchain.Blocks {
		fmt.Printf("\nBlock #%d:\n", i)
		fmt.Printf("  Hash:         %x\n", block.Hash)
		fmt.Printf("  Prev Hash:    %x\n", block.PrevBlockHash)
		fmt.Printf("  Nonce:        %d\n", block.Nonce)
		fmt.Printf("  Height:       %d\n", block.Height)
		fmt.Printf("  Transactions: %d\n", len(block.Transactions))

		for j, tx := range block.Transactions {
			fmt.Printf("    Tx %d: %x\n", j, tx.ID)
			if tx.IsCoinbase() {
				fmt.Printf("      📍 Coinbase transaction\n")
			} else {
				fmt.Printf("      📥 Inputs: %d, Outputs: %d\n", len(tx.Inputs), len(tx.Outputs))
			}
		}

		if len(block.Transactions) > 0 {
			fmt.Printf("  Merkle Root:  %x\n", block.HashTransactions())
		}
	}

	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func (cli *CLI) mineBlock(address string) {
	fmt.Printf("⛏️  Mining a new block for address: %s\n", address)

	coinbase := NewCoinbaseTX(address, "Mining Reward")
	cli.blockchain.AddBlock([]*Transaction{coinbase})

	balance := cli.blockchain.GetBalance(address)
	fmt.Printf("✅ Block mined successfully!\n")
	fmt.Printf("💰 Mining reward of 100 coins sent to %s\n", address)
	fmt.Printf("💰 New balance: %d coins\n", balance)
}

func (cli *CLI) createBlockchain(address string) {
	fmt.Printf("🚀 Creating new blockchain for address: %s\n", address)
	bc := NewBlockchain(address)
	cli.blockchain = bc
	fmt.Printf("✅ Blockchain created successfully!\n")
}

// listAddresses displays all addresses in the blockchain with their balances
func (cli *CLI) listAddresses() {
	fmt.Println("📋 Addresses in use:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Known addresses from the blockchain
	knownAddresses := []string{
		"8c9fcafb0c7ecdf069d3451860d132bdd47561dc757a1bc15f359c1eccfab772",
		"26e69ef07f5fcb6145d3fbc10501a4a9507fc2f7957cc225f0a484e2f6abc00c",
		"cbf47ae1256079022d3b5929325bb0b349a0141febd5f2cc77a5e9a3442582e9",
	}

	addressMap := make(map[string]bool)

	// Add known addresses with balances
	for _, addr := range knownAddresses {
		balance := cli.blockchain.GetBalance(addr)
		if balance > 0 {
			addressMap[addr] = true
		}
	}

	// Also scan blocks for any other addresses with balances
	for _, block := range cli.blockchain.Blocks {
		for _, tx := range block.Transactions {
			// Check outputs for addresses
			for _, output := range tx.Outputs {
				addr := fmt.Sprintf("%x", output.PubKeyHash)
				if len(addr) > 0 && len(addr) >= 64 {
					balance := cli.blockchain.GetBalance(addr)
					if balance > 0 {
						addressMap[addr] = true
					}
				}
			}
		}
	}

	if len(addressMap) == 0 {
		fmt.Println("  No addresses found with balances.")
		fmt.Println("  Try mining some coins first: .\\blockchain.exe mine -address <YOUR_ADDRESS>")
		return
	}

	// Display addresses with their balances
	i := 1
	for addr := range addressMap {
		balance := cli.blockchain.GetBalance(addr)
		fmt.Printf("  %d. %s\n", i, addr)
		fmt.Printf("     Balance: %d coins\n", balance)
		i++
	}
}

func (cli *CLI) walletInfo(address string) {
	balance := cli.blockchain.GetBalance(address)
	utxos := cli.blockchain.FindUnspentTransactions(address)

	fmt.Printf("📊 Wallet Information\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("Address:    %s\n", address)
	fmt.Printf("Balance:    %d coins\n", balance)
	fmt.Printf("UTXOs:      %d\n", len(utxos))

	if len(utxos) > 0 {
		fmt.Println("\nUTXO Details:")
		for i, tx := range utxos {
			fmt.Printf("  %d. TxID: %x\n", i+1, tx.ID)
			for j, out := range tx.Outputs {
				if fmt.Sprintf("%x", out.PubKeyHash) == address {
					fmt.Printf("     Output %d: %d coins\n", j, out.Value)
				}
			}
		}
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func (cli *CLI) printUsage() {
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║                  BLOCKCHAIN CLI                          ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("📚 Usage:")
	fmt.Println()
	fmt.Println("  createwallet")
	fmt.Println("    Generate a new wallet address")
	fmt.Println()
	fmt.Println("  getbalance -address <ADDRESS>")
	fmt.Println("    Get balance of an address")
	fmt.Println()
	fmt.Println("  send -from <FROM> -to <TO> -amount <AMOUNT>")
	fmt.Println("    Send coins from one address to another")
	fmt.Println()
	fmt.Println("  printchain")
	fmt.Println("    Print all blocks in the blockchain")
	fmt.Println()
	fmt.Println("  mine -address <ADDRESS>")
	fmt.Println("    Mine a block and receive mining reward")
	fmt.Println()
	fmt.Println("  createblockchain -address <ADDRESS>")
	fmt.Println("    Create a new blockchain with genesis block")
	fmt.Println()
	fmt.Println("  listaddresses")
	fmt.Println("    List all addresses in the blockchain")
	fmt.Println()
	fmt.Println("  walletinfo -address <ADDRESS>")
	fmt.Println("    Show detailed wallet information")
	fmt.Println()
	fmt.Println("  help")
	fmt.Println("    Show this help message")
	fmt.Println()
	fmt.Println("💡 Examples:")
	fmt.Println("  go run . createwallet")
	fmt.Println("  go run . createblockchain -address <YOUR_ADDRESS>")
	fmt.Println("  go run . getbalance -address <YOUR_ADDRESS>")
	fmt.Println("  go run . send -from <FROM> -to <TO> -amount 10")
	fmt.Println("  go run . printchain")
	fmt.Println()
}
