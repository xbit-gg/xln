package xln

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/xbit-gg/xln/lnd"
	"github.com/xbit-gg/xln/lnurl/withdraw"
	"github.com/xbit-gg/xln/models"
	"github.com/xbit-gg/xln/resources/invoice"
	"github.com/xbit-gg/xln/xlnrpc"
)

func TestLNURLServer(t *testing.T) {
	suite.Run(t, new(lnurlServerSuite))
}

type lnurlServerSuite struct {
	suite.Suite
	mockLnurlMgr  mockLnurlWithdrawManager
	mockInvMgr    mockInvoiceManager
	mockLndClient lnd.Client
	lnurlServer   lnurlServer
}

func (s *lnurlServerSuite) SetupSuite() {
	s.mockLnurlMgr = mockLnurlWithdrawManager{}
	s.mockInvMgr = mockInvoiceManager{}
	s.lnurlServer.xln = &XLN{Invoices: &s.mockInvMgr, LNURLWithdraw: &s.mockLnurlMgr}
}

func (s *lnurlServerSuite) TestRequestWithdrawGivesLNURLRPCErrorWhenErrors() {
	var (
		actualRes *xlnrpc.RequestWithdrawResponse
		err       error
	)
	expectedErr := errors.New("test error")
	expectedK1 := "test-k1"
	s.mockLnurlMgr.mockGetWithdrawRequest = func(actualK1 string) (*models.Withdraw, string, error) {
		s.Require().Equal(expectedK1, actualK1)
		return nil, "", expectedErr
	}
	actualRes, err = s.lnurlServer.RequestWithdraw(
		context.Background(),
		&xlnrpc.RequestWithdrawRequest{K1: expectedK1},
	)
	s.Require().Nil(err)
	s.Require().NotNil(actualRes)
	s.Require().Equal("ERROR", actualRes.Status)

	s.mockLnurlMgr.mockGetWithdrawRequest = func(actualK1 string) (*models.Withdraw, string, error) {
		s.Require().Equal(expectedK1, actualK1)
		return &models.Withdraw{}, "some-url", expectedErr
	}
	actualRes, err = s.lnurlServer.RequestWithdraw(
		context.Background(),
		&xlnrpc.RequestWithdrawRequest{K1: expectedK1},
	)
	s.Require().Nil(err)
	s.Require().NotNil(actualRes)
	s.Require().Equal("ERROR", actualRes.Status)
}

func (s *lnurlServerSuite) TestRequestWithdrawSucceedsWhenK1Exists() {
	expectedK1 := "test-k1"
	expectedCallback := "test-callback"
	expectedW := models.Withdraw{
		K1:          expectedK1,
		MinMsat:     10,
		MaxMsat:     20,
		Description: "test-description",
	}
	s.mockLnurlMgr.mockGetWithdrawRequest = func(actualK1 string) (*models.Withdraw, string, error) {
		s.Require().Equal(expectedK1, actualK1)
		return &expectedW, expectedCallback, nil
	}
	actualRes, err := s.lnurlServer.RequestWithdraw(
		context.Background(),
		&xlnrpc.RequestWithdrawRequest{K1: expectedK1},
	)
	s.Require().Nil(err)
	s.Require().NotNil(actualRes)
	s.Require().Equal("OK", actualRes.Status)
	s.Require().Equal("withdrawRequest", actualRes.Tag)
	s.Require().Equal(expectedCallback, actualRes.Callback)
	s.Require().Equal(expectedW.Description, actualRes.DefaultDescription)
	s.Require().Equal(expectedW.MaxMsat, actualRes.MaxWithdrawable)
	s.Require().Equal(expectedW.MinMsat, actualRes.MinWithdrawable)
	s.Require().Equal(expectedK1, actualRes.K1)
}

func (s *lnurlServerSuite) TestWithdrawGivesLNURLRPCErrorWhenParamsInvalid() {
	var (
		actualRes *xlnrpc.LNURLResponse
		err       error
	)

	testInputs := []*xlnrpc.WithdrawRequest{
		{K1: "", Pr: ""},
		{K1: "nonempty string", Pr: ""},
		{K1: "", Pr: "nonempty string"},
	}
	for _, input := range testInputs {
		actualRes, err = s.lnurlServer.Withdraw(context.Background(), input)
		s.Require().Nil(err)
		s.Require().NotNil(actualRes)
		s.Require().Equal("ERROR", actualRes.Status)
	}
}

func (s *lnurlServerSuite) TestWithdrawSucceedsWhenInvoiceMgrSucceeds() {
	expectedK1 := "test-k1"
	expectedPr := "test-payment-request"
	s.mockInvMgr.mockPayWithdrawInvoice = func(actualK1, actualPr string) error {
		s.Require().Equal(expectedK1, actualK1)
		s.Require().Equal(expectedPr, actualPr)
		return nil
	}
	actualRes, err := s.lnurlServer.Withdraw(
		context.Background(),
		&xlnrpc.WithdrawRequest{K1: expectedK1, Pr: expectedPr},
	)
	s.Require().Nil(err)
	s.Require().NotNil(actualRes)
	s.Require().Equal("OK", actualRes.Status)
	s.Require().Empty(actualRes.Reason)
}

func (s *lnurlServerSuite) TestWithdrawGivesLNURLRPCErrorWhenInvoiceMgrErrors() {
	expectedK1 := "test-k1"
	expectedPr := "test-payment-request"
	expectedErr := errors.New("expected-withdraw-error")
	s.mockInvMgr.mockPayWithdrawInvoice = func(actualK1, actualPr string) error {
		s.Require().Equal(expectedK1, actualK1)
		s.Require().Equal(expectedPr, actualPr)
		return expectedErr
	}
	actualRes, err := s.lnurlServer.Withdraw(
		context.Background(),
		&xlnrpc.WithdrawRequest{K1: expectedK1, Pr: expectedPr},
	)
	s.Require().Nil(err)
	s.Require().NotNil(actualRes)
	s.Require().Equal("ERROR", actualRes.Status)
	s.Require().NotEmpty(actualRes.Reason)
}

type mockLnurlWithdrawManager struct {
	withdraw.Manager

	mockGetWithdrawRequest func(k1 string) (*models.Withdraw, string, error)
}

func (m *mockLnurlWithdrawManager) GetWithdrawRequest(k1 string) (*models.Withdraw, string, error) {
	return m.mockGetWithdrawRequest(k1)
}

type mockInvoiceManager struct {
	invoice.Manager
	mockPayWithdrawInvoice func(k1, pr string) error
}

func (m *mockInvoiceManager) PayWithdrawInvoice(k1, pr string) error {
	return m.mockPayWithdrawInvoice(k1, pr)
}
