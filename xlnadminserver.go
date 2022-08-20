package xln

import (
	"context"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/xbit-gg/xln/models"
	"github.com/xbit-gg/xln/util"
	"github.com/xbit-gg/xln/xlnrpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type xlnAdminServer struct {
	xlnrpc.UnimplementedXlnAdminServer

	xln *XLN
}

func NewXlnAdminServer(xln *XLN) *xlnAdminServer {
	return &xlnAdminServer{xln: xln}
}

// A compile time check to ensure that xlnAdminServer fully implements the XlnAdminServer gRPC service.
var _ xlnrpc.XlnAdminServer = (*xlnAdminServer)(nil)

func (x xlnAdminServer) GetInfo(ctx context.Context, request *xlnrpc.GetAdminInfoRequest) (*xlnrpc.GetAdminInfoResponse, error) {
	log.WithField("req", request).Debug("XlnAdmin.GetInfo called")
	err := x.xln.AuthService.ValidateAdminCredentials(ctx, "XlnAdmin.GetInfo")
	if err != nil {
		st := status.New(codes.Unauthenticated, fmt.Sprintf("invalid authentication for Admin GetInfo"))
		log.WithError(err).Warn("XlnAdmin.GetInfo request failed authentication")
		return nil, st.Err()
	}
	users, err := x.xln.Users.ListUsers()
	if err != nil {
		st := status.New(codes.Internal, fmt.Sprintf("unable to list users: %v", err))
		log.WithError(err).Info("XlnAdmin.GetInfo request could not list users")
		return nil, st.Err()
	}

	return &xlnrpc.GetAdminInfoResponse{
		Version: x.xln.Version,
		Users:   uint64(len(users)),
	}, nil
}

func (x xlnAdminServer) CreateUser(ctx context.Context, request *xlnrpc.CreateUserRequest) (*xlnrpc.CreateUserResponse, error) {
	log.WithField("req", request).Debug("XlnAdmin.CreateUser called")
	err := x.xln.AuthService.ValidateAdminCredentials(ctx, "XlnAdmin.CreateUser")
	if err != nil {
		st := status.New(codes.Unauthenticated, fmt.Sprintf("invalid authentication for CreateUser"))
		log.WithError(err).Warn("CreateUser request failed authentication")
		return nil, st.Err()
	}
	err = util.ValidateUsername(request.Username)
	if err != nil {
		st := status.New(codes.InvalidArgument, fmt.Sprintf("Invalid user name: %v", err))
		log.WithError(err).Warn("CreateUser requested with invalid user name")
		return nil, st.Err()
	}
	user, err := x.xln.Users.CreateUser(request.Username)
	if err != nil {
		if strings.Contains(err.Error(), models.MsgDuplicateUsername) {
			st := status.New(codes.AlreadyExists, fmt.Sprintf("unable to create user: %v", err))
			return nil, st.Err()
		} else {
			st := status.New(codes.InvalidArgument, fmt.Sprintf("unable to create user: %v", err))
			log.WithError(err).WithFields(log.Fields{
				"user": request.Username,
			}).Info("CreateUser request could not create user")
			return nil, st.Err()
		}
	}
	log.WithFields(log.Fields{
		"user": user.Username,
	}).Info("Created user.")

	return &xlnrpc.CreateUserResponse{
		Username: user.Username,
		ApiKey:   user.ApiKey,
	}, nil
}

func (x xlnAdminServer) DeleteUser(ctx context.Context, request *xlnrpc.DeleteUserRequest) (*xlnrpc.DeleteUserResponse, error) {
	log.WithField("req", request).Debug("XlnAdmin.DeleteUser called")
	err := x.xln.AuthService.ValidateAdminCredentials(ctx, "XlnAdmin.DeleteUser")
	if err != nil {
		st := status.New(codes.Unauthenticated, "invalid authentication for DeleteUser")
		log.WithError(err).Warn("DeleteUser request failed authentication")
		return nil, st.Err()
	}
	err = x.xln.Users.DeleteUser(request.Username, strings.HasPrefix(x.xln.Config.DatabaseConnectionString, "postgres"))
	if err == models.ErrUserNotFound {
		st := status.New(codes.NotFound, err.Error())
		return nil, st.Err()
	} else if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())
		return nil, st.Err()
	} else {
		return &xlnrpc.DeleteUserResponse{}, nil
	}
}

