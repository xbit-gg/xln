package integration

import (
	"context"
	"fmt"
	"time"

	"github.com/go-test/deep"
	"github.com/xbit-gg/xln/auth"
	"github.com/xbit-gg/xln/models"
	"github.com/xbit-gg/xln/xlnrpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

func (s *integrationSuite) TestAsAdminCreateUserDeleteUser() {
	username := "integrationtestusername"

	// create user
	s.createUser(
		metadata.NewOutgoingContext(
			context.Background(), metadata.New(map[string]string{
				auth.AdminApiKeyHeader: s.config.XLNApiKey,
			})),
		username)

	// create duplicate user errors
	createUserRes, err := s.adminClient.CreateUser(
		metadata.NewOutgoingContext(
			context.Background(), metadata.New(map[string]string{
				auth.AdminApiKeyHeader: s.config.XLNApiKey,
			})),
		&xlnrpc.CreateUserRequest{Username: username})
	s.Require().Nil(createUserRes)
	s.Require().NotNil(err)
	s.Require().Contains(err.Error(), codes.AlreadyExists.String())

	ctx := metadata.NewOutgoingContext(
		context.Background(),
		metadata.New(map[string]string{
			auth.AdminApiKeyHeader: s.config.XLNApiKey,
			auth.UsernameHeader:    username,
		}))

	// list wallet
	listWalletsRes1, err := s.listWallets(ctx, false)
	s.Require().Nil(err, "should not error when listing wallets")
	s.Require().NotNil(listWalletsRes1, "list wallet response should not be empty")
	s.Require().Equal(1, len(listWalletsRes1.WalletIds), "by default one wallet id should be listed")
	s.Require().Equal(username, listWalletsRes1.WalletIds[0], "default wallet ID should be equal to user name")

	// get wallet
	wal1Res1, err := s.getWallet(ctx, username)
	s.Require().Nil(err, "should not error when getting default wallet")
	s.Require().NotNil(wal1Res1, "get wallet response should not be empty")
	s.Require().Equal(username, wal1Res1.WalletId, "default wallet's id should be user name")
	s.Require().Equal(username, wal1Res1.WalletName, "default wallet's name should be user name")
	s.Require().Equal(uint64(0), wal1Res1.Balance, "default wallet's balance should be initially 0")
	s.Require().Equal(false, wal1Res1.Locked, "default wallet's should not be initially locked")
	s.Require().WithinDuration(time.Now(), wal1Res1.CreationTime.AsTime(), time.Minute)
	s.Require().Empty(wal1Res1.LinkKey, "should initially not be linked to external wallet")
	s.Require().Empty(wal1Res1.LinkLabel, "should initially not be linked to external wallet or given link label")

	// get user txns
	noTxnsRes, err := s.client.ListUserTransactions(ctx, &xlnrpc.ListUserTransactionsRequest{})
	s.Require().Nil(err, "should not error when getting user txns")
	s.Require().NotNil(noTxnsRes, "get user txns response should not be empty")
	s.Require().Zero(noTxnsRes.TotalRecords, "new user should have no pre-existing txns")
	s.Require().Zero(len(noTxnsRes.Transactions), "new user should have no pre-existing txns")
	s.Require().Zero(noTxnsRes.NextOffset, "next offset should be 0 because it is not pagination")

	// delete nonexistent user
	deleteUserRes, err := s.adminClient.DeleteUser(
		metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			auth.AdminApiKeyHeader: s.config.XLNApiKey})),
		&xlnrpc.DeleteUserRequest{Username: fmt.Sprint(time.Now().UnixNano())})
	s.Require().Nil(deleteUserRes)
	s.Require().NotNil(err)

	// delete user
	s.softDeleteUser(
		metadata.NewOutgoingContext(
			context.Background(), metadata.New(map[string]string{
				auth.AdminApiKeyHeader: s.config.XLNApiKey,
			})),
		username)
}

