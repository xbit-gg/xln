package xlntests

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/xbit-gg/xln"
	"github.com/xbit-gg/xln/auth"
	"github.com/xbit-gg/xln/cfg"
	"github.com/xbit-gg/xln/db"
	"github.com/xbit-gg/xln/models"
	"github.com/xbit-gg/xln/xlnrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func StartServer(c *cfg.Config) {
	go func() {
		if err := cfg.ValidateEnv(); err != nil {
			log.WithError(err).Fatal("Invalid environment configuration")
		}

		x, err := xln.NewXLN(c)
		if err != nil {
			log.WithError(err).Fatal("Failed to initialize XLN")
		}
		x.Serve()
	}()
}

func SetupSqlite(dbConnStr string) (sqlite *db.DB, err error) {
	if err = os.Remove(dbConnStr); err != nil {
		return nil, err
	}
	return db.ConnectDB(dbConnStr)
}

func SetupPostgres(dbConnStr string) (postgres *db.DB, err error) {
	postgres, err = db.ConnectDB(dbConnStr)
	if err != nil {
		return nil, err
	}

	// drop all tables
	if tables, err := postgres.Migrator().GetTables(); err != nil {
		return nil, err
	} else if len(tables) == 0 {
		return nil, nil
	}

	// recreate all tables
	err = postgres.Migrator().DropTable(&models.Auth{})
	if err != nil {
		return nil, err
	}

	err = postgres.Migrator().DropTable(&models.User{})
	if err != nil {
		return nil, err
	}
	err = postgres.Migrator().DropTable(&models.Wallet{})
	if err != nil {
		return nil, err
	}
	err = postgres.Migrator().DropTable("wallet_transactions") // the many2many table defined in wallet schema
	if err != nil {
		return nil, err
	}
	err = postgres.Migrator().DropTable(&models.Invoice{})
	if err != nil {
		return nil, err
	}
	err = postgres.Migrator().DropTable(&models.Transaction{})
	if err != nil {
		return nil, err
	}
	err = postgres.Migrator().DropTable(&models.PendingInvoice{})
	if err != nil {
		return nil, err
	}
	err = postgres.Migrator().DropTable(&models.PendingPayment{})
	if err != nil {
		return nil, err
	}
	err = postgres.Migrator().DropTable(&models.Withdraw{})
	if err != nil {
		return nil, err
	}

	if tables, err := postgres.Migrator().GetTables(); err != nil {
		return nil, err
	} else if len(tables) != 0 {
		return nil, fmt.Errorf("Tables were not completely dropped: %v", tables)
	}

	// recreate all tables
	postgres, err = db.ConnectDB(dbConnStr)
	if err != nil {
		return nil, err
	} else {
		return postgres, nil
	}

}

func CreateHealthyClient(c *cfg.Config) (xlnrpc.XlnClient, *grpc.ClientConn, error) {
	var opts []grpc.DialOption
	if c.Serving.Tls.EnableTls {
		var creds credentials.TransportCredentials
		var err error
		creds, err = LoadTLSCredentials(c)
		if err != nil {
			log.WithError(err).Error("failed to load tls credentials")
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	connection, err := grpc.Dial(fmt.Sprintf("localhost:%d", c.Serving.GrpcPort), opts...)
	if err != nil {
		log.WithError(err).Error("failed to dial")
	}
	client := xlnrpc.NewXlnClient(connection)
	for i := 0; i < 5; i++ {
		header := metadata.New(map[string]string{auth.AdminApiKeyHeader: c.XLNApiKey})
		ctx := metadata.NewOutgoingContext(context.Background(), header)
		_, err = client.GetInfo(ctx, &xlnrpc.GetInfoRequest{})
		if err != nil {
			sleepDur := (time.Second / 3) * time.Duration(i+1)
			log.Warnf("sleeping for %v ns. Retrying", sleepDur)
			time.Sleep(sleepDur)
		} else {
			return client, connection, nil
		}
	}
	log.WithError(err).Error("Could not connect to server")
	return nil, nil, err
}

func CreateHealthyAdminClient(c *cfg.Config) (xlnrpc.XlnAdminClient, *grpc.ClientConn, error) {
	var opts []grpc.DialOption
	if c.Serving.Tls.EnableTls {
		var creds credentials.TransportCredentials
		var err error
		creds, err = LoadTLSCredentials(c)
		if err != nil {
			log.WithError(err).Error("failed to load tls credentials")
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	connection, err := grpc.Dial(fmt.Sprintf("localhost:%d", c.Serving.GrpcPort), opts...)
	if err != nil {
		log.WithError(err).Error("failed to dial")
	}
	client := xlnrpc.NewXlnAdminClient(connection)
	for i := 0; i < 5; i++ {
		header := metadata.New(map[string]string{auth.AdminApiKeyHeader: c.XLNApiKey})
		ctx := metadata.NewOutgoingContext(context.Background(), header)
		_, err = client.GetInfo(ctx, &xlnrpc.GetAdminInfoRequest{})
		if err != nil {
			sleepDur := (time.Second / 3) * time.Duration(i+1)
			log.Warnf("sleeping for %v ns. Retrying", sleepDur)
			time.Sleep(sleepDur)
		} else {
			return client, connection, nil
		}
	}
	log.WithError(err).Error("Could not connect to admin server")
	return nil, nil, err
}

func LoadTLSCredentials(c *cfg.Config) (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed server's certificate
	// retry if start server go routine is not done
	var (
		pemServerCA []byte
		err         error
	)
	for i := 0; i < 5; i++ {
		pemServerCA, err = ioutil.ReadFile(c.Serving.Tls.CertPath)
		if err != nil {
			return nil, err
		}
		if err != nil {
			sleepDur := (time.Second / 3) * time.Duration(i+1)
			log.Warn("cert not generated yet. Retrying...")
			time.Sleep(sleepDur)
		}
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	// Create the credentials and return it
	config := &tls.Config{
		RootCAs: certPool,
	}

	return credentials.NewTLS(config), nil
}
