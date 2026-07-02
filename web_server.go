package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/skip2/go-qrcode"
)

// WebServer handles the web wallet interface
type WebServer struct {
	blockchain *Blockchain
	templates  *template.Template
}

// NewWebServer creates a new web server
func NewWebServer(bc *Blockchain) *WebServer {
	ws := &WebServer{
		blockchain: bc,
	}

	// Create a custom template with functions
	funcMap := template.FuncMap{
		"formatTimestamp": func(timestamp int64) string {
			if timestamp == 0 {
				return "Genesis Block"
			}
			t := time.Unix(timestamp, 0)
			return t.Format("2006-01-02 15:04:05")
		},
		"formatDate": func(timestamp int64) string {
			if timestamp == 0 {
				return "Genesis"
			}
			t := time.Unix(timestamp, 0)
			return t.Format("Jan 2, 2006")
		},
		"formatTime": func(timestamp int64) string {
			if timestamp == 0 {
				return ""
			}
			t := time.Unix(timestamp, 0)
			return t.Format("15:04:05")
		},
		"formatBlockTime": func(timestamp int64) string {
			if timestamp == 0 {
				return "Genesis Block"
			}
			t := time.Unix(timestamp, 0)
			now := time.Now()
			if t.Format("2006-01-02") == now.Format("2006-01-02") {
				return "Today at " + t.Format("15:04:05")
			}
			return t.Format("Jan 2, 2006 at 15:04:05")
		},
		"timeAgo": func(timestamp int64) string {
			if timestamp == 0 {
				return "Genesis"
			}
			t := time.Unix(timestamp, 0)
			duration := time.Since(t)
			if duration < time.Minute {
				return "Just now"
			} else if duration < time.Hour {
				return fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
			} else if duration < 24*time.Hour {
				return fmt.Sprintf("%d hours ago", int(duration.Hours()))
			} else {
				days := int(duration.Hours() / 24)
				if days == 1 {
					return "Yesterday"
				}
				return fmt.Sprintf("%d days ago", days)
			}
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"lastBlock": func(blocks []*Block) *Block {
			if len(blocks) == 0 {
				return nil
			}
			return blocks[len(blocks)-1]
		},
	}

	// Parse templates with functions
	var err error
	ws.templates, err = template.New("").
		Funcs(funcMap).
		ParseGlob("webwallet/templates/*.html")
	if err != nil {
		fmt.Printf("❌ Error loading templates: %v\n", err)
		panic(err)
	}

	// Debug: List loaded templates
	fmt.Println("✅ Templates loaded successfully:")
	for _, t := range ws.templates.Templates() {
		fmt.Printf("   - %s\n", t.Name())
	}

	return ws
}

// Start starts the web server
func (ws *WebServer) Start(port string) {
	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("webwallet/static"))))

	// Routes
	http.HandleFunc("/", ws.indexHandler)
	http.HandleFunc("/wallet", ws.walletHandler)
	http.HandleFunc("/send", ws.sendHandler)
	http.HandleFunc("/explorer", ws.explorerHandler)
	http.HandleFunc("/qr", ws.qrCodeHandler) // QR Code handler
	http.HandleFunc("/api/createwallet", ws.createWalletAPI)
	http.HandleFunc("/api/balance", ws.getBalanceAPI)
	http.HandleFunc("/api/send", ws.sendCoinsAPI)
	http.HandleFunc("/api/blocks", ws.getBlocksAPI)
	http.HandleFunc("/api/transaction", ws.getTransactionAPI)
	http.HandleFunc("/api/mine", ws.mineAPI)

	fmt.Printf("🌐 Web Wallet running on http://localhost:%s\n", port)
	fmt.Printf("📱 Open your browser and navigate to http://localhost:%s\n", port)
	fmt.Println()
	fmt.Println("💡 Available pages:")
	fmt.Printf("   - Home:        http://localhost:%s/\n", port)
	fmt.Printf("   - Wallet:      http://localhost:%s/wallet?address=<ADDRESS>\n", port)
	fmt.Printf("   - Send Coins:  http://localhost:%s/send?from=<ADDRESS>\n", port)
	fmt.Printf("   - Explorer:    http://localhost:%s/explorer\n", port)
	fmt.Printf("   - QR Code:     http://localhost:%s/qr?address=<ADDRESS>\n", port)
	fmt.Println()
	fmt.Println("⚠️  Press Ctrl+C to stop the server")

	http.ListenAndServe(":"+port, nil)
}

