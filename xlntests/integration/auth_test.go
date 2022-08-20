package integration

import (
	"context"

	"github.com/xbit-gg/xln/auth"
	"github.com/xbit-gg/xln/xlnrpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

func (s *integrationSuite) TestWhenMissingHeaderFields() {
	ctx1 := metadata.NewOutgoingContext(
		context.Background(),
		metadata.New(map[string]string{}))
	s.testUserLvlMethodsUnauthorized(ctx1, codes.InvalidArgument.String(), "nonexistent")
	s.testWalletLvlMethodsUnauthorized(ctx1, codes.InvalidArgument.String(), "nonexistent")

	ctx2 := metadata.NewOutgoingContext(
		context.Background(),
		metadata.New(map[string]string{
			auth.AdminApiKeyHeader: "invalid-api-key",
		}))
	s.testUserLvlMethodsUnauthorized(ctx2, codes.InvalidArgument.String(), "nonexistent")
	s.testWalletLvlMethodsUnauthorized(ctx2, codes.InvalidArgument.String(), "nonexistent")

	ctx3 := metadata.NewOutgoingContext(
		context.Background(),
		metadata.New(map[string]string{
			auth.UsernameHeader: "nonexistent-user",
		}))
	s.testUserLvlMethodsUnauthorized(ctx3, codes.InvalidArgument.String(), "nonexistent")
	s.testWalletLvlMethodsUnauthorized(ctx3, codes.InvalidArgument.String(), "nonexistent")
}

func (s *integrationSuite) TestWhenAdminCredentialsAreInvalid() {
	s.Run("given invalid admin api key", func() {
		ctx := metadata.NewOutgoingContext(
			context.Background(),
			metadata.New(map[string]string{
				auth.AdminApiKeyHeader: "invalid-api-key",
				auth.UsernameHeader:    "nonexistent-user",
			}))
		s.testUserLvlMethodsUnauthorized(ctx, codes.Unauthenticated.String(), "nonexistent")
		s.testWalletLvlMethodsUnauthorized(ctx, codes.Unauthenticated.String(), "nonexistent")
	})
	s.Run("given valid admin api key but nonexistent user", func() {
		ctx := metadata.NewOutgoingContext(
			context.Background(),
			metadata.New(map[string]string{
				auth.AdminApiKeyHeader: s.config.XLNApiKey,
				auth.UsernameHeader:    "nonexistent-user",
			}))
		s.testUserLvlMethodsUnauthorized(ctx, codes.Unauthenticated.String(), "nonexistent")
		s.testWalletLvlMethodsUnauthorized(ctx, codes.Unauthenticated.String(), "nonexistent")
	})
}

func (s *integrationSuite) TestWhenUserCredentialsAreInvalid() {
	s.Run("given invalid user key and nonexistent user", func() {
		ctx := metadata.NewOutgoingContext(
			context.Background(),
			metadata.New(map[string]string{
				auth.UserApiKeyHeader: "invalid-user-key",
				auth.UsernameHeader:   "nonexistent-user",
			}))
		s.testUserLvlMethodsUnauthorized(ctx, codes.Unauthenticated.String(), "nonexistent")
		s.testWalletLvlMethodsUnauthorized(ctx, codes.Unauthenticated.String(), "nonexistent")
	})
	s.Run("given invalid user key on existing user", func() {
		user1 := "user1-user-auth"
		ctx := metadata.NewOutgoingContext(
			context.Background(),
			metadata.New(map[string]string{
				auth.UserApiKeyHeader: "invalid-user-key",
				auth.UsernameHeader:   user1,
			}))
		s.testUserLvlMethodsUnauthorized(ctx, codes.Unauthenticated.String(), user1)
		s.createUser(metadata.NewOutgoingContext(
			context.Background(), metadata.New(map[string]string{
				auth.AdminApiKeyHeader: s.config.XLNApiKey,
			})), user1)
		s.testUserLvlMethodsUnauthorized(ctx, codes.Unauthenticated.String(), user1)
		s.testWalletLvlMethodsUnauthorized(ctx, codes.Unauthenticated.String(), user1)
	})
}

func (s *integrationSuite) TestWhenWalletCredentialsAreInvalid() {
	s.Run("given invalid wallet key and nonexistent wallet", func() {
		ctx := metadata.NewOutgoingContext(
			context.Background(),
			metadata.New(map[string]string{
				auth.WalletApiKeyHeader: "invalid-user-key",
				auth.UsernameHeader:     "nonexistent-user",
			}))
		s.testUserLvlMethodsUnauthorized(ctx, codes.InvalidArgument.String(), "nonexistent")
		s.testWalletLvlMethodsUnauthorized(ctx, codes.Unauthenticated.String(), "nonexistent")
	})
	s.Run("given invalid wallet key on existing wallet", func() {
		user1 := "user1-wallet-auth"
		ctx := metadata.NewOutgoingContext(
			context.Background(),
			metadata.New(map[string]string{
				auth.WalletApiKeyHeader: "invalid-wallet-key",
				auth.UsernameHeader:     user1,
			}))
		s.testUserLvlMethodsUnauthorized(ctx, codes.InvalidArgument.String(), user1)
		s.testWalletLvlMethodsUnauthorized(ctx, codes.Unauthenticated.String(), user1)
		s.createUser(metadata.NewOutgoingContext(
			context.Background(), metadata.New(map[string]string{
				auth.AdminApiKeyHeader: s.config.XLNApiKey,
			})), user1)
		s.testUserLvlMethodsUnauthorized(ctx, codes.InvalidArgument.String(), user1)
		s.testWalletLvlMethodsUnauthorized(ctx, codes.Unauthenticated.String(), user1)
	})
}

func (s *integrationSuite) testUserLvlMethodsUnauthorized(ctx context.Context, expectedCode, walletId string) {
	var usrErr error
	_, usrErr = s.client.CreateWallet(ctx, &xlnrpc.CreateWalletRequest{WalletId: walletId})
	s.Require().NotNil(usrErr)
	s.Require().Contains(usrErr.Error(), expectedCode)

	_, usrErr = s.client.GetUser(ctx, &xlnrpc.GetUserRequest{})
	s.Require().NotNil(usrErr)
	s.Require().Contains(usrErr.Error(), expectedCode)

	_, usrErr = s.client.UserLinkWallet(ctx, &xlnrpc.UserLinkWalletRequest{})
	s.Require().NotNil(usrErr)
	s.Require().Contains(usrErr.Error(), expectedCode)

	_, usrErr = s.client.ListUserTransactions(ctx, &xlnrpc.ListUserTransactionsRequest{})
	s.Require().NotNil(usrErr)
	s.Require().Contains(usrErr.Error(), expectedCode)

	_, usrErr = s.client.ListWallets(ctx, &xlnrpc.ListWalletsRequest{})
	s.Require().NotNil(usrErr)
	s.Require().Contains(usrErr.Error(), expectedCode)
}

func (s *integrationSuite) testWalletLvlMethodsUnauthorized(ctx context.Context, expectedCode, walletId string) {
	var walErr error
	_, walErr = s.client.CreateInvoice(ctx, &xlnrpc.CreateInvoiceRequest{WalletId: walletId, Value: 10000})
	s.Require().NotNil(walErr)
	s.Require().Contains(walErr.Error(), expectedCode)

	_, walErr = s.client.DeleteWallet(ctx, &xlnrpc.DeleteWalletRequest{WalletId: walletId})
	s.Require().NotNil(walErr)
	s.Require().Contains(walErr.Error(), expectedCode)

	_, walErr = s.client.GetWallet(ctx, &xlnrpc.GetWalletRequest{WalletId: walletId})
	s.Require().NotNil(walErr)
	s.Require().Contains(walErr.Error(), expectedCode)

	_, walErr = s.client.GetWalletInvoice(ctx, &xlnrpc.GetWalletInvoiceRequest{WalletId: walletId})
	s.Require().NotNil(walErr)
	s.Require().Contains(walErr.Error(), expectedCode)

	_, walErr = s.client.GetWalletTransaction(ctx, &xlnrpc.GetWalletTransactionRequest{WalletId: walletId})
	s.Require().NotNil(walErr)
	s.Require().Contains(walErr.Error(), expectedCode)

	_, walErr = s.client.LinkWallet(ctx, &xlnrpc.LinkWalletRequest{WalletId: walletId})
	s.Require().NotNil(walErr)
	s.Require().Contains(walErr.Error(), expectedCode)

	_, walErr = s.client.ListWalletInvoices(ctx, &xlnrpc.ListWalletInvoicesRequest{WalletId: walletId})
	s.Require().NotNil(walErr)
	s.Require().Contains(walErr.Error(), expectedCode)

	_, walErr = s.client.ListWalletPendingInvoices(ctx, &xlnrpc.ListWalletPendingInvoicesRequest{WalletId: walletId})
	s.Require().NotNil(walErr)
	s.Require().Contains(walErr.Error(), expectedCode)

	_, walErr = s.client.ListWalletPendingPayments(ctx, &xlnrpc.ListWalletPendingPaymentsRequest{WalletId: walletId})
	s.Require().NotNil(walErr)
	s.Require().Contains(walErr.Error(), expectedCode)

	_, walErr = s.client.ListWalletTransactions(ctx, &xlnrpc.ListWalletTransactionsRequest{WalletId: walletId})
	s.Require().NotNil(walErr)
	s.Require().Contains(walErr.Error(), expectedCode)

	_, walErr = s.client.PayInvoice(ctx, &xlnrpc.PayInvoiceRequest{WalletId: walletId})
	s.Require().NotNil(walErr)
	s.Require().Contains(walErr.Error(), expectedCode)

	_, walErr = s.client.PayInvoiceSync(ctx, &xlnrpc.PayInvoiceRequest{WalletId: walletId})
	s.Require().NotNil(walErr)
	s.Require().Contains(walErr.Error(), expectedCode)

	_, walErr = s.client.UpdateWalletOptions(ctx, &xlnrpc.UpdateWalletOptionsRequest{WalletId: walletId})
	s.Require().NotNil(walErr)
	s.Require().Contains(walErr.Error(), expectedCode)

	_, walErr = s.client.Transfer(ctx, &xlnrpc.TransferRequest{WalletId: walletId, ToWalletId: "nonexistent2"})
	s.Require().NotNil(walErr)
	s.Require().Contains(walErr.Error(), expectedCode)

	_, walErr = s.client.CreateLNURLW(ctx, &xlnrpc.CreateLNURLWRequest{WalletId: walletId})
	s.Require().NotNil(walErr)
	s.Require().Contains(walErr.Error(), expectedCode)

	_, walErr = s.client.GetLNURLW(ctx, &xlnrpc.GetLNURLWRequest{WalletId: walletId, K1: "k1"})
	s.Require().NotNil(walErr)
	s.Require().Contains(walErr.Error(), expectedCode)
}
