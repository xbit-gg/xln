package xln

import (
	"context"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/xbit-gg/xln/auth"
	"github.com/xbit-gg/xln/models"
	"github.com/xbit-gg/xln/resources/invoice"
	"github.com/xbit-gg/xln/util"
	"github.com/xbit-gg/xln/xlnrpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type xlnServer struct {
	xlnrpc.UnimplementedXlnServer

	xln *XLN
}

func NewXlnServer(xln *XLN) *xlnServer {
	return &xlnServer{xln: xln}
}

// A compile time check to ensure that xlnServer fully implements the xlnServer gRPC service.
var _ xlnrpc.XlnServer = (*xlnServer)(nil)

func (x xlnServer) GetInfo(ctx context.Context, request *xlnrpc.GetInfoRequest) (*xlnrpc.GetInfoResponse, error) {
	log.Debug("Xln.GetInfo called")
	identity, err := x.xln.AuthService.GetIdentityOfApiKey(request.ApiKey)
	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())
		log.WithError(err).Warn("GetInfo request failed")
		return nil, st.Err()
	} else if identity == nil {
		return &xlnrpc.GetInfoResponse{Version: x.xln.Version}, nil
	} else {
		return &xlnrpc.GetInfoResponse{Version: x.xln.Version, Identity: &xlnrpc.GetInfoResponse_IdentityType{
			Admin:  identity.Admin,
			User:   identity.User,
			Wallet: identity.Wallet,
		}}, nil
	}

}

func (x xlnServer) CreateWallet(ctx context.Context, request *xlnrpc.CreateWalletRequest) (*xlnrpc.CreateWalletResponse, error) {
	log.WithField("req", request).Debugf("%s called", "Xln.CreateWallet")
	username, err := x.xln.AuthService.ValidateUserCredentials(ctx, "Xln.CreateWallet")
	if err != nil {
		return nil, handleAuthErr(err)
	}

	err = util.ValidateWalletID(request.WalletId)
	if err != nil {
		st := status.New(codes.InvalidArgument, fmt.Sprintf("Invalid wallet id. Reason: %v", err))
		return nil, st.Err()
	}

	err = util.ValidateWalletName(request.WalletName)
	if request.WalletName != "" && err != nil {
		st := status.New(codes.InvalidArgument, fmt.Sprintf("Invalid wallet name. Reason: %v", err))
		return nil, st.Err()
	}

	wallet, err := x.xln.Wallets.CreateWallet(username, request.WalletId, request.WalletName)
	if err != nil {
		if strings.Contains(err.Error(), models.MsgWalletIDAlreadyExists) {
			st := status.New(codes.AlreadyExists, fmt.Sprintf("Failed to create wallet. Reason: %v", err))
			return nil, st.Err()
		} else {
			st := status.New(codes.Internal, fmt.Sprintf("Failed to create wallet. Reason: %v", err))
			log.WithError(err).Warn("CreateWallet request failed.")
			return nil, st.Err()
		}
	} else {
		log.WithFields(log.Fields{
			"user":   username,
			"wallet": wallet.ID,
		}).Info("Wallet created")

		return &xlnrpc.CreateWalletResponse{
			ApiKey:     wallet.ApiKey,
			WalletId:   wallet.ID,
			WalletName: *wallet.Name}, nil
	}
}

func (x xlnServer) DeleteWallet(ctx context.Context, request *xlnrpc.DeleteWalletRequest) (*xlnrpc.DeleteWalletResponse, error) {
	log.WithField("req", request).Debug("Xln.DeleteWallet called")
	username, err := x.xln.AuthService.ValidateWalletCredentials(ctx, request.WalletId, "Xln.DeleteWallet")
	if err != nil {
		return nil, handleAuthErr(err)
	}

	err = x.xln.Wallets.DeleteWallet(username, request.WalletId, strings.HasPrefix(x.xln.Config.DatabaseConnectionString, "postgres"))
	if err == nil {
		log.WithFields(log.Fields{
			"wallet": request.WalletId,
		}).Info("Wallet deleted")
		return &xlnrpc.DeleteWalletResponse{}, nil
	} else {
		st := status.New(codes.InvalidArgument, fmt.Sprintf("Failed to delete wallet. Reason: %v", err))
		log.WithError(err).WithFields(log.Fields{
			"wallet": request.WalletId,
		}).Warn("DeleteWallet request failed.")
		return nil, st.Err()
	}
}

