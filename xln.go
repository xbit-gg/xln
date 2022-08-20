package xln

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	proxy "github.com/grpc-ecosystem/grpc-gateway/runtime"
	log "github.com/sirupsen/logrus"
	"github.com/xbit-gg/xln/auth"
	"github.com/xbit-gg/xln/cert"
	"github.com/xbit-gg/xln/cfg"
	"github.com/xbit-gg/xln/db"
	"github.com/xbit-gg/xln/lnd"
	lnAuth "github.com/xbit-gg/xln/lnurl/auth"
	"github.com/xbit-gg/xln/lnurl/withdraw"
	"github.com/xbit-gg/xln/resources/invoice"
	"github.com/xbit-gg/xln/resources/pendinginvoices"
	"github.com/xbit-gg/xln/resources/pendingpayments"
	"github.com/xbit-gg/xln/resources/user"
	"github.com/xbit-gg/xln/resources/wallet"
	"github.com/xbit-gg/xln/util"
	"github.com/xbit-gg/xln/xlnrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var version = "0.0.1"

// Acceptable HTTP headers
var xlnHeaders = map[string]struct{}{
	auth.UsernameHeader:     {},
	auth.AdminApiKeyHeader:  {},
	auth.WalletApiKeyHeader: {},
	auth.UserApiKeyHeader:   {},
}

// XLN stores the data associated with an XLN instance.
type XLN struct {
	Version string
	Config  *cfg.Config

	LndClient *lnd.Client
	DB        *db.DB

	Users           user.Manager
	Wallets         wallet.Manager
	Invoices        invoice.Manager
	PendingInvoices pendinginvoices.Manager
	PendingPayments pendingpayments.Manager
	LNURLAuths      lnAuth.Manager
	LNURLWithdraw   withdraw.Manager

	AuthService auth.Service
}

// NewXLN returns a new XLN initialized with config.
func NewXLN(config *cfg.Config) (*XLN, error) {
	xln := &XLN{Version: version}
	xln.Config = config
	log.SetLevel(config.LogLevel)
	log.Debug("Verbose debug logging enabled")

	if config.ShowVersion {
		log.Fatal("XLN version: ", version)
	}

	// Create data directory
	if !util.FileExists(config.XLNDir) {
		err := os.MkdirAll(config.XLNDir, 0700)
		if err != nil {
			return nil, fmt.Errorf("failed to create xln directory: %v", err)
		}
	}

	// Connect to LND
	xln.LndClient = lnd.NewClient()
	err := xln.LndClient.Connect(config.Lnd.Address, config.Lnd.TlsCert, config.Lnd.AdminMacaroon)
	if err != nil {
		log.Error("Unable to connect to LND. Ensure the [LND] section of the XLN config is correctly configured")
		return nil, fmt.Errorf("failed to connect to LND instance: %v", err)
	}

	// Init DB
	xln.DB, err = db.ConnectDB(config.DatabaseConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize DB: %v", err)
	}

	// Setup Managers
	xln.Users = user.NewManager(xln.DB)
	xln.Wallets = wallet.NewManager(xln.DB)
	xln.Invoices = invoice.NewManager(xln.LndClient, xln.Wallets, xln.DB, xln.Config.MaxPayment)
	xln.PendingInvoices = pendinginvoices.NewManager(xln.DB)
	xln.PendingPayments = pendingpayments.NewManager(xln.DB)
	xln.LNURLAuths = lnAuth.NewManager(xln.Config.Serving.Hostname, xln.DB)
	xln.LNURLWithdraw = withdraw.NewManager(xln.Config.Serving.Hostname, xln.DB)

	// Initialize Services
	xln.AuthService = auth.NewService(xln.Config.XLNApiKey, &xln.Users, &xln.Wallets)

	return xln, nil
}

// Serve starts serving the APIs that are configured to be enabled.
// The caller must persist the main thread for serving to be effective.
func (xln *XLN) Serve() {

	if xln.Config.Serving.Grpc {
		err := xln.startGrpc()
		if err != nil {
			log.WithError(err).Fatal("Failed to start gRPC")
		}
		if xln.Config.Serving.Rest {
			err := xln.startRestProxy()
			if err != nil {
				log.WithError(err).Fatal("Failed to start REST proxy")
			}
		}
	}
}

func (xln *XLN) startGrpc() error {
	opts, err := xln.grpcServeOptions()
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer(opts...)

	grpcPort := xln.Config.Serving.GrpcPort
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", grpcPort))
	if err != nil {
		return err
	}

	// XLN Servers
	xlnrpc.RegisterXlnServer(grpcServer, NewXlnServer(xln))
	xlnrpc.RegisterXlnAdminServer(grpcServer, NewXlnAdminServer(xln))
	xlnrpc.RegisterLNURLServer(grpcServer, NewLNURLServer(xln))

	go func(lis net.Listener) {
		log.WithField("port", grpcPort).Info("gRPC server listening")
		err = grpcServer.Serve(lis)
		if err != nil {
			log.WithError(err).Fatal("gRPC server failed to serve")
		}
	}(lis)
	return nil
}

func (xln *XLN) grpcServeOptions() ([]grpc.ServerOption, error) {
	var opts []grpc.ServerOption
	if xln.Config.Serving.Tls.EnableTls {
		opts = xln.getTLSOptions()
	}

	return opts, nil
}

