package main

import "time"

type Auth struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expiresAt"`
}

type User struct {
	Email string `json:"email"`
	Pwd   string `json:"pwd"`
	Name  string `json:"name"`
}

type Bank struct {
	OwningUserId  string `json:"owningUserId"`
	BankId        string `json:"bankId"`
	BankName      string `json:"bankName"`
	AccountNumber string `json:"accountNumber"`
}

type BankAccount struct {
	BankId         string  `json:"bankId"`
	AccountId      string  `json:"accountId"`
	AccountName    string  `json:"accountName"`
	AccountType    string  `json:"accountType"`
	Last4          string  `json:"last4"`
	CurrentBalance float64 `json:"currentBalance"`
}

type Card struct {
	AccountId   string `json:"accountId"`
	CardId      string `json:"cardId"`
	Last4       string `json:"last4"`
	ExpiryMonth string `json:"expiryMonth"`
	ExpiryYear  string `json:"expiryYear"`
	CVV         string `json:"cvv"`
	Active      bool   `json:"active"`
}

type Transaction struct {
	AccountId       string    `json:"accountId"`
	TransactionId   string    `json:"transactionId"`
	TransactionDate time.Time `json:"transactionDate"`
	Amount          float64   `json:"amount"`
	TransactionType string    `json:"transactionType"`
	Description     string    `json:"description"`
	CardId          *string   `json:"cardId"`
}