func (x xlnServer) UpdateWalletOptions(ctx context.Context, request *xlnrpc.UpdateWalletOptionsRequest) (*xlnrpc.UpdateWalletOptionsResponse, error) {
	log.WithField("req", request).Debug("Xln.UpdateWalletOptions called")
	username, err := x.xln.AuthService.ValidateWalletCredentials(ctx, request.WalletId, "Xln.UpdateWalletOptions")
	if err != nil {
		return nil, handleAuthErr(err)
	}

	walletOptions := models.WalletOptions{}
	if request.Lock && request.Unlock {
		st := status.New(codes.InvalidArgument, "Invalid wallet options. "+
			"Reason: cannot set both `lock` and `unlock` to `true`")
		log.Warn("Invalid wallet options")
		return nil, st.Err()
	}
	if request.Lock {
		locked := true
		walletOptions.Locked = &locked
	}
	if request.Unlock {
		locked := false
		walletOptions.Locked = &locked
	}
	if request.WalletName != "" {
		err := util.ValidateWalletName(request.WalletName)
		if err != nil {
			st := status.New(codes.InvalidArgument, fmt.Sprintf("Invalid wallet name. Reason: %v", err))
			return nil, st.Err()
		}
		walletOptions.Name = &request.WalletName
	}
	err = x.xln.Wallets.UpdateWalletOptions(username, request.WalletId, &walletOptions)
	if err == nil {
		log.WithFields(log.Fields{
			"wallet": request.WalletId,
		}).Info("Wallet updated")
		return &xlnrpc.UpdateWalletOptionsResponse{}, nil
	} else {
		st := status.New(codes.InvalidArgument, fmt.Sprintf("Failed to update wallet. Reason: %v", err))
		log.WithError(err).WithFields(log.Fields{
			"wallet": request.WalletId,
		}).Warn("UpdateWalletOptions request failed.")
		return nil, st.Err()
	}
}

func (x xlnServer) ListWallets(ctx context.Context, request *xlnrpc.ListWalletsRequest) (*xlnrpc.ListWalletsResponse, error) {
	log.WithField("req", request).Debug("Xln.ListWallets called")
	username, err := x.xln.AuthService.ValidateUserCredentials(ctx, "Xln.ListWallets")
	if err != nil {
		return nil, handleAuthErr(err)
	}

	wallets, err := x.xln.Wallets.ListWallets(username)
	if err != nil {
		st := status.New(codes.Internal, fmt.Sprintf("Failed to list wallets. Reason: %v", err))
		log.WithError(err).WithFields(log.Fields{
			"user": username,
		}).Warn("ListWallets request failed")
		return nil, st.Err()
	}
	var walletIds []string
	var walletsData []*xlnrpc.Wallet
	for _, wallet := range wallets {
		walletIds = append(walletIds, wallet.ID)
	}
	if request.Data {
		for _, wallet := range wallets {
			walletsData = append(walletsData, convertWalletData(wallet))
		}
	}
	return &xlnrpc.ListWalletsResponse{
		WalletIds: walletIds,
		Data:      walletsData,
	}, nil
}

func (x xlnServer) GetWallet(ctx context.Context, request *xlnrpc.GetWalletRequest) (*xlnrpc.GetWalletResponse, error) {
	log.WithField("req", request).Debug("Xln.GetWallet called")
	username, err := x.xln.AuthService.ValidateWalletCredentials(ctx, request.WalletId, "Xln.GetWallet")
	if err != nil {
		return nil, handleAuthErr(err)
	}

	wallet, err := x.xln.Wallets.GetWallet(username, request.WalletId)
	if err == models.ErrWalletNotFound {
		return nil, status.New(codes.NotFound, fmt.Sprintf("Failed to get wallet. Reason: %v", err)).Err()
	} else if err != nil {
		st := status.New(codes.Internal, fmt.Sprintf("Failed to get wallet. Reason: %v", err))
		log.WithError(err).Warn("GetWallet request failed")
		return nil, st.Err()
	}
	txns, _, _, err := x.xln.Wallets.ListWalletTransactions(username, request.WalletId, time.Time{}, time.Now().UTC(), 0, 1, true)
	if err != nil {
		st := status.New(codes.Internal, fmt.Sprintf("Failed to get latest wallet transaction. Reason: %v", err))
		log.WithError(err).Warn("GetWallet request failed to get latest transaction")
		return nil, st.Err()
	}

	res := xlnrpc.GetWalletResponse{
		ApiKey:       wallet.ApiKey,
		WalletId:     wallet.ID,
		WalletName:   *wallet.Name,
		Balance:      wallet.Balance,
		CreationTime: timestamppb.New(wallet.CreatedAt),
		Locked:       wallet.Locked,
	}
	if wallet.LinkKey != nil {
		res.LinkKey = *wallet.LinkKey
	}
	if wallet.LinkLabel != nil {
		res.LinkLabel = *wallet.LinkLabel
	}
	if len(txns) > 0 {
		res.LatestTransaction = convertTransaction(txns[0])
	}

	return &res, nil
}