func (x xlnAdminServer) UpdateWallet(ctx context.Context, request *xlnrpc.UpdateWalletRequest) (*xlnrpc.UpdateWalletResponse, error) {
	log.WithField("req", request).Debug("XlnAdmin.UpdateWallet called")
	err := x.xln.AuthService.ValidateAdminCredentials(ctx, "XlnAdmin.UpdateWallet")
	if err != nil {
		st := status.New(codes.Unauthenticated, fmt.Sprintf("Invalid authentication for UpdateWallet. Reason: %v", err))
		log.WithError(err).Warn("UpdateWallet request failed authentication")
		return nil, st.Err()
	}
	walletOptions := models.WalletOptions{}
	if request.Lock && request.Unlock {
		st := status.New(codes.InvalidArgument, "Invalid wallet options. "+
			"Reason: cannot set both `lock` and `unlock` to `true`")
		log.Warn("Invalid wallet options")
		return nil, st.Err()
	}
	if request.UpdateBalance {
		newBalance := request.Balance
		walletOptions.Balance = &newBalance
	} else if request.Balance != 0 {
		st := status.New(codes.InvalidArgument, "Cannot update balance without setting updateBalance parameter to true")
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
			log.WithError(err).Warn("UpdateWallet requested with invalid wallet name")
			return nil, st.Err()
		}
		walletOptions.Name = &request.WalletName
	}
	err = x.xln.Wallets.UpdateWalletOptions(request.Username, request.WalletId, &walletOptions)
	if err == models.ErrWalletNotFound {
		st := status.New(codes.NotFound, err.Error())
		return nil, st.Err()
	} else if err != nil {
		st := status.New(codes.InvalidArgument, fmt.Sprintf("Failed to update wallet. Reason: %v", err))
		log.WithError(err).WithFields(log.Fields{
			"wallet": request.WalletId,
		}).Warn("UpdateWallet request failed.")
		return nil, st.Err()
	} else {
		log.WithFields(log.Fields{
			"wallet": request.WalletId,
		}).Info("Wallet updated")
		return &xlnrpc.UpdateWalletResponse{}, nil
	}
}

func (x xlnAdminServer) ListUsers(ctx context.Context, request *xlnrpc.ListUsersRequest) (*xlnrpc.ListUsersResponse, error) {
	log.WithField("req", request).Debug("XlnAdmin.ListUsers called")
	err := x.xln.AuthService.ValidateAdminCredentials(ctx, "XlnAdmin.ListUsers")
	if err != nil {
		st := status.New(codes.Unauthenticated, fmt.Sprintf("invalid authentication for ListUsers"))
		log.WithError(err).Warn("ListUsers request failed authentication")
		return nil, st.Err()
	}
	users, err := x.xln.Users.ListUsers()
	if err != nil {
		st := status.New(codes.InvalidArgument, fmt.Sprintf("unable to list users: %v", err))
		log.WithError(err).Error("ListUsers request could not list users")
		return nil, st.Err()
	}
	var usernames []string
	for _, user := range users {
		usernames = append(usernames, user.Username)
	}
	return &xlnrpc.ListUsersResponse{
		Usernames: usernames,
	}, nil
}

func (x xlnAdminServer) AdminDeleteWallet(ctx context.Context, request *xlnrpc.AdminDeleteWalletRequest) (*xlnrpc.AdminDeleteWalletResponse, error) {
	log.WithField("req", request).Debug("XlnAdmin.DeleteUser called")
	err := x.xln.AuthService.ValidateAdminCredentials(ctx, "XlnAdmin.DeleteUser")
	if err != nil {
		st := status.New(codes.Unauthenticated, "invalid authentication for DeleteUser")
		log.WithError(err).Warn("DeleteUser request failed authentication")
		return nil, st.Err()
	}
	err = x.xln.Wallets.AdminDeleteWallet(request.Username, request.WalletId, strings.HasPrefix(x.xln.Config.DatabaseConnectionString, "postgres"))
	if err == models.ErrWalletNotFound {
		st := status.New(codes.NotFound, err.Error())
		return nil, st.Err()
	} else if err != nil {
		st := status.New(codes.Internal, fmt.Sprintf("Failed to delete wallet. Reason: %v", err.Error()))
		return nil, st.Err()
	} else {
		return &xlnrpc.AdminDeleteWalletResponse{}, nil
	}
}

