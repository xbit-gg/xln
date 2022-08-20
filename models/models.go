package models

import (
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	// Wallet methods

	// CreateWallet adds the wallet to the database
	// Errors if the database action fails
	CreateWallet(tx *gorm.DB, wallet *Wallet) error

	// DeleteWalletZeroBalance removes wallet record from database
	// Errors if the database action fails
	DeleteWalletZeroBalance(tx *gorm.DB, username, walletId string) error

	// DeleteWallet removes wallet record from database
	// Errors if the database action fails
	DeleteWallet(tx *gorm.DB, username, walletId string) error

	// DeleteUserWallets removes all wallet record from database belonging to user
	// Errors if the database action fails
	DeleteUserWallets(tx *gorm.DB, username string) error

	// UpdateWalletOptions updates wallet with the wallet options provided
	// Errors if three is no matching record or if the database action fails
	UpdateWalletOptions(tx *gorm.DB, username, walletId string, walletOptions *WalletOptions) error

	// GetWallet returns wallet
	// Errors if the database action fails or if record not found
	GetWallet(tx *gorm.DB, username, walletId string) (*Wallet, error)

	// ListUserWallets returns all the wallets associated with a user
	// Errors if the database action fails
	ListUserWallets(tx *gorm.DB, username string) ([]*Wallet, error)

	// UpdateWalletWithBalance updates the wallet with the balance
	// Errors if the database action fails
	UpdateWalletWithBalance(tx *gorm.DB, username, walletId string, newBalance uint64) (*Wallet, error)

	// IncrementWalletBalance increments the wallet with the balance in a single atomic transaction
	// Errors if the database action fails
	IncrementWalletBalance(tx *gorm.DB, username, walletId string, deltaBalance uint64) error

	// ** DO NOT USE ** on a nonexistent wallet for consistent log messages. If
	// used on a nonexistent wallet, then if it cannot find the wallet (with sufficient balance),
	// it will return error of insufficient balance.
	// DecrementWalletBalance decrements the wallet with the balance in a single atomic transaction
	// if and only if the balance is above 0
	// Errors if the database action fails, insufficient balance, locked rows
	DecrementWalletBalance(tx *gorm.DB, username, walletId string, deltaBalance uint64) error

	// WalletsFromUser determines whether all wallets belong a user.
	// The first return item is true if they both belong to the user, and false if they do not or if it is undetermined.
	// The second return item is nil if the database action fails, and otherwise returns both wallets.
	// The third return item is nil if there is no database error. Otherwise, it returns the database error.
	WalletsFromUser(tx *gorm.DB, username string, walletIds []string) (bool, []Wallet, error)

	// LockWalletRecordForUpdate locks the row with FOR UPDATE strength and returns the wallet if it exists, and an error
	// if it does not exist or if the wallet is locked.
	// Is not user-facing.
	LockWalletRecordForUpdate(tx *gorm.DB, username, walletId string) (*Wallet, error)

	// GetWalletWithApiKey retrieves a single wallet with apiKey
	// errors if there is no such wallet, or if there is an internal db error
	GetWalletWithApiKey(tx *gorm.DB, apiKey string) (*Wallet, error)

	// UpdateWalletLink records that a wallet has been linked successfully using a key if the key is not nil.
	// If key is nil then the wallet is unlinked.
	// errors if there is no such wallet, or if there is an internal db error
	UpdateWalletLink(tx *gorm.DB, username, walletId string, key, label *string) error

	GetLinkedWallet(tx *gorm.DB, key string) (*Wallet, error)

	ListUserLinkedWallets(tx *gorm.DB, username string) ([]*Wallet, error)

	// GetConfirmedBalance returns the wallet balance, less any pending payments
	GetConfirmedBalance(tx *gorm.DB, username, walletId string) (uint64, error)

	// User methods

	// CreateUser adds the user to the database
	// Returns an error if the database action fails. Otherwise returns nil.
	CreateUser(tx *gorm.DB, user *User) error

	// DeleteUser deletes a user
	DeleteUser(tx *gorm.DB, username string) error

	// GetUser returns a user if it exists
	GetUser(tx *gorm.DB, username string) (*User, error)

	// UserExists checks if user exists
	UserExists(tx *gorm.DB, username string) (bool, error)

	// GetUserWithApiKey retrieves a single user with apiKey
	// errors if there is no such user, or if there is an internal db error
	GetUserWithApiKey(tx *gorm.DB, apiKey string) (*User, error)

	GetLinkedUser(tx *gorm.DB, key string) (*User, error)

	UpdateUserLink(tx *gorm.DB, username string, k1, label *string) error

	// Pending invoice methods

	// CreatePendingInvoice adds the PendingInvoice to the database
	// Errors if the database action fails
	CreatePendingInvoice(tx *gorm.DB, pendingInv *PendingInvoice) error

	// DeletePendingInvoice removes the PendingInvoice record from the database
	// Errors if the database action fails
	DeletePendingInvoice(tx *gorm.DB, paymentHash string) error

	// ListWalletPendingInvoices gets pending invoice records for a wallet
	// Errors if the database action fails
	ListWalletPendingInvoices(tx *gorm.DB, username, walletId string) ([]*PendingInvoice, error)

	// ListPendingInvoices gets all pending invoice records
	// Errors if the database action fails
	ListPendingInvoices(tx *gorm.DB) ([]*PendingInvoice, error)

	// GetPendingInvoice returns the invoice that matches the paymentHash
	// Errors if the database action fails
	GetPendingInvoice(tx *gorm.DB, paymentHash string) (*PendingInvoice, error)

	// Pending payments methods

	// CreatePendingPayment adds the PendingPayment to the database.
	// Errors if the database action fails or if record not found
	CreatePendingPayment(tx *gorm.DB, pendingPayment *PendingPayment) error

	// DeletePendingPayment removes the PendingPayment record from the database.
	// Errors if the database action fails or if record not found
	DeletePendingPayment(tx *gorm.DB, pendingPayment *PendingPayment) error

	// ListPendingPayments gets all pending payment records
	// Errors if the database action fails
	ListPendingPayments(tx *gorm.DB) ([]*PendingPayment, error)

	// ListWalletPendingPayments lists pending payments for a wallet
	// Errors if the database action fails
	ListWalletPendingPayments(tx *gorm.DB, username, walletId string) ([]*PendingPayment, error)

	// ListWithdrawalsPendingPayments lists pending withdraws
	// Errors if the database action fails
	ListPendingWithdraws(tx *gorm.DB, username, walletId, k1 string) ([]*PendingPayment, error)

	// GetPendingPayment returns the pending payment
	// Errors if the database action fails or if record not found
	GetPendingPayment(tx *gorm.DB, paymentHash string) (*PendingPayment, error)

	// Transaction methods

	// CreateTransaction adds a transactions to the database
	// Errors if the database action fails
	CreateTransaction(tx *gorm.DB, txs *Transaction) error

	// CreateTransactions adds the set of transactions to the database
	// Errors if the database action fails
	CreateTransactions(tx *gorm.DB, txs *[]Transaction) error

	// GetTransaction returns the transaction matching the ID
	// Errors if the database action fails or if record not found
	GetTransaction(tx *gorm.DB, transactionId string) (*Transaction, error)

	// GetWalletTransaction returns the transaction that matches ID and belongs to wallet
	// Errors if the database action fails or if record not found
	GetWalletTransaction(tx *gorm.DB, username, walletId string, transactionId string) (*Transaction, error)

	// ListWalletTransactions lists the transactions for a wallet with walletId within the
	// provided time range. If offset is nonzero, then the search results will skip that amount;
	// if limit is nonzero, then the number of results will be at most equal to the limit;
	// if descending is false then results will be returned chronollogically by creation time.
	// Returns: the corresponding list of transactions; the next index of the search result sequence;
	// the total number of records that belong to the user in the given time frame.
	ListWalletTransactions(tx *gorm.DB, username, walletId string, startTime, endTime time.Time,
		offset, limit uint, descending bool) (txns []*Transaction, nextOffset int, total uint64, err error)

	// ListUserTransactions lists the transactions for a user with username within the
	// provided time range. If offset is nonzero, then the search results will skip that amount;
	// if limit is nonzero, then the number of results will be at most equal to the limit;
	// if descending is fakse then results will be returned chronollogically by creation time.
	// Returns: the corresponding list of transactions; the next index of the search result sequence;
	// the total number of records that belong to the user in the given time frame.
	ListUserTransactions(tx *gorm.DB, username string, startTime, endTime time.Time,
		offset, limit uint, descending bool) (txns []*Transaction, nextOffset int, total uint64, err error)

	// NullifyTransactionSender sets transaction sender to nil
	NullifyTransactionSender(tx *gorm.DB, username, id string) error

	// NullifyTransactionRecipient sets transaction recipient to nil
	NullifyTransactionRecipient(tx *gorm.DB, username, id string) error

	// Invoice methods

	// CreateInvoice creates an invoice with a unique payment hash.
	// Errors if the database action fails.
	CreateInvoice(tx *gorm.DB, invoice *Invoice) error

	// ListWalletInvoices lists all the invoices associated with the wallet
	ListWalletInvoices(tx *gorm.DB, username, walletId string) ([]*Invoice, error)

	// GetInvoice retrieves an invoice from the DB that matches the paymentHash.
	// Errors if there does not exist an invoice with the paymentHash or if the db action fails.
	GetInvoice(tx *gorm.DB, paymentHash string) (*Invoice, error)

	// GetWalletInvoice gets an invoice if it is associated with the wallet
	GetWalletInvoice(tx *gorm.DB, username, walletId, paymentHash string) (*Invoice, error)

	// UpdateInvoiceSettleTime updates the settled timestamp of record with given paymentHash
	UpdateInvoiceSettleTime(tx *gorm.DB, paymentHash string, time *time.Time) error

	// NullifyInvoiceRecipient removes all references to user or wallet as a recipient.
	NullifyInvoiceRecipient(tx *gorm.DB, username, id string) error

	// NullifyInvoiceSender removes all references to user or wallet as a sender.
	NullifyInvoiceSender(tx *gorm.DB, username, id string) error

	// SetInvoiceSenderAmount sets the sender and amount on an existing invoice
	SetInvoiceSenderAmount(tx *gorm.DB, paymentHash, username, id string, amount int64) error

	// Auth methods

	// CreateAuth
	CreateAuth(tx *gorm.DB, auth *Auth) error

	// Authenticate
	Authenticate(tx *gorm.DB, k1 string) error

	// GetAuth
	GetAuth(tx *gorm.DB, k1 string) (*Auth, error)

	// Withdraw methods

	// CreateWithdraw creates a record of a reusable withdraw template
	CreateWithdraw(tx *gorm.DB, withdraw *Withdraw) (*Withdraw, error)

	// GetWalletWithdraw retrieves a withdrawal record belonging to a wallet
	GetWalletWithdraw(tx *gorm.DB, username, walletId, k1 string, includeExpired bool) (*Withdraw, error)

	// GetWithdraw retrieves a withdrawal record
	GetWithdraw(tx *gorm.DB, k1 string, includeExpired bool) (*Withdraw, error)

	// IncrementWithdrawCount
	IncrementWithdrawCount(tx *gorm.DB, k1 string) error
}
type repository struct {
}

func NewRepository() Repository {
	return &repository{}
}