func (s *integrationSuite) TestCreateWalletGetWalletListWalletAsAdmin() {
	username := "user-TestCreateWalletGetWalletListWallet"
	ctx := metadata.NewOutgoingContext(
		context.Background(), metadata.New(map[string]string{
			auth.AdminApiKeyHeader: s.config.XLNApiKey,
			auth.UsernameHeader:    username,
		}))

	// create user
	s.createUser(metadata.NewOutgoingContext(
		context.Background(), metadata.New(map[string]string{
			auth.AdminApiKeyHeader: s.config.XLNApiKey,
		})), username)
	s.testCreateWalletGetWalletListWallet(ctx, username)
	s.db.Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.User{})
}

func (s *integrationSuite) TestCreateWalletGetWalletListWalletAsUser() {
	username := "user-TestCreateWalletGetWalletListWallet"
	// create user
	userRes, _ := s.createUser(metadata.NewOutgoingContext(
		context.Background(), metadata.New(map[string]string{
			auth.AdminApiKeyHeader: s.config.XLNApiKey,
		})), username)
	ctx := metadata.NewOutgoingContext(
		context.Background(), metadata.New(map[string]string{
			auth.UserApiKeyHeader: userRes.ApiKey,
			auth.UsernameHeader:   username,
		}))

	s.testCreateWalletGetWalletListWallet(ctx, username)
}

func (s *integrationSuite) TestInvoicesAsAdmin() {
	u1 := "user-TestAsAdminInvoices-1"
	u2 := "user-TestAsAdminInvoices-2"
	u1w2 := "user-TestAsAdminInvoices-1-wallet-2"
	u2w2 := "user-TestAsAdminInvoices-2-wallet-2"
	adminCtx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		auth.AdminApiKeyHeader: s.config.XLNApiKey}))
	// create user
	s.createUser(adminCtx, u1)
	s.createUser(adminCtx, u2)
	ctxU1 := metadata.NewOutgoingContext(
		context.Background(),
		metadata.New(map[string]string{
			auth.AdminApiKeyHeader: s.config.XLNApiKey,
			auth.UsernameHeader:    u1,
		}))
	ctxU2 := metadata.NewOutgoingContext(
		context.Background(),
		metadata.New(map[string]string{
			auth.AdminApiKeyHeader: s.config.XLNApiKey,
			auth.UsernameHeader:    u2,
		}))
	s.createWallet(ctxU1, u1w2)
	s.createWallet(ctxU2, u2w2)
	s.testInvoices(ctxU1, ctxU1, ctxU2, ctxU2, u1, u2, u1w2, u2w2)
}

func (s *integrationSuite) TestInvoicesAsUser() {
	u1 := "user-TestAsAdminInvoices-1"
	u2 := "user-TestAsAdminInvoices-2"
	u1w2 := "user-TestAsAdminInvoices-1-wallet-2"
	u2w2 := "user-TestAsAdminInvoices-2-wallet-2"
	adminCtx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		auth.AdminApiKeyHeader: s.config.XLNApiKey}))
	u1Res, _ := s.createUser(adminCtx, u1)
	u2Res, _ := s.createUser(adminCtx, u2)
	ctxU1 := metadata.NewOutgoingContext(
		context.Background(),
		metadata.New(map[string]string{
			auth.UserApiKeyHeader: u1Res.ApiKey,
			auth.UsernameHeader:   u1,
		}))
	ctxU2 := metadata.NewOutgoingContext(
		context.Background(),
		metadata.New(map[string]string{
			auth.UserApiKeyHeader: u2Res.ApiKey,
			auth.UsernameHeader:   u2,
		}))
	s.createWallet(ctxU1, u1w2)
	s.createWallet(ctxU2, u2w2)
	s.testInvoices(ctxU1, ctxU1, ctxU2, ctxU2, u1, u2, u1w2, u2w2)
}

