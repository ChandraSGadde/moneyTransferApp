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
	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.From == req.To {
		http.Error(w, "Cannot transfer to the same account", http.StatusBadRequest)
		return
	}

	globalMutex.Lock()
	sender, senderExists := accounts[req.From]
	receiver, receiverExists := accounts[req.To]

	if !senderExists || !receiverExists {
		globalMutex.Unlock()
		http.Error(w, "Invalid account(s)", http.StatusBadRequest)
		return
	}

	// Lock both sender and receiver to avoid race conditions
	sender.Mutex.Lock()
	receiver.Mutex.Lock()

defer sender.Mutex.Unlock()

defer receiver.Mutex.Unlock()

	if sender.Balance < req.Amount {
		globalMutex.Unlock()
		http.Error(w, "Insufficient funds", http.StatusBadRequest)
		return
	}

	sender.Balance -= req.Amount
	receiver.Balance += req.Amount

	globalMutex.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func getBalance(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	globalMutex.Lock()
	account, exists := accounts[name]
	globalMutex.Unlock()

	if !exists {
		http.Error(w, "Account not found", http.StatusBadRequest)
		return
	}

	account.Mutex.Lock()
	defer account.Mutex.Unlock()

	json.NewEncoder(w).Encode(map[string]int{"balance": account.Balance})
}

func main() {
	http.HandleFunc("/transfer", transferMoney)
	http.HandleFunc("/balance", getBalance)

	fmt.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

