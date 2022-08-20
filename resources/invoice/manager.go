package invoice

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"github.com/xbit-gg/xln/db"
	"github.com/xbit-gg/xln/lnd"
	"github.com/xbit-gg/xln/models"
	"github.com/xbit-gg/xln/resources/wallet"
	"gorm.io/gorm"
)

// Any more and you should just use VISA
var feePercentLimit = 0.03

type Manager interface {
	CreateInvoice(username, walletId, memo string, value int64, expiry int64) (*models.Invoice, error)
	ListWalletInvoices(username, walletId string) ([]*models.Invoice, error)
	GetWalletInvoice(username, walletId, paymentHash string) (*models.Invoice, error)
	GetInvoice(paymentHash string) (*models.Invoice, error)
	PayInvoice(username, walletId, pr string, sync bool) (*Payment, error)
	PayWithdrawInvoice(k1, pr string) error
	PayInvoiceAmount(username, walletId, pr string, sync bool, amount int64) (*Payment, error)
}

type manager struct {
	lnClient             lnrpc.LightningClient
	routerClient         routerrpc.RouterClient
	wallets              wallet.Manager
	pendingInvoiceCache  *cache.Cache
	pendingPaymentCache  *cache.Cache
	pendingWithdrawCache *cache.Cache
	muWdraw              sync.Mutex
	db                   *db.DB
	maxPayment           int64
}

func NewManager(lndClient *lnd.Client, walletManager wallet.Manager, db *db.DB, maxPayment int64) Manager {
	m := &manager{
		lnClient:            lnrpc.NewLightningClient(lndClient.Conn),
		routerClient:        routerrpc.NewRouterClient(lndClient.Conn),
		wallets:             walletManager,
		pendingInvoiceCache: cache.New(time.Hour, 6*time.Hour),
		pendingPaymentCache: cache.New(time.Hour, 6*time.Hour),
		db:                  db,
		maxPayment:          maxPayment,
	}
	m.handleStalePayments()
	m.handleStaleInvoices()
	go m.trackPendingInvoices()

	return m
}

type Payment struct {
	Success       bool
	AmountMsat    uint64
	FeeMsat       uint64
	FailureReason string
}

// CreateInvoice creates an LN invoice payable to wallet with a given value in millisatoshis.
func (m *manager) CreateInvoice(username string, walletId string, memo string, value int64, expiry int64) (*models.Invoice, error) {
	if value > m.maxPayment {
		log.WithField("value", value).Warn("CreateInvoice called with too large a value")
		return nil, fmt.Errorf("invoice of size %d msat is greater than the maximum payment size", value)
	}
	wallet, err := m.getAndValidateWallet(username, walletId)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user":   username,
			"wallet": walletId,
		}).Warn("CreateInvoice failed")
		return nil, err
	}

	inv := &lnrpc.Invoice{
		Memo:      memo,
		ValueMsat: value,
		Expiry:    expiry,
	}
	invoice, err := m.lnClient.AddInvoice(context.Background(), inv)
	if err != nil {
		log.WithError(err).WithField("invoice", inv).Warn("Unable to create invoice")
		return nil, err
	}

	decodePayRes, err := m.lnClient.DecodePayReq(context.Background(), &lnrpc.PayReqString{PayReq: invoice.PaymentRequest})
	if err != nil {
		log.WithError(err).Error("Unable to decode pay response")
		return nil, err
	}

	paymentHash := base64.StdEncoding.EncodeToString(invoice.RHash)
	pendingInv := &models.PendingInvoice{
		WalletID:       walletId,
		WalletUsername: username,
		PaymentHash:    paymentHash,
		Amount:         uint64(value),
	}
	err = m.db.Repo.CreatePendingInvoice(m.db.DB, pendingInv)
	if err != nil {
		log.WithError(err).Error("Failed to add pending invoice to DB")
		return nil, err
	}
	m.pendingInvoiceCache.SetDefault(paymentHash, pendingInv)

	xlnInvoice := models.Invoice{
		Amount:            uint64(value),
		Memo:              memo,
		PaymentHash:       paymentHash,
		PaymentRequest:    invoice.PaymentRequest,
		Pubkey:            decodePayRes.Destination,
		RecipientID:       &wallet.ID,
		RecipientUsername: &wallet.Username,
		Settled:           time.Time{},
		Timestamp:         time.Now().UTC(),
	}
	err = m.db.Repo.CreateInvoice(m.db.DB, &xlnInvoice)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"wallet":       walletId,
			"payment_hash": paymentHash,
		}).Error("Unable to add invoice to the database")
		return nil, err
	}
	return &xlnInvoice, nil
}

func (m *manager) ListWalletInvoices(username, walletId string) ([]*models.Invoice, error) {
	return m.db.Repo.ListWalletInvoices(m.db.DB, username, walletId)
}