func (x xlnAdminServer) GetInvoice(ctx context.Context, request *xlnrpc.GetInvoiceRequest) (*xlnrpc.GetInvoiceResponse, error) {
	log.WithField("req", request).Debug("XlnAdmin.GetInvoice called")
	err := x.xln.AuthService.ValidateAdminCredentials(ctx, "XlnAdmin.GetInvoice")
	if err != nil {
		st := status.New(codes.Unauthenticated, fmt.Sprintf("Invalid authentication for GetInvoice. Reason: %v", err))
		log.WithError(err).Warn("GetInvoice request failed authentication")
		return nil, st.Err()
	}
	invoice, err := x.xln.Invoices.GetInvoice(request.PaymentHash)
	if err == models.ErrInvoiceNotFound {
		st := status.New(codes.NotFound, err.Error())
		return nil, st.Err()
	} else if err != nil {
		st := status.New(codes.Internal, fmt.Sprintf("Failed to get invoice. Reason: %v", err))
		log.WithError(err).Warn("GetInvoice request failed")
		return nil, st.Err()
	} else {
		res := xlnrpc.GetInvoiceResponse{
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

func (x xlnAdminServer) ListPendingInvoices(ctx context.Context, request *xlnrpc.ListPendingInvoicesRequest) (*xlnrpc.ListPendingInvoicesResponse, error) {
	log.WithField("req", request).Debug("XlnAdmin.ListPendingInvoices called")
	err := x.xln.AuthService.ValidateAdminCredentials(ctx, "XlnAdmin.ListPendingInvoices")
	if err != nil {
		st := status.New(codes.Unauthenticated, fmt.Sprintf("Invalid authentication for ListPendingInvoices. Reason: %v", err))
		log.WithError(err).Warn("ListPendingInvoices request failed authentication")
		return nil, st.Err()
	}
	pendingInvoices, err := x.xln.PendingInvoices.ListPendingInvoices()
	if err != nil {
		st := status.New(codes.Internal, fmt.Sprintf("Failed to list pending invoices. Reason: %v", err))
		log.WithError(err).Warn("ListPendingInvoices request failed")
		return nil, st.Err()
	} else {
		var res xlnrpc.ListPendingInvoicesResponse
		for _, invoice := range pendingInvoices {
			res.PendingInvoices = append(res.PendingInvoices, &xlnrpc.PendingInvoiceSummary{
				Amount:      invoice.Amount,
				CreatedAt:   timestamppb.New(invoice.CreatedAt),
				PaymentHash: invoice.PaymentHash,
			})
			res.TotalAmount += invoice.Amount
		}
		return &res, nil
	}
}

func (x xlnAdminServer) ListPendingPayments(ctx context.Context, request *xlnrpc.ListPendingPaymentsRequest) (*xlnrpc.ListPendingPaymentsResponse, error) {
	log.WithField("req", request).Debug("Xln.ListPendingPayments called")
	err := x.xln.AuthService.ValidateAdminCredentials(ctx, "XlnAdmin.ListPendingPayments")
	if err != nil {
		st := status.New(codes.Unauthenticated, fmt.Sprintf("Invalid authentication for ListPendingPayments. Reason: %v", err))
		log.WithError(err).Warn("ListPendingPayments request failed authentication")
		return nil, st.Err()
	}
	pendingPayments, err := x.xln.PendingPayments.ListPendingPayments()
	if err != nil {
		st := status.New(codes.Internal, fmt.Sprintf("Failed to list pending invoices. Reason: %v", err))
		log.WithError(err).Warn("ListPendingPayments request failed")
		return nil, st.Err()
	} else {
		var res xlnrpc.ListPendingPaymentsResponse
		for _, payment := range pendingPayments {
			res.PendingPayments = append(res.PendingPayments, &xlnrpc.PaymentSummary{
				Amount:      payment.Amount,
				CreatedAt:   timestamppb.New(payment.CreatedAt),
				PaymentHash: payment.PaymentHash,
			})
			res.TotalAmount += payment.Amount
		}
		return &res, nil
	}
}