func (x xlnServer) ListWalletPendingInvoices(ctx context.Context, request *xlnrpc.ListWalletPendingInvoicesRequest) (*xlnrpc.ListWalletPendingInvoicesResponse, error) {
	log.WithField("req", request).Debug("Xln.ListWalletPendingInvoices called")
	username, err := x.xln.AuthService.ValidateWalletCredentials(ctx, request.WalletId, "Xln.ListWalletPendingInvoices")
	if err != nil {
		return nil, handleAuthErr(err)
	}

	pendingInvoices, err := x.xln.PendingInvoices.ListWalletPendingInvoices(username, request.WalletId)
	if err != nil {
		st := status.New(codes.Internal, fmt.Sprintf("Failed to list pending invoices. Reason: %v", err))
		log.WithError(err).Warn("ListWalletPendingInvoices request failed")
		return nil, st.Err()
	} else {
		var res xlnrpc.ListWalletPendingInvoicesResponse
		for _, invoice := range pendingInvoices {
			res.PendingInvoices = append(res.PendingInvoices, &xlnrpc.WalletPendingInvoiceSummary{
				Amount:      invoice.Amount,
				CreatedAt:   timestamppb.New(invoice.CreatedAt),
				PaymentHash: invoice.PaymentHash,
			})
			res.TotalAmount += invoice.Amount
		}
		return &res, nil
	}
}

func (x xlnServer) ListWalletPendingPayments(ctx context.Context, request *xlnrpc.ListWalletPendingPaymentsRequest) (*xlnrpc.ListWalletPendingPaymentsResponse, error) {
	log.WithField("req", request).Debug("Xln.ListWalletPendingPayments called")
	username, err := x.xln.AuthService.ValidateWalletCredentials(ctx, request.WalletId, "Xln.ListWalletPendingPayments")
	if err != nil {
		return nil, handleAuthErr(err)
	}
	pendingPayments, err := x.xln.PendingPayments.ListWalletPendingPayments(username, request.WalletId)
	if err != nil {
		st := status.New(codes.Internal, fmt.Sprintf("Failed to list pending invoices. Reason: %v", err))
		log.WithError(err).Warn("ListWalletPendingPayments request failed")
		return nil, st.Err()
	} else {
		var res xlnrpc.ListWalletPendingPaymentsResponse
		for _, payment := range pendingPayments {
			res.PendingPayments = append(res.PendingPayments, &xlnrpc.WalletPaymentSummary{
				Amount:      payment.Amount,
				CreatedAt:   timestamppb.New(payment.CreatedAt),
				PaymentHash: payment.PaymentHash,
			})
			res.TotalAmount += payment.Amount
		}
		return &res, nil
	}
}

func (x xlnServer) ListWalletTransactions(ctx context.Context, request *xlnrpc.ListWalletTransactionsRequest) (*xlnrpc.ListWalletTransactionsResponse, error) {
	log.WithField("req", request).Debug("Xln.ListWalletTransactions called")
	username, err := x.xln.AuthService.ValidateWalletCredentials(ctx, request.WalletId, "Xln.ListWalletTransactions")
	if err != nil {
		return nil, handleAuthErr(err)
	}

	var startTime, endTime time.Time
	if request.FromTime != nil {
		startTime = request.FromTime.AsTime().UTC()
	}
	if request.ToTime != nil {
		endTime = request.ToTime.AsTime().UTC()
	} else {
		endTime = time.Now().UTC()
	}
	transactions, nextOffset, total, err := x.xln.Wallets.ListWalletTransactions(username, request.WalletId, startTime, endTime, uint(request.Offset), uint(request.Limit), request.Descending)
	if err != nil {
		st := status.New(codes.Internal, fmt.Sprintf("Failed to list transcations. Reason: %v", err))
		log.WithError(err).Warn("ListWalletTransactions request failed")
		return nil, st.Err()
	}
	var txns []*xlnrpc.Transaction
	for _, tx := range transactions {
		txns = append(txns, convertTransaction(tx))
	}

	return &xlnrpc.ListWalletTransactionsResponse{Transactions: txns, NextOffset: int32(nextOffset), TotalRecords: total}, nil
}

