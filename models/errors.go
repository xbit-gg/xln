package models

import "errors"

var (
	// auth
	MsgAuthNotFound     = "could not find ln auth record"
	MsgGetAuthFailed    = "failed to get ln auth record"
	MsgCreateAuthFailed = "failed to create ln auth record"
	MsgAuthFailed       = "failed to authenticate"

	// invoice
	MsgInvoiceNotFound                = "could not find invoice"
	MsgListInvoicesFailed             = "failed to list invoices"
	MsgGetInvoiceFailed               = "failed to get invoice"
	MsgCreateInvoiceFailed            = "failed to create invoice"
	MsgUpdateInvoiceSettledDateFailed = "failed to update invoice settlement date"
	MsgNullifyInvoiceSenderFailed     = "failed to set invoice sender to nil"
	MsgNullifyInvoiceRecipientFailed  = "failed to set invoice recipient to nil"
	MsgUpdateInvoiceSenderFailed      = "failed to update invoice sender"

	// pending invoice
	MsgPendingInvoiceNotFound          = "could not find pending invoice"
	MsgGetPendingInvoiceFailed         = "failed to get pending invoice"
	MsgListWalletPendingInvoicesFailed = "failed to list wallet's pending invoices"
	MsgListPendingInvoicesFailed       = "failed to list pending invoices"
	MsgCreatePendingInvoiceFailed      = "failed to create pending invoice"
	MsgDeletePendingInvoiceFailed      = "failed to delete pending invoice"

	// pending payment
	MsgPendingPaymentNotFound          = "could not find pending payment"
	MsgGetPendingPaymentFailed         = "failed to get pending payment"
	MsgListWalletPendingPaymentsFailed = "failed to list wallet's pending payments"
	MsgListPendingPaymentsFailed       = "failed to list pending payments"
	MsgCreatePendingPaymentFailed      = "failed to create pending payment"
	MsgDeletePendingPaymentFailed      = "failed to delete pending payment"

	// transaction
	MsgGetTransactionFailed              = "failed to get transaction"
	MsgTransactionNotFound               = "could not find transaction"
	MsgListWalletTransactionsFailed      = "failed to list transactions for wallet"
	MsgCreateTransactionFailed           = "failed to create transaction"
	MsgCreateTransactionsFailed          = "failed to create transactions"
	MsgListUserTransactionsFailed        = "failed to list user's transactions"
	MsgNullifyTransactionSenderFailed    = "failed to set transaction sender to nil"
	MsgNullifyTranscationRecipientFailed = "failed to set transaction recipient to nil"

	// user
	MsgUserNotFound            = "could not find user"
	MsgCreateUserFailed        = "failed to create user"
	MsgGetUserFailed           = "failed to get user"
	MsgDeleteUserFailed        = "failed to delete user"
	MsgGetUserWithApiKeyFailed = "failed to get a user with specific apiKey"
	MsgUpdateUserFailed        = "failed to update user"
	MsgLinkedUserNotFound      = "could not find user linked to a ln wallet"
	MsgGetLinkedUserFailed     = "failed to get user that is linked to ln wallet"
	MsgDuplicateUsername       = "user with that username already exists"
	MsgUserExistsFailed        = "failed to determine if user exists"

	// wallet
	MsgCreateWalletFailed                 = "failed to create wallet"
	MsgWalletIDAlreadyExists              = "wallet with that id already exists"
	MsgWalletDeleteFailed                 = "failed to delete wallet"
	MsgMultipleWalletDeleteFailed         = "failed to delete multiple wallets"
	MsgGetWalletFailed                    = "failed to get wallet"
	MsgWalletNotFound                     = "could not find wallet"
	MsgListWalletFailed                   = "failed to list wallets"
	MsgListLinkedWalletFailed             = "failed to list linked wallets"
	MsgDeterminingIfWalletsFromUserFailed = "failed to determine if wallets are from user"
	MsgLockWalletRecordFailed             = "failed to lock wallet record"
	MsgGetBalanceFailed                   = "failed to get balance of wallet"
	MsgDecrementWalletFailed              = "failed to decrement balance of wallet"
	MsgIncrementWalletFailed              = "failed to increment balance of wallet"
	MsgUpdateWalletBalanceFailed          = "failed to update balance of wallet"
	MsgUpdateWalletFailed                 = "failed to update wallet"
	MsgGetWalletWithApiKeyFailed          = "failed to get wallet with specific apiKey"
	MsgLinkedWalletNotFound               = "could not find record of wallet linked to a ln wallet"
	MsgGetLinkedWalletFailed              = "failed to get wallet that is linked to ln wallet"
	MsgGetConfirmedBalanceFailed          = "failed to confirm the current wallet balance"

	// Withdraw
	MsgCreateWithdrawFailed             = "failed to create withdraw"
	MsgGetWithdrawFailed                = "failed to get withdraw"
	MsgWithdrawNotFound                 = "could not find withdraw"
	MsgListWalletPendingWithdrawsFailed = "failed to list wallet's pending withdraws"

	// misc
	MsgReceivedNil                      = "expected to receive record but received nil instead"
	MsgCannotHaveLabelForNilValue       = "cannot assign a label to a nil value"
	MsgIDAlreadyInUse                   = "ID is already in use"
	gormMsgSubstrUniqueConstraintFailed = "UNIQUE constraint failed"                       // e.g. UNIQUE constraint failed: wallets.id, wallets.username
	gormMsgSubstrDuplicateKey           = "duplicate key value violates unique constraint" // e.g. duplicate key value violates unique constraint \"wallets_pkey\"

	// errors
	ErrInternal                       = errors.New("internal database error")
	ErrInvalidParams                  = errors.New("invalid parameters")
	ErrCannotUpdateLockedWallet       = errors.New("cannot update locked wallet") // user cannot update a locked wallet
	ErrCannotTransactWithLockedWallet = errors.New("cannot transact with locked wallet")
	ErrLockedWallet                   = errors.New("wallet locked")
	ErrMaxNumResultsExceeded          = errors.New("number of results exceed the maximum allowed to be returned")
	ErrInsufficientBalance            = errors.New("wallet has insufficient balance")
	ErrNonZeroBalance                 = errors.New("wallet has non-zero balance")
	ErrInvoiceNotFound                = errors.New(MsgInvoiceNotFound)
	ErrUserNotFound                   = errors.New(MsgUserNotFound)
	ErrWalletNotFound                 = errors.New(MsgWalletNotFound)
	ErrPendingPaymentNotFound         = errors.New(MsgPendingPaymentNotFound)
	ErrPendingInvoiceNotFound         = errors.New(MsgPendingInvoiceNotFound)
	ErrTransactionNotFound            = errors.New(MsgTransactionNotFound)
	ErrAuthNotFound                   = errors.New(MsgAuthNotFound)
	ErrLinkedUserNotFound             = errors.New(MsgLinkedUserNotFound)
	ErrLinkedWalletNotFound           = errors.New(MsgLinkedWalletNotFound)
	ErrWithdrawNotFound               = errors.New(MsgWithdrawNotFound)
)
