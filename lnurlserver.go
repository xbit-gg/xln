package xln

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/xbit-gg/xln/xlnrpc"
)

const (
	OkStatus       = "OK"
	ErrorStatus    = "ERROR"
	TagWithdrawReq = "withdrawRequest"
	errorDetails   = "error details: %v"
)

var (
	SuccessResponse = &xlnrpc.LNURLResponse{Status: OkStatus}
)

type lnurlServer struct {
	xlnrpc.UnimplementedLNURLServer

	xln *XLN
}

func NewLNURLServer(xln *XLN) *lnurlServer {
	return &lnurlServer{xln: xln}
}

func (l lnurlServer) Auth(ctx context.Context, request *xlnrpc.AuthRequest) (*xlnrpc.LNURLResponse, error) {
	log.WithField("req", request).Debug("Xln.Auth called")
	err := l.xln.LNURLAuths.LNURLAuthenticate(request.K1, request.Sig, request.Key)
	if err != nil {
		return &xlnrpc.LNURLResponse{Status: ErrorStatus, Reason: err.Error()}, nil
	} else {
		return SuccessResponse, nil
	}
}

func (x lnurlServer) RequestWithdraw(ctx context.Context, request *xlnrpc.RequestWithdrawRequest) (*xlnrpc.RequestWithdrawResponse, error) {
	log.WithField("req", request).Debug("LNURL.RequestWithdraw called")
	w, callback, err := x.xln.LNURLWithdraw.GetWithdrawRequest(request.K1)
	if err != nil {
		log.WithError(err).Warn("RequestWithdraw request failed")
		res := xlnrpc.RequestWithdrawResponse{
			Status: ErrorStatus,
			Reason: fmt.Sprintf(errorDetails, err),
		}
		return &res, nil
	}

	res := xlnrpc.RequestWithdrawResponse{
		Status:             OkStatus,
		Tag:                TagWithdrawReq,
		Callback:           callback,
		DefaultDescription: w.Description,
		MinWithdrawable:    w.MinMsat,
		MaxWithdrawable:    w.MaxMsat,
		K1:                 w.K1,
	}

	return &res, nil
}

func (x lnurlServer) Withdraw(ctx context.Context, request *xlnrpc.WithdrawRequest) (*xlnrpc.LNURLResponse, error) {
	log.WithField("req", request).Debug("LNURL.Withdraw called")
	if request.K1 == "" {
		res := xlnrpc.LNURLResponse{
			Status: ErrorStatus,
			Reason: fmt.Sprintf(errorDetails, "must provide k1"),
		}
		return &res, nil
	} else if request.Pr == "" {
		res := xlnrpc.LNURLResponse{
			Status: ErrorStatus,
			Reason: fmt.Sprintf(errorDetails, "must provide payment request"),
		}
		return &res, nil
	}

	err := x.xln.Invoices.PayWithdrawInvoice(request.K1, request.Pr)
	if err != nil {
		log.WithError(err).Warn("Withdraw request failed")
		res := xlnrpc.LNURLResponse{
			Status: ErrorStatus,
			Reason: fmt.Sprintf(errorDetails, err),
		}
		return &res, nil
	}

	return SuccessResponse, nil
}