func (s *integrationSuite) TestInvoicesAsWallet() {
	u1 := "user-TestAsAdminInvoices-1"
	u2 := "user-TestAsAdminInvoices-2"
	u1w2 := "user-TestAsAdminInvoices-1-wallet-2"
	u2w2 := "user-TestAsAdminInvoices-2-wallet-2"
	adminCtx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		auth.AdminApiKeyHeader: s.config.XLNApiKey}))
	s.createUser(adminCtx, u1)
	s.createUser(adminCtx, u2)

	adminCtxU1 := metadata.NewOutgoingContext(
		context.Background(),
		metadata.New(map[string]string{
			auth.AdminApiKeyHeader: s.config.XLNApiKey,
			auth.UsernameHeader:    u1,
		}))
	adminCtxU2 := metadata.NewOutgoingContext(
		context.Background(),
		metadata.New(map[string]string{
			auth.AdminApiKeyHeader: s.config.XLNApiKey,
			auth.UsernameHeader:    u2,
		}))
	u1w1res, _ := s.getWallet(adminCtxU1, u1)
	u1w2res, _ := s.createWallet(adminCtxU1, u1w2)
	u2w1res, _ := s.getWallet(adminCtxU2, u2)
	u2w2res, _ := s.createWallet(adminCtxU2, u2w2)

	ctxU1W1 := metadata.NewOutgoingContext(
		context.Background(),
		metadata.New(map[string]string{
			auth.WalletApiKeyHeader: u1w1res.ApiKey,
			auth.UsernameHeader:     u1,
		}))
	ctxU1W2 := metadata.NewOutgoingContext(
		context.Background(),
		metadata.New(map[string]string{
			auth.WalletApiKeyHeader: u1w2res.ApiKey,
			auth.UsernameHeader:     u1,
		}))
	ctxU2W1 := metadata.NewOutgoingContext(
		context.Background(),
		metadata.New(map[string]string{
			auth.WalletApiKeyHeader: u2w1res.ApiKey,
			auth.UsernameHeader:     u2,
		}))
	ctxU2W2 := metadata.NewOutgoingContext(
		context.Background(),
		metadata.New(map[string]string{
			auth.WalletApiKeyHeader: u2w2res.ApiKey,
			auth.UsernameHeader:     u2,
		}))
	s.testInvoices(ctxU1W1, ctxU1W2, ctxU2W1, ctxU2W2, u1, u2, u1w2, u2w2)
}