func (m *manager) GetWalletInvoice(username, walletId, paymentHash string) (*models.Invoice, error) {
	return m.db.Repo.GetWalletInvoice(m.db.DB, username, walletId, paymentHash)
}

func (m *manager) GetInvoice(paymentHash string) (*models.Invoice, error) {
	return m.db.Repo.GetInvoice(m.db.DB, paymentHash)
}

// PayInvoice uses wallet to pay the invoice specified by paymentRequest.
func (m *manager) PayInvoice(username string, walletId string, pr string, sync bool) (*Payment, error) {
	wallet, err := m.getAndValidateWallet(username, walletId)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user":   username,
			"wallet": walletId,
			"pr":     pr,
		}).Warn("PayInvoice failed")
		return nil, err
	}
	payreq, err := m.lnClient.DecodePayReq(context.Background(), &lnrpc.PayReqString{PayReq: pr})
	if err != nil {
		log.WithError(err).WithField("pr", pr).Warn("PayInvoice called with invalid payment request format")
		return nil, errors.New("invalid payment request format")
	}
	if payreq.NumMsat == 0 {
		return nil, errors.New("amount must be specified when paying a zero amount invoice")
	}
	return m.payInvoice(wallet, pr, payreq, payreq.NumMsat, sync)
}

func (m *manager) PayWithdrawInvoice(k1, pr string) error {
	withdrawal, err := m.db.Repo.GetWithdraw(m.db.DB, k1, false)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"withdraw": k1,
			"pr":       pr,
		}).Error("failed to pay withdrawal")
		return err
	}
	wallet, err := m.getAndValidateWallet(withdrawal.Username, withdrawal.WalletID)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"username": withdrawal.Username,
			"wallet":   withdrawal.WalletID,
			"withdraw": k1,
			"pr":       pr,
		}).Warn("failed to pay withdrawal")
		return err
	}

	if withdrawal.MaxUse != 0 {
		if withdrawal.Uses >= withdrawal.MaxUse {
			return fmt.Errorf("exceeded maximum number of allowed withdrawals")
		}
		pWithdrawals, err := m.db.Repo.ListPendingWithdraws(m.db.DB, withdrawal.Username, withdrawal.WalletID, k1)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"username": withdrawal.Username,
				"wallet":   withdrawal.WalletID,
				"withdraw": k1,
				"pr":       pr,
			}).Error("failed to pay withdrawal")
			return err
		} else if withdrawal.Uses+uint(len(pWithdrawals)) >= withdrawal.MaxUse {
			return fmt.Errorf("processing other withdrawals. Retry.")
		}

	}

	payreq, err := m.lnClient.DecodePayReq(context.Background(), &lnrpc.PayReqString{PayReq: pr})
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"pr":       pr,
			"username": withdrawal.Username,
			"wallet":   withdrawal.WalletID,
			"withdraw": k1,
		}).Warn("PayWithdrawInvoice called with invalid payment request format")
		return errors.New("invalid payment request format")
	}
	if payreq.NumMsat == 0 {
		return errors.New("amount must be specified when paying a zero amount invoice")
	}

	if int64(withdrawal.MinMsat) > payreq.NumMsat {
		return fmt.Errorf("amount is below the minimum withdrawable value")
	} else if withdrawal.MaxMsat != 0 && payreq.NumMsat > int64(withdrawal.MaxMsat) {
		return fmt.Errorf("amount exceeds the maximum withdrawable value")
	} else {
		go func() {
			_, err := m.payInvoice(wallet, pr, payreq, payreq.NumMsat, true)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"user":     withdrawal.Username,
					"wallet":   withdrawal.WalletID,
					"withdraw": k1,
					"pr":       pr,
				}).Error("Failed to withdraw sats")
				return
			}
			m.db.Repo.IncrementWithdrawCount(m.db.DB, k1)
		}()
		return nil
	}
}

func (m *manager) PayInvoiceAmount(username string, walletId string, pr string, sync bool, amount int64) (*Payment, error) {
	wallet, err := m.getAndValidateWallet(username, walletId)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"username": username,
			"walletId": walletId,
			"pr":       pr,
		}).Warn("PayInvoiceAmount failed")
		return nil, err
	}
	payreq, err := m.lnClient.DecodePayReq(context.Background(), &lnrpc.PayReqString{PayReq: pr})
	if err != nil {
		log.WithError(err).WithField("pr", pr).Warn("PayInvoiceAmount called with invalid payment request format")
		return nil, errors.New("invalid payment request format")
	}
	if payreq.NumMsat != 0 && amount != payreq.NumMsat {
		return nil, fmt.Errorf("provided amount %d does not satisfy payment request amount %d", amount, payreq.NumMsat)
	}
	return m.payInvoice(wallet, pr, payreq, amount, sync)
}