// indexHandler serves the home page
func (ws *WebServer) indexHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title   string
		Message string
	}{
		Title:   "Blockchain Web Wallet",
		Message: "Welcome to your blockchain wallet!",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := ws.templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		fmt.Printf("❌ Error executing index.html: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// walletHandler serves the wallet page
func (ws *WebServer) walletHandler(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	balance := ws.blockchain.GetBalance(address)

	// Get transaction history
	var transactions []TransactionDisplay
	processedTxIDs := make(map[string]bool)

	for _, block := range ws.blockchain.Blocks {
		for _, tx := range block.Transactions {
			txID := fmt.Sprintf("%x", tx.ID)
			if processedTxIDs[txID] {
				continue
			}

			if tx.IsCoinbase() {
				// Check if address is the recipient of coinbase
				for _, out := range tx.Outputs {
					if string(out.PubKeyHash) == address {
						transactions = append(transactions, TransactionDisplay{
							TxID:      txID,
							To:        address,
							Amount:    out.Value,
							Type:      "Mining Reward",
							Block:     block.Height,
							Time:      formatTimestamp(block.Timestamp),
							Timestamp: block.Timestamp,
						})
						processedTxIDs[txID] = true
						break
					}
				}
				continue
			}

			// Check if this address is the sender (look in inputs)
			isSender := false
			for _, in := range tx.Inputs {
				if string(in.PubKey) == address {
					isSender = true
					break
				}
			}

			// Check if this address is the receiver (look in outputs)
			isReceiver := false
			var receivedAmount int
			for _, out := range tx.Outputs {
				if string(out.PubKeyHash) == address {
					isReceiver = true
					receivedAmount = out.Value
				}
			}

			// If address is sender (not receiver)
			if isSender && !isReceiver {
				// Find the recipient (first output that's not the sender)
				var recipient string
				var sentAmount int
				for _, out := range tx.Outputs {
					if string(out.PubKeyHash) != address {
						recipient = string(out.PubKeyHash)
						sentAmount = out.Value
						break
					}
				}

				if recipient != "" && sentAmount > 0 {
					transactions = append(transactions, TransactionDisplay{
						TxID:      txID,
						From:      address,
						To:        recipient,
						Amount:    sentAmount,
						Type:      "Sent",
						Block:     block.Height,
						Time:      formatTimestamp(block.Timestamp),
						Timestamp: block.Timestamp,
					})
					processedTxIDs[txID] = true
				}
			}

			// If address is receiver (not sender)
			if isReceiver && !isSender {
				// Find who sent it (get from inputs)
				var from string
				for _, in := range tx.Inputs {
					from = string(in.PubKey)
					break
				}

				if receivedAmount > 0 && from != "" {
					transactions = append(transactions, TransactionDisplay{
						TxID:      txID,
						From:      from,
						To:        address,
						Amount:    receivedAmount,
						Type:      "Received",
						Block:     block.Height,
						Time:      formatTimestamp(block.Timestamp),
						Timestamp: block.Timestamp,
					})
					processedTxIDs[txID] = true
				}
			}

			// If address is both sender and receiver (change), skip
			if isSender && isReceiver {
				continue
			}
		}
	}

	data := WalletPageData{
		Address:      address,
		Balance:      balance,
		Transactions: transactions,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := ws.templates.ExecuteTemplate(w, "wallet.html", data)
	if err != nil {
		fmt.Printf("❌ Error executing wallet.html: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// sendHandler serves the send coins page
func (ws *WebServer) sendHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("📝 sendHandler called with method: %s\n", r.Method)
	fmt.Printf("   URL: %s\n", r.URL.String())

	if r.Method == "POST" {
		from := r.FormValue("from")
		to := r.FormValue("to")
		amountStr := r.FormValue("amount")

		fmt.Printf("   POST data - from: %s, to: %s, amount: %s\n", from, to, amountStr)

		amount, err := strconv.Atoi(amountStr)
		if err != nil {
			data := struct {
				From   string
				To     string
				Amount int
				Error  string
			}{
				From:   from,
				To:     to,
				Amount: amount,
				Error:  "Invalid amount",
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			ws.templates.ExecuteTemplate(w, "send.html", data)
			return
		}

		// Check balance first
		balance := ws.blockchain.GetBalance(from)
		if balance < amount {
			data := struct {
				From   string
				To     string
				Amount int
				Error  string
			}{
				From:   from,
				To:     to,
				Amount: amount,
				Error:  fmt.Sprintf("Insufficient funds. You have %d coins, trying to send %d", balance, amount),
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			ws.templates.ExecuteTemplate(w, "send.html", data)
			return
		}

		// Create a wallet from the from address
		wallet := &Wallet{PublicKey: []byte(from)}

		// Create the transaction
		tx, err := ws.blockchain.SendCoins(wallet, to, amount)
		if err != nil {
			data := struct {
				From   string
				To     string
				Amount int
				Error  string
			}{
				From:   from,
				To:     to,
				Amount: amount,
				Error:  err.Error(),
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			ws.templates.ExecuteTemplate(w, "send.html", data)
			return
		}

		// For web wallet, skip signing since we don't have the private key
		fmt.Println("   Transaction created, adding to blockchain...")

		// Add the transaction to the blockchain
		err = ws.blockchain.AddTransaction(tx)
		if err != nil {
			data := struct {
				From   string
				To     string
				Amount int
				Error  string
			}{
				From:   from,
				To:     to,
				Amount: amount,
				Error:  err.Error(),
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			ws.templates.ExecuteTemplate(w, "send.html", data)
			return
		}

		http.Redirect(w, r, "/wallet?address="+from, http.StatusSeeOther)
		return
	}

	// GET request
	from := r.URL.Query().Get("from")
	fmt.Printf("   GET from: %s\n", from)

	if from == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := struct {
		From  string
		Error string
	}{
		From:  from,
		Error: "",
	}

	fmt.Printf("   Rendering send.html template with data: %+v\n", data)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := ws.templates.ExecuteTemplate(w, "send.html", data)
	if err != nil {
		fmt.Printf("❌ Error executing send.html: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		fmt.Println("✅ send.html rendered successfully")
	}
}

// explorerHandler serves the block explorer page
func (ws *WebServer) explorerHandler(w http.ResponseWriter, r *http.Request) {
	totalSupply := len(ws.blockchain.Blocks) * 100

	data := struct {
		Blocks      []*Block
		TotalSupply int
	}{
		Blocks:      ws.blockchain.Blocks,
		TotalSupply: totalSupply,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := ws.templates.ExecuteTemplate(w, "explorer.html", data)
	if err != nil {
		fmt.Printf("❌ Error executing explorer.html: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ============================================
// QR Code Handler
// ============================================

// qrCodeHandler generates a QR code for a wallet address
func (ws *WebServer) qrCodeHandler(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "Address is required", http.StatusBadRequest)
		return
	}

	// Default size
	size := 256
	if sizeParam := r.URL.Query().Get("size"); sizeParam != "" {
		if s, err := strconv.Atoi(sizeParam); err == nil && s > 0 && s <= 1024 {
			size = s
		}
	}

	// Generate QR code with medium error correction
	png, err := qrcode.Encode(address, qrcode.Medium, size)
	if err != nil {
		http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		return
	}

	// Set proper content type
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 24 hours
	w.Write(png)
}

// ============================================
// API Handlers
// ============================================

// createWalletAPI creates a new wallet
func (ws *WebServer) createWalletAPI(w http.ResponseWriter, r *http.Request) {
	wallet := NewWallet()

	response := APIResponse{
		Success: true,
		Message: "Wallet created successfully",
		Data: map[string]string{
			"address":   wallet.GetAddressString(),
			"publicKey": fmt.Sprintf("%x", wallet.PublicKey),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getBalanceAPI gets the balance for an address
func (ws *WebServer) getBalanceAPI(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Message: "Address is required",
		})
		return
	}

	balance := ws.blockchain.GetBalance(address)

	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"address": address,
			"balance": balance,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// sendCoinsAPI sends coins from one address to another
func (ws *WebServer) sendCoinsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Message: "Method not allowed",
		})
		return
	}

	var req SendRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Message: "Invalid request",
		})
		return
	}

	wallet := &Wallet{PublicKey: []byte(req.From)}
	tx, err := ws.blockchain.SendCoins(wallet, req.To, req.Amount)
	if err != nil {
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Skip signing for API
	err = ws.blockchain.AddTransaction(tx)
	if err != nil {
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	response := APIResponse{
		Success: true,
		Message: "Transaction sent successfully",
		Data: map[string]interface{}{
			"txid": fmt.Sprintf("%x", tx.ID),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getBlocksAPI returns all blocks
func (ws *WebServer) getBlocksAPI(w http.ResponseWriter, r *http.Request) {
	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"blocks": ws.blockchain.Blocks,
			"count":  len(ws.blockchain.Blocks),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getTransactionAPI returns a specific transaction
func (ws *WebServer) getTransactionAPI(w http.ResponseWriter, r *http.Request) {
	txid := r.URL.Query().Get("txid")
	if txid == "" {
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Message: "Transaction ID is required",
		})
		return
	}

	for _, block := range ws.blockchain.Blocks {
		for _, tx := range block.Transactions {
			if fmt.Sprintf("%x", tx.ID) == txid {
				response := APIResponse{
					Success: true,
					Data: map[string]interface{}{
						"transaction": tx,
						"block":       block.Height,
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
				return
			}
		}
	}

	json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Message: "Transaction not found",
	})
}

// mineAPI mines a new block
func (ws *WebServer) mineAPI(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Message: "Address is required",
		})
		return
	}

	coinbase := NewCoinbaseTX(address, "Mining Reward")
	ws.blockchain.AddBlock([]*Transaction{coinbase})

	response := APIResponse{
		Success: true,
		Message: "Block mined successfully",
		Data: map[string]interface{}{
			"address": address,
			"reward":  100,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ============================================
// Helper Functions
// ============================================

// formatTimestamp formats a Unix timestamp to a readable string
func formatTimestamp(timestamp int64) string {
	if timestamp == 0 {
		return "Genesis Block"
	}
	t := time.Unix(timestamp, 0)
	return t.Format("2006-01-02 15:04:05")
}

// ============================================
// Types for the web server
// ============================================

// WalletPageData represents data for the wallet page
type WalletPageData struct {
	Address      string
	Balance      int
	Transactions []TransactionDisplay
}

// TransactionDisplay represents a transaction for display
type TransactionDisplay struct {
	TxID      string
	From      string
	To        string
	Amount    int
	Type      string
	Block     int
	Time      string
	Timestamp int64
}

// SendRequest represents a send coins request
type SendRequest struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int    `json:"amount"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
