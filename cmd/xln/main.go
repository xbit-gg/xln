package main

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/xbit-gg/xln"
	"github.com/xbit-gg/xln/cfg"
	"github.com/xbit-gg/xln/util"
)

// main is the xln command entrypoint.
func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.Info("Starting XLN...")
	if err := cfg.ValidateEnv(); err != nil {
		log.WithError(err).Fatal("Invalid environment configuration")
	}
	conf := getConfig()

	x, err := xln.NewXLN(conf)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize XLN")
	}

	x.Serve()
	select {}
}

func getConfig() *cfg.Config {
	conf := cfg.DefaultConfig()

	// Create data directory
	if !util.FileExists(conf.XLNDir) {
		log.WithField("dir", conf.XLNDir).Info("XLN directory does not exist. Creating it")
		err := os.MkdirAll(conf.XLNDir, 0700)
		if err != nil {
			log.WithError(err).Fatal("Failed to create data directory")
		}
	}

	confPath := fmt.Sprintf("%s/xln.conf", conf.XLNDir)
	err := cfg.WriteConfig(conf, filepath.FromSlash(confPath), false)
	if err != nil {
		log.WithError(err).Fatal("Unable to write default config")
	}
	// Initial parse to get alternative config path
	conf, err = cfg.ParseConfigFlags(conf)
	if err != nil {
		log.WithError(err).Fatal("Failed to parse flags")
	}
	// Load config from file
	conf, err = cfg.LoadConfig(conf, conf.XLNConfig)
	if err != nil {
		log.WithError(err).Warn("Failed to load config from file")
	} else {
		log.WithField("path", conf.XLNConfig).Debug("Loaded config from file")
	}
	// Override file options with CLI options
	conf, err = cfg.ParseConfigFlags(conf)
	if err != nil {
		log.WithError(err).Fatal("Failed to parse flags")
	}

	return conf
}