func (x xlnServer) GetWalletTransaction(ctx context.Context, request *xlnrpc.GetWalletTransactionRequest) (*xlnrpc.GetWalletTransactionResponse, error) {
	log.WithField("req", request).Debug("Xln.GetWalletTransaction called")
	username, err := x.xln.AuthService.ValidateWalletCredentials(ctx, request.WalletId, "Xln.GetWalletTransaction")
	if err != nil {
		return nil, handleAuthErr(err)
	}

	transaction, err := x.xln.Wallets.GetTransaction(username, request.WalletId, request.TxId)
	if err != nil {
		st := status.New(codes.Internal, fmt.Sprintf("Failed to get transcation. Reason: %v", err))
		log.WithError(err).Warn("GetWalletTransaction request failed")
		return nil, st.Err()
	}

	return convertFullTransaction(transaction), nil
}

func (x xlnServer) CreateInvoice(ctx context.Context, request *xlnrpc.CreateInvoiceRequest) (*xlnrpc.CreateInvoiceResponse, error) {
	log.WithField("req", request).Debug("Xln.CreateInvoice called")
	username, err := x.xln.AuthService.ValidateWalletCredentials(ctx, request.WalletId, "Xln.CreateInvoice")
	if err != nil {
		return nil, handleAuthErr(err)
	}

	inv, err := x.xln.Invoices.CreateInvoice(username, request.WalletId, request.Memo, request.Value, request.Expiry)
	if err != nil {
		log.WithError(err).Warn("CreateInvoice request failed")
		st := status.New(codes.InvalidArgument, fmt.Sprintf("Failed to create invoice. Reason: %v", err))
		return nil, st.Err()
	}
	return &xlnrpc.CreateInvoiceResponse{
		PaymentHash:    inv.PaymentHash,
		PaymentRequest: inv.PaymentRequest,
	}, nil
}

func (x xlnServer) ListWalletInvoices(ctx context.Context, request *xlnrpc.ListWalletInvoicesRequest) (*xlnrpc.ListWalletInvoicesResponse, error) {
	log.WithField("req", request).Debug("Xln.ListWalletInvoices called")
	username, err := x.xln.AuthService.ValidateWalletCredentials(ctx, request.WalletId, "Xln.ListWalletInvoices")
	if err != nil {
		return nil, handleAuthErr(err)
	}
	invoices, err := x.xln.Invoices.ListWalletInvoices(username, request.WalletId)
	if err != nil {
		st := status.New(codes.Internal, fmt.Sprintf("Failed to list  invoices. Reason: %v", err))
		log.WithError(err).Warn("ListWalletInvoices request failed")
		return nil, st.Err()
	} else {
		var res xlnrpc.ListWalletInvoicesResponse
		for _, invoice := range invoices {
			res.Invoices = append(res.Invoices, invoice.PaymentHash)
		}
		return &res, nil
	}
}

func (x xlnServer) GetWalletInvoice(ctx context.Context, request *xlnrpc.GetWalletInvoiceRequest) (*xlnrpc.GetWalletInvoiceResponse, error) {
	log.WithField("req", request).Debug("Xln.GetWalletInvoice called")
	username, err := x.xln.AuthService.ValidateWalletCredentials(ctx, request.WalletId, "Xln.GetWalletInvoice")
	if err != nil {
		return nil, handleAuthErr(err)
	}
	invoice, err := x.xln.Invoices.GetWalletInvoice(username, request.WalletId, request.PaymentHash)
	if err != nil {
		st := status.New(codes.Internal, fmt.Sprintf("Failed to get invoice. Reason: %v", err))
		log.WithError(err).Warn("GetWalletInvoice request failed")
		return nil, st.Err()
	} else {
		res := xlnrpc.GetWalletInvoiceResponse{
			Amount:         invoice.Amount,
			Memo:           invoice.Memo,
			PaymentHash:    invoice.PaymentHash,
			PaymentRequest: invoice.PaymentRequest,
			Pubkey:         invoice.Pubkey,
			SettledAt:      timestamppb.New(invoice.Settled),
			Timestamp:      timestamppb.New(invoice.Timestamp),
		}
		if invoice.Preimage != nil {
			res.Preimage = *invoice.Preimage
		}
		if invoice.RecipientID != nil && invoice.RecipientUsername != nil {
			res.RecipientId = *invoice.RecipientID
			res.RecipientUsername = *invoice.RecipientUsername
		}
		if invoice.SenderID != nil && invoice.SenderUsername != nil {
			res.SenderId = *invoice.SenderID
			res.SenderUsername = *invoice.SenderUsername
		}
		return &res, nil
	}
}