func (m *manager) payInvoice(wal *models.Wallet, pr string, payreq *lnrpc.PayReq, amount int64, sync bool) (*Payment, error) {
	if payreq.NumMsat > m.maxPayment {
		log.WithField("value", payreq.NumMsat).Warn("payInvoice called with too large a value")
		return nil, fmt.Errorf("size %d msat is greater than the maximum payment size", payreq.NumMsat)
	}

	// if pending invoice in db then it is self-payment
	hexPayH, err := hex.DecodeString(payreq.PaymentHash)
	if err != nil {
		log.WithError(err).Error("failed to decode payment request's payment hash")
		return nil, err
	}
	payH := base64.StdEncoding.EncodeToString(hexPayH)
	if pending, err := m.db.Repo.GetPendingInvoice(m.db.DB, payH); err == nil {
		if err := m.handleSelfPayments(wal.Username, wal.ID, pending.WalletUsername, pending.WalletID, payH, amount); err == nil {
			return &Payment{
				Success:    true,
				AmountMsat: uint64(amount),
			}, nil
		} else {
			return nil, err
		}
	} else if err != models.ErrPendingInvoiceNotFound {
		return nil, err
	}
	log.WithFields(log.Fields{
		"user":   wal.Username,
		"wallet": wal.ID,
		"pr":     pr,
	}).Info("processing external payment")

	var payment *lnrpc.Payment
	err = m.db.Transaction(func(tx *gorm.DB) error {
		cBal, err := m.db.Repo.GetConfirmedBalance(tx, wal.Username, wal.ID)
		if err != nil {
			return err
		}
		if cBal < uint64(amount) {
			log.WithFields(log.Fields{
				"user":   wal.Username,
				"wallet": wal.ID,
				"pr":     pr,
			}).Warn("Wallet attempted payment with insufficient funds")
			return errors.New("insufficient funds")
		}

		// Clear amount when paying an invoice with specified amount
		specifiedAmount := amount
		if payreq.NumMsat > 0 {
			specifiedAmount = 0
		}
		req := &routerrpc.SendPaymentRequest{
			PaymentRequest: pr,
			FeeLimitMsat: int64(math.Min(
				float64(payreq.NumMsat)*feePercentLimit,
				float64(cBal-uint64(payreq.NumMsat)))),
			TimeoutSeconds: 30,
			AmtMsat:        specifiedAmount,
		}
		stream, err := m.routerClient.SendPaymentV2(context.Background(), req)
		if err != nil {
			log.WithError(err).Error("Error calling SendPaymentV2")
			return fmt.Errorf("error sending payment: %v", err)
		}

		payment, err = stream.Recv()
		if err != nil {
			log.WithError(err).WithField("wallet", wal.ID).Warn("Failed to pay an invoice")
			return err
		}
		pendingPayment := &models.PendingPayment{
			WalletID:       wal.ID,
			WalletUsername: wal.Username,
			PaymentHash:    payment.PaymentHash,
			Amount:         uint64(amount),
		}
		err = m.db.Repo.CreatePendingPayment(tx, pendingPayment)
		if err != nil {
			log.WithError(err).Error("Failed to add pending payment to DB")
			return err
		}
		m.pendingPaymentCache.SetDefault(payment.PaymentHash, pendingPayment)
		invoice := models.Invoice{
			Amount:         uint64(amount),
			PaymentHash:    payment.PaymentHash,
			PaymentRequest: pr,
			Pubkey:         payreq.Destination,
			SenderID:       &wal.ID,
			SenderUsername: &wal.Username,
			Settled:        time.Time{},
			Timestamp:      time.Now().UTC(),
		}
		if err := m.db.Repo.CreateInvoice(tx, &invoice); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"paymentRequest": pr,
			"destination":    payreq.Destination,
			"user":           wal.Username,
			"wallet":         wal.ID,
		}).Error("failed to finalize payment")
		return nil, err
	}
	if sync {
		return m.trackPendingPayment(payment.PaymentHash), nil
	} else {
		go m.trackPendingPayment(payment.PaymentHash)
	}

	return nil, nil
}

// validateWallet throws an error if the wallet, which exists, is not valid.
// e.g. it is locked
func (m *manager) getAndValidateWallet(username, walletId string) (*models.Wallet, error) {
	wallet, err := m.wallets.GetWallet(username, walletId)
	if err != nil {
		return nil, err
	}
	if wallet.Locked {
		log.WithFields(log.Fields{
			"wallet": wallet.ID,
			"user":   wallet.Username,
		}).Info("cannot transact with locked wallet")
		return nil, errors.New("cannot transact with locked wallet")
	}
	return wallet, nil
}