func (s *integrationSuite) testCreateWalletGetWalletListWallet(ctx context.Context, username string) {
	// create wallet
	s.createWallet(ctx, "wallet1")

	// get wallet
	getWal1Res1, err := s.getWallet(ctx, "wallet1")
	s.Require().Nil(err)
	s.Require().NotNil(getWal1Res1)
	s.Require().Equal("wallet1", getWal1Res1.WalletId)
	s.Require().Equal("wallet1", getWal1Res1.WalletName)
	s.Require().Equal(uint64(0), getWal1Res1.Balance, "default wallet's balance should be initially 0")
	s.Require().Equal(false, getWal1Res1.Locked, "default wallet's should not be initially locked")
	s.Require().WithinDuration(time.Now(), getWal1Res1.CreationTime.AsTime(), time.Minute)
	s.Require().NotEmpty(getWal1Res1.ApiKey)
	s.Require().Equal(44, len(getWal1Res1.ApiKey))
	s.Require().IsType("string", getWal1Res1.ApiKey)

	// list wallets
	listWalRes1, err := s.listWallets(ctx, true)
	s.Require().Nil(err)
	s.Require().NotNil(listWalRes1)
	s.Require().Nil(deep.Equal([]string{username, "wallet1"}, listWalRes1.WalletIds))
	s.Require().Equal(username, listWalRes1.Data[0].Id)
	s.Require().Equal("wallet1", listWalRes1.Data[1].Id)

	// create wallet
	s.createWalletWithName(ctx, "wallet2", "wallet2-name")

	// get wallet
	getWal2Res1, err := s.getWallet(ctx, "wallet2")
	s.Require().Nil(err)
	s.Require().NotNil(getWal2Res1)
	s.Require().Equal("wallet2", getWal2Res1.WalletId)
	s.Require().Equal("wallet2-name", getWal2Res1.WalletName)
	s.Require().Equal(uint64(0), getWal2Res1.Balance, "default wallet's balance should be initially 0")
	s.Require().Equal(false, getWal2Res1.Locked, "default wallet's should not be initially locked")
	s.Require().WithinDuration(time.Now(), getWal2Res1.CreationTime.AsTime(), time.Minute)

	// list wallets
	listWalRes2, err := s.listWallets(ctx, false)
	s.Require().Nil(err)
	s.Require().NotNil(listWalRes2)
	s.Require().Equal(3, len(listWalRes2.WalletIds))
	s.Require().Contains(listWalRes2.WalletIds, "wallet1")
	s.Require().Contains(listWalRes2.WalletIds, "wallet2")
	s.Require().Contains(listWalRes2.WalletIds, username)

	// delete wallet
	deleteWalletRes, err := s.deleteWallet(ctx, "wallet2")
	s.Require().Nil(err)
	s.Require().NotNil(deleteWalletRes)

	// list wallets
	listWalRes3, err := s.listWallets(ctx, false)
	s.Require().Nil(err)
	s.Require().NotNil(listWalRes3)
	s.Require().Equal(2, len(listWalRes3.WalletIds))
	s.Require().Contains(listWalRes3.WalletIds, "wallet1")
	s.Require().Contains(listWalRes3.WalletIds, username)

	// create wallet fails when already exists
	duplicateWalletRes, err := s.client.CreateWallet(ctx, &xlnrpc.CreateWalletRequest{WalletId: "wallet1"})
	s.Require().NotNil(err, "duplicate wallet creation should result in error")
	s.Require().Contains(err.Error(), codes.AlreadyExists.String())
	s.Require().Nil(duplicateWalletRes)

	s.softDeleteUser(metadata.NewOutgoingContext(
		context.Background(), metadata.New(map[string]string{
			auth.AdminApiKeyHeader: s.config.XLNApiKey,
		})), username)
}