func (x xlnServer) PayInvoice(ctx context.Context, request *xlnrpc.PayInvoiceRequest) (*xlnrpc.PayInvoiceResponse, error) {
	log.WithField("req", request).Debug("Xln.PayInvoice called")
	username, err := x.xln.AuthService.ValidateWalletCredentials(ctx, request.WalletId, "Xln.PayInvoice")
	if err != nil {
		return nil, handleAuthErr(err)
	}

	if request.Amount > 0 {
		_, err = x.xln.Invoices.PayInvoiceAmount(username, request.WalletId, request.PaymentRequest, false, int64(request.Amount))
	} else {
		_, err = x.xln.Invoices.PayInvoice(username, request.WalletId, request.PaymentRequest, false)
	}
	if err != nil {
		st := status.New(codes.InvalidArgument, fmt.Sprintf("Failed to pay invoice. Reason: %v", err))
		return &xlnrpc.PayInvoiceResponse{PaymentInitiated: false}, st.Err()
	}
	return &xlnrpc.PayInvoiceResponse{PaymentInitiated: true}, nil
}

func (x xlnServer) PayInvoiceSync(ctx context.Context, request *xlnrpc.PayInvoiceRequest) (*xlnrpc.PayInvoiceSyncResponse, error) {
	log.WithField("req", request).Debug("Xln.PayInvoiceSync called")
	username, err := x.xln.AuthService.ValidateWalletCredentials(ctx, request.WalletId, "Xln.PayInvoiceSync")
	if err != nil {
		return nil, handleAuthErr(err)
	}

	var payment *invoice.Payment
	if request.Amount > 0 {
		payment, err = x.xln.Invoices.PayInvoiceAmount(username, request.WalletId, request.PaymentRequest, true, int64(request.Amount))
	} else {
		payment, err = x.xln.Invoices.PayInvoice(username, request.WalletId, request.PaymentRequest, true)
	}
	if err != nil {
		st := status.New(codes.InvalidArgument, fmt.Sprintf("Failed to pay invoice. Reason: %v", err))
		return &xlnrpc.PayInvoiceSyncResponse{
			Success:       false,
			Amount:        0,
			FeesPaid:      0,
			FailureReason: "INVALID",
		}, st.Err()
	} else if payment == nil {
		return &xlnrpc.PayInvoiceSyncResponse{
			Success:       true,
			Amount:        request.Amount,
			FeesPaid:      0,
			FailureReason: "",
		}, nil
	} else {
		return &xlnrpc.PayInvoiceSyncResponse{
			Success:       payment.Success,
			Amount:        payment.AmountMsat,
			FeesPaid:      payment.FeeMsat,
			FailureReason: payment.FailureReason,
		}, nil
	}
}

func (x xlnServer) Transfer(ctx context.Context, request *xlnrpc.TransferRequest) (*xlnrpc.TransferResponse, error) {
	log.WithField("req", request).Debug("Xln.Transfer called")
	username, err := x.xln.AuthService.ValidateWalletCredentials(ctx, request.WalletId, "Xln.Transfer")
	if err != nil {
		return nil, handleAuthErr(err)
	}

	txn, err := x.xln.Wallets.Transfer(username, request.WalletId, request.ToWalletId, request.Amount)
	if err == models.ErrCannotTransactWithLockedWallet {
		return nil, status.New(codes.FailedPrecondition, fmt.Sprintf(
			"Failed to transfer funds between wallets. Reason: %v", err)).Err()
	} else if err != nil {
		st := status.New(codes.Internal, fmt.Sprintf("Failed to transfer funds between wallets. Reason: %v", err))
		log.WithError(err).WithFields(log.Fields{
			"user":        username,
			"recipientId": request.WalletId,
			"senderId":    request.ToWalletId,
			"amount":      request.Amount,
		}).Warn("Transfer request failed.")
		return nil, st.Err()
	} else {
		log.WithError(err).WithFields(log.Fields{
			"from": request.WalletId,
			"to":   request.ToWalletId,
		}).Info("sent funds between wallets")

		return &xlnrpc.TransferResponse{Success: true, TransactionId: txn.ID}, nil
	}
}

