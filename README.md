# Go Blockchain

A complete blockchain implementation in Go with a beautiful web wallet, CLI interface, and persistent storage.
---

## Features

| Feature | Description |
|---------|-------------|
| **Core Blockchain** | Blocks, transactions, and Proof-of-Work mining |
| **Cryptography** | ECDSA signatures with P-256 curve, SHA-256 hashing |
| **Wallet System** | Create and manage wallets with secure key generation |
| **CLI Interface** | Full command-line control for all operations |
| **Web Wallet** | Beautiful, modern web interface with real-time updates |
| **REST API** | Programmatic access to blockchain data |
| **Persistent Storage** | BoltDB for blockchain persistence |
| **UTXO Model** | Unspent transaction outputs for balance tracking |
| **Double-Spend Prevention** | Secure transaction validation |
| **Auto-Refresh** | Real-time balance updates in the web wallet |

---

## Tech Stack

<div align="center">

| Technology | Purpose |
|------------|---------|
| **Go** | Core blockchain logic |
| **BoltDB** | Persistent storage |
| **HTML/CSS/JavaScript** | Web interface |
| **Mona Sans** | Custom typography |
| **ECDSA** | Cryptographic signatures |

</div>

---

## Installation

### Prerequisites

- [Go](https://golang.org/dl/) 1.21 or higher
- [Git](https://git-scm.com/downloads)

### Clone the Repository

```bash
git clone https://github.com/yourusername/blockchain.git
cd blockchain
```

### Install Dependencies

```bash
go mod download
```

### Build the Project

```bash
go build -o blockchain.exe
```

---

## Usage

### CLI Mode
╔═══════════════════════════════════════════════════════════╗
║                  BLOCKCHAIN CLI                          ║
╚═══════════════════════════════════════════════════════════╝

```bash
# Create a new wallet
./blockchain.exe createwallet

# Check balance
./blockchain.exe getbalance -address <ADDRESS>

# Send coins
./blockchain.exe send -from <FROM> -to <TO> -amount <AMOUNT>

# Mine a block
./blockchain.exe mine -address <ADDRESS>

# View blockchain
./blockchain.exe printchain

# List all addresses
./blockchain.exe listaddresses

# Get wallet details
./blockchain.exe walletinfo -address <ADDRESS>
```

### Web Wallet Mode
╔═══════════════════════════════════════════════════════════╗
║                  BLOCKCHAIN WEB WALLET                   ║
╚═══════════════════════════════════════════════════════════╝

```bash
# Start the web server
./blockchain.exe web

# Open in browser
http://localhost:8080
```

### Interactive Demo Mode

```bash
# Run with no arguments for interactive demo
./blockchain.exe
```

---

## Commands Reference

| Command | Description |
|---------|-------------|
| `createwallet` | Generate a new wallet address |
| `getbalance -address <ADDRESS>` | Check balance of an address |
| `send -from <FROM> -to <TO> -amount <AMOUNT>` | Send coins |
| `mine -address <ADDRESS>` | Mine a new block |
| `printchain` | Display the entire blockchain |
| `listaddresses` | List all addresses with balances |
| `walletinfo -address <ADDRESS>` | Show detailed wallet info |
| `web` | Start the web wallet server |

---

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/createwallet` | GET | Create a new wallet |
| `/api/balance?address=<ADDRESS>` | GET | Get wallet balance |
| `/api/send` | POST | Send coins |
| `/api/blocks` | GET | Get all blocks |
| `/api/transaction?txid=<TXID>` | GET | Get transaction details |
| `/api/mine?address=<ADDRESS>` | GET | Mine a block |

---

## Project Structure

```
blockchain/
├── main.go              # Entry point
├── web_server.go        # Web wallet server
├── cli.go               # CLI interface
├── wallet.go            # Wallet management
├── blockchain.go        # Core blockchain logic
├── block.go             # Block structure
├── transaction.go       # Transaction handling
├── proofofwork.go       # Mining algorithm
├── merkle.go            # Merkle tree
├── storage.go           # Database operations
├── webwallet/           # Web wallet files
│   ├── templates/       # HTML templates
│   │   ├── index.html
│   │   ├── wallet.html
│   │   ├── send.html
│   │   └── explorer.html
│   └── static/          # CSS, JS, fonts
│       ├── css/
│       │   └── style.css
│       ├── js/
│       │   └── app.js
│       └── fonts/
│           └── Mona-Sans.var.woff2
└── go.mod               # Dependencies
```

---

## Security Features

| Feature | Description |
|---------|-------------|
| **ECDSA Digital Signatures** | P-256 curve for transaction signing |
| **SHA-256 Hashing** | Cryptographic hashing for blocks |
| **Proof-of-Work** | Mining difficulty for network security |
| **UTXO Model** | Prevents double-spending |
| **Blockchain Verification** | Chain integrity validation |

---

## Testing

```bash
# Run all tests
go test ./...

# Run specific test
go test -v -run TestBlockchain
```

---

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- Go community for the excellent language and tools
- BoltDB for simple and fast database
- All contributors and users of this project

---



### Built using Go

[⬆ Back to Top](#-go-blockchain)



