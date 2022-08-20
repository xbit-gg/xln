package cfg

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/lightningnetwork/lnd"
	log "github.com/sirupsen/logrus"
)

// Config contains the options used to initialize XLN.
type Config struct {
	XLNDir                   string    `long:"xlndir" description:"The base directory that contains xln's data, logs, config file, etc."`
	XLNConfig                string    `long:"config" description:"The path to the xln config file."`
	XLNApiKey                string    `long:"apikey" description:"The api key required to manage xln and xln users."`
	DatabaseConnectionString string    `long:"db" description:"The SQLite or PostgreSQL database connection string."`
	MaxPayment               int64     `long:"maxpayment" description:"The maximum payment size in millisatoshis."`
	ShowVersion              bool      `short:"v" long:"version" description:"Displays the version and then terminates."`
	LogLevel                 log.Level `long:"log" description:"Logrus log level."`

	Serving *Serving `group:"Serving" namespace:"serving"`
	Lnd     *Lnd     `group:"LND" namespace:"lnd"`
}

// DefaultConfig returns a Config populated with default options.
func DefaultConfig() *Config {
	restPort := uint16(5551)
	grpcPort := uint16(5550)
	return &Config{
		XLNDir:                   fmt.Sprintf("%s/.xln", os.Getenv("HOME")),
		XLNConfig:                fmt.Sprintf("%s/.xln/xln.conf", os.Getenv("HOME")),
		XLNApiKey:                "hodl",
		DatabaseConnectionString: fmt.Sprintf("%s/.xln/xln.db", os.Getenv("HOME")),
		MaxPayment:               500000 * 1000,
		ShowVersion:              false,
		LogLevel:                 log.InfoLevel,

		Serving: &Serving{
			Tls: &Tls{
				EnableTls:         true,
				KeyPath:           fmt.Sprintf("%s/.xln/server-key.pem", os.Getenv("HOME")),
				CertPath:          fmt.Sprintf("%s/.xln/server-cert.pem", os.Getenv("HOME")),
				CertValidityHours: 720,
			},
			Hostname: fmt.Sprintf("localhost:%v", restPort),
			Rest:     true,
			RestPort: restPort,
			Grpc:     true,
			GrpcPort: grpcPort,
		},
		Lnd: &Lnd{
			Address:       "127.0.0.1:10001",
			TlsCert:       fmt.Sprintf("%s/.polar/networks/1/volumes/lnd/alice/tls.cert", os.Getenv("HOME")),
			AdminMacaroon: fmt.Sprintf("%s/.polar/networks/1/volumes/lnd/alice/data/chain/bitcoin/regtest/admin.macaroon", os.Getenv("HOME")),
		},
	}
}

func CleanConfig(preCfg *Config) (*Config, error) {
	preCfg.XLNDir = lnd.CleanAndExpandPath(preCfg.XLNDir)
	preCfg.XLNConfig = lnd.CleanAndExpandPath(preCfg.XLNConfig)

	return preCfg, nil
}

func ParseConfigFlags(preCfg *Config) (*Config, error) {
	_, err := flags.Parse(preCfg)
	if err != nil {
		return nil, err
	}

	return CleanConfig(preCfg)
}

func LoadConfig(preCfg *Config, file string) (*Config, error) {
	err := flags.IniParse(file, preCfg)
	if err != nil {
		return nil, err
	}

	return CleanConfig(preCfg)
}

func WriteConfig(config *Config, file string, overwrite bool) error {
	p := flags.NewParser(config, flags.Default)
	iniParser := flags.NewIniParser(p)

	if overwrite {
		return iniParser.WriteFile(file, flags.IniIncludeDefaults)
	} else {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return iniParser.WriteFile(file, flags.IniIncludeDefaults)
		}
	}
	return nil
}
