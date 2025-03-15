package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type Account struct {
	Name    string
	Balance int
	Mutex   sync.Mutex
}

type TransferRequest struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int    `json:"amount"`
}

var accounts = map[string]*Account{
	"Mark": {Name: "Mark", Balance: 100},
	"Jane": {Name: "Jane", Balance: 50},
	"Adam": {Name: "Adam", Balance: 0},
}

var globalMutex sync.Mutex

func transferMoney(w http.ResponseWriter, r *http.Request) {
	log.Println("Received transfer request")
	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Error decoding request:", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("Processing transfer: %s -> %s, Amount: %d\n", req.From, req.To, req.Amount)

	if req.From == req.To {
		log.Println("Transfer to same account is not allowed")
		http.Error(w, "Cannot transfer to the same account", http.StatusBadRequest)
		return
	}

	globalMutex.Lock()
	log.Println("Global mutex locked")
	sender, senderExists := accounts[req.From]
	receiver, receiverExists := accounts[req.To]

	if !senderExists || !receiverExists {
		globalMutex.Unlock()
		log.Println("Invalid account(s) in request")
		http.Error(w, "Invalid account(s)", http.StatusBadRequest)
		return
	}

	sender.Mutex.Lock()
	receiver.Mutex.Lock()

	defer sender.Mutex.Unlock()
	defer receiver.Mutex.Unlock()

	if sender.Balance < req.Amount {
		globalMutex.Unlock()
		log.Println("Insufficient funds for transfer")
		http.Error(w, "Insufficient funds", http.StatusBadRequest)
		return
	}

	sender.Balance -= req.Amount
	receiver.Balance += req.Amount
	log.Printf("Transfer successful: %s -> %s, New Balances: %s=%d, %s=%d\n", req.From, req.To, req.From, sender.Balance, req.To, receiver.Balance)

	globalMutex.Unlock()
	log.Println("Global mutex unlocked")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func getBalance(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	log.Printf("Fetching balance for account: %s\n", name)

	globalMutex.Lock()
	account, exists := accounts[name]
	globalMutex.Unlock()

	if !exists {
		log.Println("Account not found")
		http.Error(w, "Account not found", http.StatusBadRequest)
		return
	}

	account.Mutex.Lock()
	defer account.Mutex.Unlock()

	log.Printf("Balance fetched: %s=%d\n", name, account.Balance)
	json.NewEncoder(w).Encode(map[string]int{"balance": account.Balance})
}

func main() {
	log.Println("Starting server on port 8080...")
	http.HandleFunc("/transfer", transferMoney)
	http.HandleFunc("/balance", getBalance)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
