package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func setupTestAccounts() {
	fmt.Println("Setting up test accounts")
	accounts = map[string]*Account{
		"Mark": {Name: "Mark", Balance: 100, Mutex: sync.Mutex{}},
		"Jane": {Name: "Jane", Balance: 50, Mutex: sync.Mutex{}},
		"Adam": {Name: "Adam", Balance: 0, Mutex: sync.Mutex{}},
	}
	fmt.Println("Test accounts initialized")
}

func TestTransferMoney_Success(t *testing.T) {
	setupTestAccounts()
	fmt.Println("Starting TestTransferMoney_Success")

	reqBody := TransferRequest{
		From:   "Mark",
		To:     "Jane",
		Amount: 30,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/transfer", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(transferMoney)
	handler.ServeHTTP(rr, req)

	fmt.Printf("Response code: %d, Body: %s", rr.Code, rr.Body.String())

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	expectedSenderBalance := 70
	expectedReceiverBalance := 80

	if accounts["Mark"].Balance != expectedSenderBalance {
		t.Errorf("Expected Mark's balance to be %d, got %d", expectedSenderBalance, accounts["Mark"].Balance)
	}

	if accounts["Jane"].Balance != expectedReceiverBalance {
		t.Errorf("Expected Jane's balance to be %d, got %d", expectedReceiverBalance, accounts["Jane"].Balance)
	}

	fmt.Println("TestTransferMoney_Success completed successfully")
}

func TestTransferMoney_InsufficientFunds(t *testing.T) {
	setupTestAccounts()
	fmt.Println("Starting TestTransferMoney_InsufficientFunds")

	reqBody := TransferRequest{
		From:   "Jane",
		To:     "Mark",
		Amount: 100, // More than Jane has
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/transfer", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(transferMoney)
	handler.ServeHTTP(rr, req)

	fmt.Printf("Response code: %d, Body: %s", rr.Code, rr.Body.String())

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}

	fmt.Println("TestTransferMoney_InsufficientFunds completed successfully")
}

func TestTransferMoney_SameAccount(t *testing.T) {
	setupTestAccounts()
	fmt.Println("Starting TestTransferMoney_SameAccount")

	reqBody := TransferRequest{
		From:   "Mark",
		To:     "Mark",
		Amount: 10,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/transfer", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(transferMoney)
	handler.ServeHTTP(rr, req)

	fmt.Printf("Response code: %d, Body: %s", rr.Code, rr.Body.String())

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}

	fmt.Println("TestTransferMoney_SameAccount completed successfully")
}

func TestGetBalance(t *testing.T) {
	setupTestAccounts()
	fmt.Println("Starting TestGetBalance")

	req := httptest.NewRequest(http.MethodGet, "/balance?name=Mark", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getBalance)
	handler.ServeHTTP(rr, req)

	fmt.Printf("Response code: %d, Body: %s", rr.Code, rr.Body.String())

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var resp map[string]int
	json.Unmarshal(rr.Body.Bytes(), &resp)

	expectedBalance := 100
	if resp["balance"] != expectedBalance {
		t.Errorf("Expected balance to be %d, got %d", expectedBalance, resp["balance"])
	}

	fmt.Println("TestGetBalance completed successfully")
}