func (xln *XLN) startRestProxy() error {
	ctx := context.Background()

	grpcPort := xln.Config.Serving.GrpcPort
	httpPort := xln.Config.Serving.RestPort

	// Register and start REST proxy
	var opts []grpc.DialOption
	if xln.Config.Serving.Tls.EnableTls {
		creds, err := credentials.NewClientTLSFromFile(xln.Config.Serving.Tls.CertPath, "")
		if err != nil {
			log.WithError(err).Fatal("Could not create TLS client")
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	mux := proxy.NewServeMux(proxyServeOptions()...)
	err := xlnrpc.RegisterXlnHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%d", grpcPort), opts)
	if err != nil {
		log.WithError(err).Fatal("Failed to register proxy Xln handler")
	}
	log.Debug("Registered XlnServer REST handler")
	err = xlnrpc.RegisterXlnAdminHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%d", grpcPort), opts)
	if err != nil {
		log.WithError(err).Fatal("Failed to register proxy Xln handler")
	}
	log.Debug("Registered XlnAdminServer REST handler")
	err = xlnrpc.RegisterLNURLHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%d", grpcPort), opts)
	if err != nil {
		log.WithError(err).Fatal("Failed to register proxy Xln handler")
	}
	log.Debug("Registered LNURLServer REST handler")

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", httpPort),
		Handler: allowCORS(mux, []string{"*"}),
	}
	if xln.Config.Serving.Tls.EnableTls {
		go func() {
			log.WithField("port", httpPort).Info("HTTPS server listening")
			err := srv.ListenAndServeTLS(xln.Config.Serving.Tls.CertPath, xln.Config.Serving.Tls.KeyPath)
			if err != nil {
				log.WithError(err).Fatal("HTTPS server failed to serve")
			}
		}()
	} else {
		go func() {
			log.WithField("port", httpPort).Info("HTTP server listening")
			err := srv.ListenAndServe()
			if err != nil {
				log.WithError(err).Fatal("HTTP server failed to serve")
			}
		}()
	}
	return nil
}

func proxyServeOptions() []proxy.ServeMuxOption {
	var opts []proxy.ServeMuxOption
	opts = append(opts, proxy.WithMarshalerOption(proxy.MIMEWildcard, &proxy.JSONPb{EmitDefaults: true}))
	opts = append(opts, proxy.WithIncomingHeaderMatcher(matchXlnHeaders))

	return opts
}

func matchXlnHeaders(header string) (string, bool) {
	header = strings.ToLower(header)
	_, contains := xlnHeaders[header]
	return header, contains
}

// allowCORS wraps the given http.Handler with a function that adds the
// Access-Control-Allow-Origin header to the response.
func allowCORS(handler http.Handler, origins []string) http.Handler {
	allowHeaders := "Access-Control-Allow-Headers"
	allowMethods := "Access-Control-Allow-Methods"
	allowOrigin := "Access-Control-Allow-Origin"

	// If the user didn't supply any origins that means CORS is disabled
	// and we should return the original handler.
	if len(origins) == 0 {
		return handler
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Skip everything if the browser doesn't send the Origin field.
		if origin == "" {
			handler.ServeHTTP(w, r)
			return
		}

		// Set the static header fields first.
		headers, i := make([]string, len(xlnHeaders)), 0
		for key := range xlnHeaders {
			headers[i] = key
			i++
		}
		headers = append(headers, "content-type")
		w.Header().Set(
			allowHeaders,
			strings.Join(headers, ", "),
		)
		w.Header().Set(allowMethods, "GET, POST, PATCH, DELETE")

		// Either we allow all origins or the incoming request matches
		// a specific origin in our list of allowed origins.
		for _, allowedOrigin := range origins {
			if allowedOrigin == "*" || origin == allowedOrigin {
				// Only set allowed origin to requested origin.
				w.Header().Set(allowOrigin, origin)

				break
			}
		}

		// For a pre-flight request we only need to send the headers
		// back. No need to call the rest of the chain.
		if r.Method == "OPTIONS" {
			return
		}

		// Everything's prepared now, we can pass the request along the
		// chain of handlers.
		handler.ServeHTTP(w, r)
	})
}

// getTLSConfig returns TLS gRPC server options
func (xln *XLN) getTLSOptions() []grpc.ServerOption {
	if !util.FileExists(xln.Config.Serving.Tls.CertPath) || !util.FileExists(xln.Config.Serving.Tls.KeyPath) {
		cert.GenCertPair(
			xln.Config.Serving.Tls.CertPath,
			xln.Config.Serving.Tls.KeyPath,
			time.Duration(time.Hour.Nanoseconds()*xln.Config.Serving.Tls.CertValidityHours),
		)
	}
	creds, err := credentials.NewServerTLSFromFile(
		xln.Config.Serving.Tls.CertPath, xln.Config.Serving.Tls.KeyPath,
	)

	_, x509cert, err := cert.LoadCert(xln.Config.Serving.Tls.CertPath, xln.Config.Serving.Tls.KeyPath)
	if x509cert == nil || err != nil || time.Now().UTC().After(x509cert.NotAfter) {
		err := os.Remove(xln.Config.Serving.Tls.CertPath)
		if err != nil {
			log.WithError(err).Fatal("Failed to load server cert and/or its corresponding private key")
		}
		err = os.Remove(xln.Config.Serving.Tls.KeyPath)
		if err != nil {
			log.WithError(err).Fatal("Failed to load server cert and/or its corresponding private key")
		}
		cert.GenCertPair(
			xln.Config.Serving.Tls.CertPath,
			xln.Config.Serving.Tls.KeyPath,
			time.Duration(time.Hour.Nanoseconds()*xln.Config.Serving.Tls.CertValidityHours),
		)
	}
	if err != nil {
		log.WithError(err).Fatal("Failed to load server cert and/or its corresponding private key")
	}
	serverOpts := []grpc.ServerOption{grpc.Creds(creds)}
	return serverOpts
}