func (s *integrationSuite) testInvoices(ctxU1W1, ctxU1W2, ctxU2W1, ctxU2W2 context.Context, u1, u2, u1w2, u2w2 string) {
	adminCtxt := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		auth.AdminApiKeyHeader: s.config.XLNApiKey}))

	// manually change user 2 wallet 2 balance
	_, err := s.adminClient.UpdateWallet(adminCtxt,
		&xlnrpc.UpdateWalletRequest{Username: u2, WalletId: u2w2, Balance: 5000, UpdateBalance: true})
	s.Require().Nil(err, err)
	u2w2data, err := s.client.GetWallet(ctxU2W2, &xlnrpc.GetWalletRequest{WalletId: u2w2})
	s.Require().Nil(err, err)
	s.Require().Equal(uint64(5000), u2w2data.Balance)

	// create invoice on user 1 wallet 2
	createInvc1Res, err := s.client.CreateInvoice(ctxU1W2, &xlnrpc.CreateInvoiceRequest{WalletId: u1w2, Value: 5000, Memo: "test-invoice"})
	s.Require().Nil(err)
	s.Require().NotNil(createInvc1Res)

	// ensure records of invoice are created and return data as expected
	invc1, err := s.client.GetWalletInvoice(ctxU1W2, &xlnrpc.GetWalletInvoiceRequest{WalletId: u1w2, PaymentHash: createInvc1Res.PaymentHash})
	s.Require().Nil(err)
	s.Require().NotNil(invc1)
	s.Require().Equal(createInvc1Res.PaymentHash, invc1.PaymentHash)
	s.Require().Equal(u1w2, invc1.RecipientId)
	s.Require().Equal(u1, invc1.RecipientUsername)
	s.Require().Empty(invc1.SenderId)
	s.Require().Empty(invc1.SenderUsername)
	s.Require().Equal(uint64(5000), invc1.Amount)

	pndInvcsUsr1, err := s.client.ListWalletPendingInvoices(ctxU1W2, &xlnrpc.ListWalletPendingInvoicesRequest{WalletId: u1w2})
	s.Require().Nil(err)
	s.Require().NotNil(pndInvcsUsr1)
	s.Require().Equal(1, len(pndInvcsUsr1.PendingInvoices))
	s.Require().Equal(uint64(5000), pndInvcsUsr1.TotalAmount)
	s.Require().Equal(invc1.PaymentHash, pndInvcsUsr1.PendingInvoices[0].PaymentHash)
	s.Require().Equal(uint64(5000), pndInvcsUsr1.PendingInvoices[0].Amount)
	invc1CreatedAt := pndInvcsUsr1.PendingInvoices[0].CreatedAt.AsTime()
	s.Require().True(time.Now().Add(-time.Second).Before(invc1CreatedAt))
	s.Require().True(time.Now().After(invc1CreatedAt))

	invc1Admin, err := s.adminClient.GetInvoice(adminCtxt,
		&xlnrpc.GetInvoiceRequest{PaymentHash: createInvc1Res.PaymentHash})
	s.Require().Nil(err)
	s.Require().NotNil(invc1Admin)
	s.Require().Equal(invc1.Amount, invc1Admin.Amount)
	s.Require().Equal(invc1.Memo, invc1Admin.Memo)
	s.Require().Equal(invc1.PaymentHash, invc1Admin.PaymentHash)
	s.Require().Equal(invc1.PaymentRequest, invc1Admin.PaymentRequest)
	s.Require().Equal(invc1.Preimage, invc1Admin.Preimage)
	s.Require().Equal(invc1.Pubkey, invc1Admin.Pubkey)
	s.Require().Equal(invc1.RecipientId, invc1Admin.RecipientId)
	s.Require().Equal(invc1.RecipientUsername, invc1Admin.RecipientUsername)
	s.Require().Equal(invc1.SenderId, invc1Admin.SenderId)
	s.Require().Equal(invc1.SenderUsername, invc1Admin.SenderUsername)
	s.Require().Equal(invc1.SettledAt, invc1Admin.SettledAt)
	s.Require().Equal(invc1.Timestamp, invc1Admin.Timestamp)

	pndInvc1Admin, err := s.adminClient.ListPendingInvoices(adminCtxt, &xlnrpc.ListPendingInvoicesRequest{})
	s.Require().Nil(err, err)
	s.Require().NotNil(pndInvc1Admin)
	s.Require().Equal(1, len(pndInvc1Admin.PendingInvoices))
	s.Require().Equal(uint64(5000), pndInvc1Admin.TotalAmount, pndInvc1Admin.TotalAmount)
	s.Require().Equal(invc1.PaymentHash, pndInvc1Admin.PendingInvoices[0].PaymentHash)
	s.Require().Equal(uint64(5000), pndInvc1Admin.PendingInvoices[0].Amount)
	invc1CreatedAtAdmin := pndInvc1Admin.PendingInvoices[0].CreatedAt.AsTime()
	s.Require().True(time.Now().Add(-time.Second).Before(invc1CreatedAtAdmin))
	s.Require().True(time.Now().After(invc1CreatedAtAdmin))

	invcsUsr1W2, err := s.client.ListWalletInvoices(ctxU1W2, &xlnrpc.ListWalletInvoicesRequest{WalletId: u1w2})
	s.Require().Nil(err)
	s.Require().NotNil(invcsUsr1W2)
	s.Require().Equal(1, len(invcsUsr1W2.Invoices))
	s.Require().Equal(invc1.PaymentHash, invcsUsr1W2.Invoices[0])

	// create 2nd invoice and confirm invoices are recorded properly
	createInv2, err := s.client.CreateInvoice(ctxU2W2, &xlnrpc.CreateInvoiceRequest{WalletId: u2w2, Value: 4000, Memo: "test-invoice"})
	s.Require().Nil(err)
	s.Require().NotNil(createInv2)

	invc2, err := s.client.GetWalletInvoice(ctxU2W2, &xlnrpc.GetWalletInvoiceRequest{WalletId: u2w2, PaymentHash: createInv2.PaymentHash})
	s.Require().Nil(err)
	s.Require().NotNil(invc2)
	s.Require().Equal(createInv2.PaymentHash, invc2.PaymentHash)
	s.Require().Equal(u2w2, invc2.RecipientId)
	s.Require().Equal(u2, invc2.RecipientUsername)
	s.Require().Empty(invc2.SenderId)
	s.Require().Empty(invc2.SenderUsername)
	s.Require().Equal(uint64(4000), invc2.Amount)

	pndInvcsUsr2, err := s.client.ListWalletPendingInvoices(ctxU2W2, &xlnrpc.ListWalletPendingInvoicesRequest{WalletId: u2w2})
	s.Require().Nil(err)
	s.Require().NotNil(pndInvcsUsr2)
	s.Require().Equal(1, len(pndInvcsUsr2.PendingInvoices))
	s.Require().Equal(invc2.PaymentHash, pndInvcsUsr2.PendingInvoices[0].PaymentHash)
	s.Require().Equal(uint64(4000), pndInvcsUsr2.TotalAmount)

	invcsUsr2, err := s.client.ListWalletInvoices(ctxU2W2, &xlnrpc.ListWalletInvoicesRequest{WalletId: u2w2})
	s.Require().Nil(err)
	s.Require().NotNil(invcsUsr2)
	s.Require().Equal(1, len(invcsUsr2.Invoices))
	s.Require().Equal(invc2.PaymentHash, invcsUsr2.Invoices[0])

	pndInvcsAdmin, err := s.adminClient.ListPendingInvoices(adminCtxt, &xlnrpc.ListPendingInvoicesRequest{})
	s.Require().Nil(err)
	s.Require().NotNil(pndInvcsAdmin)
	s.Require().Equal(uint64(9000), pndInvcsAdmin.TotalAmount)
	s.Require().Equal(2, len(pndInvcsAdmin.PendingInvoices))
	s.Require().Equal(invc1.PaymentHash, pndInvcsAdmin.PendingInvoices[0].PaymentHash)
	s.Require().Equal(invc2.PaymentHash, pndInvcsAdmin.PendingInvoices[1].PaymentHash)

	// do invoice ln node self-payment
	s.Require().GreaterOrEqual(u2w2data.Balance, invc1.Amount)
	u2W2PaysU1W2, err := s.client.PayInvoiceSync(ctxU2W2, &xlnrpc.PayInvoiceRequest{WalletId: u2w2, PaymentRequest: invc1.PaymentRequest})
	s.Require().Nil(err)
	s.Require().NotNil(u2W2PaysU1W2)
	s.Require().True(u2W2PaysU1W2.Success)
	s.Require().Equal(invc1.Amount, u2W2PaysU1W2.Amount)
	s.Require().Zero(u2W2PaysU1W2.FeesPaid)
	s.Require().Empty(u2W2PaysU1W2.FailureReason)

	u2W2AfterPay, err := s.client.GetWallet(ctxU2W2, &xlnrpc.GetWalletRequest{WalletId: u2w2})
	s.Require().Nil(err)
	s.Require().NotNil(u2W2AfterPay)
	s.Require().Equal(uint64(5000)-invc1.Amount, u2W2AfterPay.Balance)
	u2W2AfterRcv, err := s.client.GetWallet(ctxU1W2, &xlnrpc.GetWalletRequest{WalletId: u1w2})
	s.Require().Nil(err)
	s.Require().NotNil(u2W2AfterRcv)
	s.Require().Equal(invc1.Amount, u2W2AfterRcv.Balance)

	getInvc1Paid, err := s.adminClient.GetInvoice(adminCtxt,
		&xlnrpc.GetInvoiceRequest{PaymentHash: createInvc1Res.PaymentHash})
	s.Require().Nil(err)
	s.Require().NotNil(getInvc1Paid)
	s.Require().Equal(invc1.Amount, getInvc1Paid.Amount)
	s.Require().Equal(invc1.Memo, getInvc1Paid.Memo)
	s.Require().Equal(invc1.PaymentHash, getInvc1Paid.PaymentHash)
	s.Require().Equal(invc1.PaymentRequest, getInvc1Paid.PaymentRequest)
	s.Require().Equal(invc1.Preimage, getInvc1Paid.Preimage)
	s.Require().Equal(invc1.Pubkey, getInvc1Paid.Pubkey)
	s.Require().Equal(invc1.RecipientId, getInvc1Paid.RecipientId)
	s.Require().Equal(invc1.RecipientUsername, getInvc1Paid.RecipientUsername)
	s.Require().Equal(u2w2, getInvc1Paid.SenderId)
	s.Require().Equal(u2, getInvc1Paid.SenderUsername)
	s.Require().Equal(invc1.Timestamp, getInvc1Paid.Timestamp)
	selfPaidSettledAt := getInvc1Paid.SettledAt.AsTime()
	s.Require().True(time.Now().Add(-time.Second).Before(selfPaidSettledAt))
	s.Require().True(time.Now().After(selfPaidSettledAt))

	u1w2Txns, err := s.client.ListWalletTransactions(ctxU1W2, &xlnrpc.ListWalletTransactionsRequest{WalletId: u1w2})
	s.Require().Nil(err)
	s.Require().NotNil(u1w2Txns)
	s.Require().Equal(1, len(u1w2Txns.Transactions))

	// delete user ensure corresponding invoices still visible
	s.softDeleteUser(adminCtxt, u1)
	getInvc1UsrDelete, err := s.adminClient.GetInvoice(adminCtxt,
		&xlnrpc.GetInvoiceRequest{PaymentHash: createInvc1Res.PaymentHash})
	s.Require().Nil(err)
	s.Require().NotNil(getInvc1UsrDelete)
	s.Require().Equal(invc1.Amount, getInvc1UsrDelete.Amount)
	s.Require().Equal(invc1.Memo, getInvc1UsrDelete.Memo)
	s.Require().Equal(invc1.PaymentHash, getInvc1UsrDelete.PaymentHash)
	s.Require().Equal(invc1.PaymentRequest, getInvc1UsrDelete.PaymentRequest)
	s.Require().Equal(invc1.Preimage, getInvc1UsrDelete.Preimage)
	s.Require().Equal(invc1.Pubkey, getInvc1UsrDelete.Pubkey)
	s.Require().Empty(getInvc1UsrDelete.RecipientUsername)
	s.Require().Empty(getInvc1UsrDelete.RecipientId)
	s.Require().Equal(getInvc1Paid.SenderId, getInvc1UsrDelete.SenderId)
	s.Require().Equal(getInvc1Paid.SenderUsername, getInvc1UsrDelete.SenderUsername)
	s.Require().Equal(getInvc1Paid.SettledAt, getInvc1UsrDelete.SettledAt)
	s.Require().Equal(invc1.Timestamp, getInvc1UsrDelete.Timestamp)

	delU1W2Txn, err := s.db.Repo.GetTransaction(s.db.DB, u1w2Txns.Transactions[0].Id)
	s.Require().Nil(err)
	s.Require().NotNil(delU1W2Txn)
	s.Require().Empty(delU1W2Txn.ToID)
	s.Require().Empty(delU1W2Txn.ToUsername)
	s.Require().Equal(u2w2, *delU1W2Txn.FromID)
	s.Require().Equal(u2, *delU1W2Txn.FromUsername)

	s.softDeleteUser(adminCtxt, u2)
}
