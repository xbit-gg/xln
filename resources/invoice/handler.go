package invoice

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	log "github.com/sirupsen/logrus"
	"github.com/xbit-gg/xln/models"
	"gorm.io/gorm"
)

func (m *manager) handleStalePayments() {
	pendingPayments, err := m.db.Repo.ListPendingPayments(m.db.DB)
	if err != nil {
		log.WithError(err).Fatal("Failed to get pending payments from DB")
	}
	if len(pendingPayments) > 0 {
		log.WithField("total", len(pendingPayments)).Info("Handling payment updates that occurred while XLN was offline")
	}
	for _, pp := range pendingPayments {
		log.WithField("payment", pp.PaymentHash).Debug("Handling stale pending payment")
		m.pendingPaymentCache.SetDefault(pp.PaymentHash, pp)
		go m.trackPendingPayment(pp.PaymentHash)

		// Prevent overloading LND
		time.Sleep(1 * time.Second)
	}
}

func (m *manager) handleStaleInvoices() {
	pendingInvoices, err := m.db.Repo.ListPendingInvoices(m.db.DB)
	if err != nil {
		log.WithError(err).Fatal("Failed to get pending invoices from DB")
	}
	if len(pendingInvoices) > 0 {
		log.WithField("amount", len(pendingInvoices)).Info("Handling invoice updates that occurred while XLN was offline")
	}
	for _, pi := range pendingInvoices {
		m.pendingInvoiceCache.SetDefault(pi.PaymentHash, pi)
		paymentHash, err := base64.StdEncoding.DecodeString(pi.PaymentHash)
		if err != nil {
			log.WithError(err).Fatal("Unable to decode paymentHash while handling invoice")
		}

		invoice, err := m.lnClient.LookupInvoice(context.Background(), &lnrpc.PaymentHash{RHash: paymentHash})
		if err != nil {
			log.WithError(err).Fatal("Failed to lookup stale invoice")
		}
		m.finalizeInvoice(invoice)

		// Prevent overloading LND
		// For some reason, LND will return "could not locate invoice" when sending requests too fast
		time.Sleep(1 * time.Second)
	}
}

func (m *manager) trackPendingInvoices() {
	stream, err := m.lnClient.SubscribeInvoices(context.Background(), &lnrpc.InvoiceSubscription{AddIndex: 0})
	if err != nil {
		log.WithError(err).Fatal("Failed to subscribe to invoice events")
	}
	for {
		invoice, err := stream.Recv()
		if err == io.EOF {
			log.Fatal("Subscription to invoice events closed by LND")
		}
		if err != nil {
			log.WithError(err).Fatal("Error getting invoice event from subscription")
		}
		log.WithFields(log.Fields{
			"paymentHash": base64.StdEncoding.EncodeToString(invoice.RHash),
			"state":       invoice.State.String(),
		}).Debug("Received invoice event")
		if invoice.State == lnrpc.Invoice_SETTLED || invoice.State == lnrpc.Invoice_CANCELED {
			m.finalizeInvoice(invoice)
		}
	}
}

func (m *manager) trackPendingPayment(paymentHash string) *Payment {
	pHash, err := hex.DecodeString(paymentHash)
	if err != nil {
		log.WithError(err).Fatal("Unable to decode paymentHash while handling payment")
	}

	stream, err := m.routerClient.TrackPaymentV2(context.Background(), &routerrpc.TrackPaymentRequest{
		PaymentHash:       pHash,
		NoInflightUpdates: true,
	})
	if err != nil {
		log.WithError(err).Fatal("Failed to track outgoing payment")
	}
	lnPayment, err := stream.Recv()
	if err != nil {
		log.WithError(err).Fatal("Failed to track outgoing payment")
	}
	log.WithFields(log.Fields{
		"hash":   lnPayment.PaymentHash,
		"value":  lnPayment.ValueMsat,
		"status": lnPayment.Status.String(),
		"reason": lnPayment.FailureReason.String(),
	}).Debug("Pending payment update")
	m.finalizePayment(paymentHash, lnPayment.Status == lnrpc.Payment_SUCCEEDED, lnPayment.FeeMsat)
	var failureReason string
	if lnPayment.FailureReason != lnrpc.PaymentFailureReason_FAILURE_REASON_NONE {
		failureReason = lnPayment.FailureReason.String()
	}
	return &Payment{
		Success:       lnPayment.Status == lnrpc.Payment_SUCCEEDED,
		AmountMsat:    uint64(lnPayment.ValueMsat),
		FeeMsat:       uint64(lnPayment.FeeMsat),
		FailureReason: failureReason,
	}
}