func (x xlnServer) ListUserTransactions(ctx context.Context, request *xlnrpc.ListUserTransactionsRequest) (*xlnrpc.ListUserTransactionsResponse, error) {
	log.WithField("req", request).Debug("Xln.ListUserTransactions called")
	username, err := x.xln.AuthService.ValidateUserCredentials(ctx, "Xln.ListUserTransactions")
	if err != nil {
		return nil, handleAuthErr(err)
	}

	var startTime, endTime time.Time
	if request.FromTime != nil {
		startTime = request.FromTime.AsTime().UTC()
	}
	if request.ToTime != nil {
		endTime = request.ToTime.AsTime().UTC()
	} else {
		endTime = time.Now().UTC()
	}
	transactions, nextOffset, total, err := x.xln.Users.ListUserTransactions(username, startTime, endTime, uint(request.Offset), uint(request.Limit), request.Descending)
	if err != nil {
		st := status.New(codes.Internal, fmt.Sprintf("Failed to list transcations. Reason: %v", err))
		log.WithError(err).WithField("user", username).Warn("ListUserTransactions request failed")
		return nil, st.Err()
	}
	var txns []*xlnrpc.Transaction
	for _, tx := range transactions {
		txns = append(txns, convertTransaction(tx))
	}

	return &xlnrpc.ListUserTransactionsResponse{Transactions: txns, NextOffset: int32(nextOffset), TotalRecords: total}, nil
}

func (x xlnServer) Validate(ctx context.Context, request *xlnrpc.ValidateRequest) (*xlnrpc.ValidateResponse, error) {
	log.WithField("req", request).Debug("Xln.Validate called")
	var (
		errMsgs     []string
		numValiated int = 0
	)
	if request.Username != "" {
		numValiated += 1
		if err := util.ValidateUsername(request.Username); err != nil {
			errMsgs = append(errMsgs, err.Error())
		}
	}
	if request.WalletId != "" {
		numValiated += 1
		if err := util.ValidateWalletID(request.WalletId); err != nil {
			errMsgs = append(errMsgs, err.Error())
		}
	}
	if request.WalletName != "" {
		numValiated += 1
		if err := util.ValidateWalletName(request.WalletName); err != nil {
			errMsgs = append(errMsgs, err.Error())
		}
	}

	if numValiated == 0 {
		st := status.New(codes.InvalidArgument, "If fields were provided they are unrecognized and unvalidatable")
		return nil, st.Err()
	} else if len(errMsgs) == 0 {
		return &xlnrpc.ValidateResponse{Valid: true, Reason: ""}, nil
	} else {
		return &xlnrpc.ValidateResponse{Valid: false, Reason: strings.Join(errMsgs, "; ")}, nil
	}
}

func (x xlnServer) GetUser(ctx context.Context, request *xlnrpc.GetUserRequest) (*xlnrpc.GetUserResponse, error) {
	log.WithField("req", request).Debug("Xln.GetUser called")
	username, err := x.xln.AuthService.ValidateUserCredentials(ctx, "Xln.GetUser")
	if err != nil {
		return nil, handleAuthErr(err)
	}
	user, err := x.xln.Users.GetUser(username)
	if err == models.ErrUserNotFound {
		return nil, status.New(codes.NotFound, fmt.Sprintf("Failed to get user. Reason: %v", err)).Err()
	} else if err != nil {
		st := status.New(codes.Internal, fmt.Sprintf("Failed to get user. Reason: %v", err))
		log.WithError(err).Warn("GetWallet request failed")
		return nil, st.Err()
	}

	res := xlnrpc.GetUserResponse{
		ApiKey:    user.ApiKey,
		CreatedAt: timestamppb.New(user.CreatedAt),
	}
	if user.LinkKey != nil {
		res.LinkKey = *user.LinkKey
	}
	if user.LinkLabel != nil {
		res.LinkLabel = *user.LinkLabel
	}

	return &res, nil
}

func (x xlnServer) UserLinkWallet(ctx context.Context, request *xlnrpc.UserLinkWalletRequest) (*xlnrpc.UserLinkWalletResponse, error) {
	log.WithField("req", request).Debug("Xln.UserLinkWallet called")
	username, err := x.xln.AuthService.ValidateUserCredentials(ctx, "Xln.UserLinkWallet")
	if err != nil {
		return nil, handleAuthErr(err)
	}

	lnurl, err := x.xln.LNURLAuths.UserLinkAuth(username, request.Label)
	if err != nil {
		st := status.New(codes.Internal, fmt.Sprintf("Failed to complete link wallet request. Reason: %v", err))
		log.WithError(err).WithField("user", username).Warn("UserLinkWallet request failed")
		return nil, st.Err()
	}
	return &xlnrpc.UserLinkWalletResponse{Lnurl: lnurl}, nil
}

func (x xlnServer) LinkWallet(ctx context.Context, request *xlnrpc.LinkWalletRequest) (*xlnrpc.LinkWalletResponse, error) {
	log.WithField("req", request).Debug("Xln.LinkWallet called")
	username, err := x.xln.AuthService.ValidateWalletCredentials(ctx, request.WalletId, "Xln.LinkWallet")
	if err != nil {
		return nil, handleAuthErr(err)
	}

	lnurl, err := x.xln.LNURLAuths.WalletLinkAuth(username, &request.WalletId, request.Label)
	if err != nil {
		st := status.New(codes.Internal, fmt.Sprintf("Failed to complete link wallet request. Reason: %v", err))
		log.WithError(err).WithField("user", username).Warn("UserLinkWallet request failed")
		return nil, st.Err()
	}
	return &xlnrpc.LinkWalletResponse{Lnurl: lnurl}, nil
}

func (x xlnServer) UserLogin(ctx context.Context, request *xlnrpc.UserLoginRequest) (*xlnrpc.UserLoginResponse, error) {
	log.WithField("req", request).Debug("Xln.UserLogin called")
	lnurl, err := x.xln.LNURLAuths.UserAuth(request.Username)
	if err != nil {
		st := status.New(codes.Internal, fmt.Sprintf("Failed to start user login. Reason: %v", err))
		log.WithError(err).WithField("user", request.Username).Warn("UserLogin request failed")
		return nil, st.Err()
	}
	return &xlnrpc.UserLoginResponse{Lnurl: lnurl}, nil
}

func (x xlnServer) WalletLogin(ctx context.Context, request *xlnrpc.WalletLoginRequest) (*xlnrpc.WalletLoginResponse, error) {
	log.WithField("req", request).Debug("Xln.WalletLogin called")
	lnurl, err := x.xln.LNURLAuths.WalletAuth(request.Username, &request.WalletId)
	if err != nil {
		st := status.New(codes.Internal, fmt.Sprintf("Failed to start wallet login. Reason: %v", err))
		log.WithError(err).WithField("wallet", request.WalletId).Warn("WalletLogin request failed")
		return nil, st.Err()
	}
	return &xlnrpc.WalletLoginResponse{Lnurl: lnurl}, nil
}

func (x xlnServer) LoginStatus(ctx context.Context, request *xlnrpc.LoginStatusRequest) (*xlnrpc.LoginStatusResponse, error) {
	log.WithField("req", request).Debug("Xln.LoginStatus called")
	apiKey, err := x.xln.LNURLAuths.GetAuthKey(request.K1)
	if err != nil {
		st := status.New(codes.Unauthenticated, fmt.Sprintf("Not logged in. Reason: %v", err))
		log.WithError(err).WithField("k1", request.K1).Debug("LoginStatus request failed")
		return nil, st.Err()
	}
	return &xlnrpc.LoginStatusResponse{ApiKey: apiKey}, nil
}

func convertWalletData(wallet *models.Wallet) *xlnrpc.Wallet {
	walletData := xlnrpc.Wallet{
		Id:           wallet.ID,
		Balance:      wallet.Balance,
		CreationTime: timestamppb.New(wallet.CreatedAt),
	}
	if wallet.Name != nil {
		walletData.Name = *wallet.Name
	}
	return &walletData
}