func (m *manager) finalizePayment(paymentHash string, success bool, feesPaid int64) {
	var pendingPayment *models.PendingPayment
	if pp, contains := m.pendingPaymentCache.Get(paymentHash); contains {
		pendingPayment = pp.(*models.PendingPayment)
		m.pendingPaymentCache.Delete(paymentHash)
	} else {
		log.WithField("hash", paymentHash).Debug("Cache miss for pending payment when finalizing")
		var err error
		pendingPayment, err = m.db.Repo.GetPendingPayment(m.db.DB, paymentHash)
		if err != nil {
			log.WithError(err).Fatal("Failed to lookup completed payment")
		}
	}

	err := m.db.Transaction(func(tx *gorm.DB) error {
		// Clear pending payment

		if err := m.db.Repo.DeletePendingPayment(tx, pendingPayment); err != nil {
			return err
		}
		if !success {
			return nil
		}
		err := m.db.Repo.DecrementWalletBalance(tx, pendingPayment.WalletUsername,
			pendingPayment.WalletID, pendingPayment.Amount-uint64(feesPaid))
		if err != nil {
			return err
		}

		paymentTime := time.Now().UTC()
		// Record transaction
		walTx := &models.Transaction{
			FromID:       &pendingPayment.WalletID,
			FromUsername: &pendingPayment.WalletUsername,
			ToID:         nil,
			ToUsername:   nil,
			Amount:       pendingPayment.Amount,
			FeesPaid:     uint64(feesPaid),
			InvoiceID:    &paymentHash,
			UpdatedAt:    paymentTime,
		}
		if err := m.db.Repo.CreateTransaction(tx, walTx); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"InvoiceID":    paymentHash,
				"fromID":       pendingPayment.WalletID,
				"fromUsername": pendingPayment.WalletUsername,
			}).Error("Failed to create transaction when finalizing payment")
			return err
		} else if err := m.db.Repo.UpdateInvoiceSettleTime(tx, paymentHash, &paymentTime); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"InvoiceID":    paymentHash,
				"fromID":       pendingPayment.WalletID,
				"fromUsername": pendingPayment.WalletUsername,
			}).Error("Failed to update invoice")
			return err
		}
		return nil
	})
	if err != nil {
		log.WithError(err).Fatal("Failed to update wallet balance after completed payment")
	}
	log.WithFields(log.Fields{
		"hash":   pendingPayment.PaymentHash,
		"wallet": pendingPayment.WalletID,
	}).Debug("Payment finalized")
}

func (m *manager) finalizeInvoice(invoice *lnrpc.Invoice) {
	paymentHash := base64.StdEncoding.EncodeToString(invoice.RHash)
	var pendingInvoice *models.PendingInvoice
	if pi, contains := m.pendingInvoiceCache.Get(paymentHash); contains {
		pendingInvoice = pi.(*models.PendingInvoice)
		m.pendingInvoiceCache.Delete(pendingInvoice.PaymentHash)
	} else {
		log.WithField("hash", paymentHash).Debug("Cache miss for pending invoice when finalizing")
		var err error
		pendingInvoice, err = m.db.Repo.GetPendingInvoice(m.db.DB, paymentHash)
		if err != nil {
			log.WithError(err).WithField("hash", paymentHash).Debug("Finalized invoice not associated with XLN")
			return
		}
	}

	log.WithFields(log.Fields{
		"wallet":      pendingInvoice.WalletID,
		"paymentHash": pendingInvoice.PaymentHash,
	}).Debug("finalizing invoice")

	err := m.db.Transaction(func(tx *gorm.DB) error {
		// Clear pending invoice
		res := tx.Delete(pendingInvoice)
		if res.Error != nil {
			return res.Error
		}
		if invoice.State != lnrpc.Invoice_SETTLED {
			return nil
		}

		// Update wallet balance
		wal := &models.Wallet{}
		wal, err := m.db.Repo.GetWallet(m.db.DB, pendingInvoice.WalletUsername, pendingInvoice.WalletID)
		if err != nil {
			return err
		}
		wal.Balance = wal.Balance + uint64(invoice.AmtPaidMsat)
		tx.Save(wal)

		// Record transaction
		walTx := &models.Transaction{
			FromID:       nil,
			FromUsername: nil,
			ToID:         &wal.ID,
			ToUsername:   &wal.Username,
			Amount:       uint64(invoice.AmtPaidMsat),
			FeesPaid:     0,
			InvoiceID:    &paymentHash,
			UpdatedAt:    time.Now(),
		}
		if err := m.db.Repo.CreateTransaction(tx, walTx); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.WithError(err).Fatal("Failed to update wallet balance after finalized invoice")
	}
	log.WithFields(log.Fields{
		"hash":   pendingInvoice.PaymentHash,
		"wallet": pendingInvoice.WalletID,
	}).Debug("Invoice finalized")
}

func (m *manager) handleSelfPayments(sUsername, sId, rUsername, rId string, payHash string, amount int64) error {
	err := m.db.Transaction(func(tx *gorm.DB) error {
		if cBal, err := m.db.Repo.GetConfirmedBalance(tx, sUsername, sId); err != nil {
			return err
		} else if cBal < uint64(amount) {
			log.WithFields(log.Fields{
				"user":        sUsername,
				"wallet":      sId,
				"paymentHash": payHash,
			}).Warn("Wallet attempted payment with insufficient funds")
			return errors.New("insufficient funds")
		}

		if err := m.db.Repo.SetInvoiceSenderAmount(tx, payHash, sUsername, sId, amount); err != nil {
			return err
		}
		err := m.db.Repo.CreateTransaction(tx, &models.Transaction{
			FromID:       &sId,
			FromUsername: &sUsername,
			ToID:         &rId,
			ToUsername:   &rUsername,
			Amount:       uint64(amount),
			FeesPaid:     0,
			InvoiceID:    &payHash,
			UpdatedAt:    time.Now().UTC(),
		})
		if err != nil {
			return err
		}
		if err := m.db.Repo.DeletePendingInvoice(tx, payHash); err != nil {
			return err
		}
		if err := m.db.Repo.IncrementWalletBalance(tx, rUsername, rId, uint64(amount)); err != nil {
			return err
		}
		if err := m.db.Repo.DecrementWalletBalance(tx, sUsername, sId, uint64(amount)); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	if _, contains := m.pendingInvoiceCache.Get(payHash); contains {
		m.pendingInvoiceCache.Delete(payHash)
	} else {
		log.WithField("paymentHash", payHash).Warn("Paid an invoice, but there was no record of a corresponding pending invoice")
	}

	return nil
}