func (x xlnServer) CreateLNURLW(ctx context.Context, request *xlnrpc.CreateLNURLWRequest) (*xlnrpc.CreateLNURLWResponse, error) {
	log.WithField("req", request).Debug("Xln.CreateLNURLW called")
	username, err := x.xln.AuthService.ValidateWalletCredentials(ctx, request.WalletId, "Xln.GenerateLNURLW")
	if err != nil {
		return nil, handleAuthErr(err)
	}
	var expiryUTC time.Time
	if request.ExpireAt != nil {
		expiryUTC = request.ExpireAt.AsTime().UTC()
	}
	lnurl, err := x.xln.LNURLWithdraw.CreateLNURLW(username, request.WalletId, request.Description, request.MinMsats, request.MaxMsats,
		uint(request.MaxReuses), expiryUTC)
	if err != nil {
		st := status.New(codes.Internal, fmt.Sprintf("Failed to create LNURLW. Reason: %v", err))
		log.WithError(err).Warn("CreateLNURLW request failed")
		return nil, st.Err()
	}

	res := xlnrpc.CreateLNURLWResponse{
		Url: lnurl,
	}

	return &res, nil
}

func (x xlnServer) GetLNURLW(ctx context.Context, request *xlnrpc.GetLNURLWRequest) (*xlnrpc.GetLNURLWResponse, error) {
	log.WithField("req", request).Debug("Xln.GetLNURLW called")
	username, err := x.xln.AuthService.ValidateWalletCredentials(ctx, request.WalletId, "Xln.GenerateLNURLW")
	if err != nil {
		return nil, handleAuthErr(err)
	}
	lnurl, err := x.xln.LNURLWithdraw.GetLNURLW(username, request.WalletId, request.K1)
	if err != nil {
		st := status.New(codes.Internal, fmt.Sprintf("Failed to generate LNURLW. Reason: %v", err))
		log.WithError(err).Warn("GetLNURLW request failed")
		return nil, st.Err()
	}

	res := xlnrpc.GetLNURLWResponse{
		Url: lnurl,
	}

	return &res, nil
}

func convertTransaction(transaction *models.Transaction) *xlnrpc.Transaction {
	tx := &xlnrpc.Transaction{}
	tx.Id = transaction.ID
	tx.Time = timestamppb.New(transaction.CreatedAt)
	if transaction.FromID != nil {
		tx.FromWallet = *transaction.FromID
	}
	if transaction.FromUsername != nil {
		tx.FromUser = *transaction.FromUsername
	}
	if transaction.ToID != nil {
		tx.ToWallet = *transaction.ToID
	}
	if transaction.ToUsername != nil {
		tx.ToUser = *transaction.ToUsername
	}
	tx.Amount = transaction.Amount
	tx.FeesPaid = transaction.FeesPaid
	return tx
}

func convertFullTransaction(tx *models.Transaction) *xlnrpc.GetWalletTransactionResponse {
	res := &xlnrpc.GetWalletTransactionResponse{
		Id:           tx.ID,
		CreationTime: timestamppb.New(tx.CreatedAt),
		UpdateTime:   timestamppb.New(tx.UpdatedAt),
		Amount:       tx.Amount,
		FeesPaid:     tx.FeesPaid,
	}
	if tx.FromID != nil {
		res.From = *tx.FromID
	}
	if tx.ToID != nil {
		res.To = *tx.ToID
	}
	if tx.Label != nil {
		res.Label = *tx.Label
	}
	if tx.Invoice != nil {
		res.Invoice = &xlnrpc.Invoice{
			PaymentHash:    tx.Invoice.PaymentHash,
			PaymentRequest: tx.Invoice.PaymentRequest,
			Pubkey:         tx.Invoice.Pubkey,
			Memo:           tx.Invoice.Memo,
			Amount:         tx.Invoice.Amount,
		}
		res.Invoice.Amount = tx.Invoice.Amount
		res.Invoice.Time = timestamppb.New(tx.Invoice.Timestamp)
		if tx.Invoice.RecipientID != nil {
			res.Invoice.RecipientId = *tx.Invoice.RecipientID
		}
		if tx.Invoice.RecipientUsername != nil {
			res.Invoice.RecipientUsername = *tx.Invoice.RecipientUsername
		}
		if tx.Invoice.SenderID != nil {
			res.Invoice.SenderId = *tx.Invoice.SenderID
		}
		if tx.Invoice.SenderUsername != nil {
			res.Invoice.SenderUsername = *tx.Invoice.RecipientUsername
		}
	}
	return res
}

func handleAuthErr(err error) error {
	var st *status.Status
	switch err {
	case auth.ErrInvalidHeaderFormat, auth.ErrMissingUsername, auth.ErrParsingContext, auth.ErrMissingUsernameOrWallet:
		st = status.New(codes.InvalidArgument, err.Error())
	default:
		st = status.New(codes.Unauthenticated, auth.MsgFailedAuthentication)
	}
	return st.Err()
}
