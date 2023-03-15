/*
Copyright 2023 G DATA Software AG.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"os"
	"sigs.k8s.io/external-dns/provider/webhook"
	"time"

	"github.com/alecthomas/kingpin"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/external-dns/endpoint"
)

// Version is the current version of the app, generated at build time
var Version = "unknown"

type Config struct {
	CoreDNSConfig
	dryRun                      bool
	logFormat                   string
	LogLevel                    string
	webhookProviderReadTimeout  time.Duration
	webhookProviderWriteTimeout time.Duration
}

// allLogLevelsAsStrings returns all logrus levels as a list of strings
func allLogLevelsAsStrings() []string {
	var levels []string
	for _, level := range log.AllLevels {
		levels = append(levels, level.String())
	}
	return levels
}

func (cfg *Config) ParseFlags(args []string) error {
	app := kingpin.New("external-dns-coredns-webhook", "ExternalDNS CoreDNS webhook")
	app.Version(Version)
	app.DefaultEnvars()
	app.Flag("dry-run", "When enabled, prints DNS record changes rather than actually performing them (default: disabled)").
		BoolVar(&cfg.dryRun)
	app.Flag("log-format", "The format in which log messages are printed (default: text, options: text, json)").
		Default("text").EnumVar(&cfg.logFormat, "text", "json")
	app.Flag("log-level", "Set the level of logging. (default: info, options: panic, debug, info, warning, error, fatal").
		Default("info").EnumVar(&cfg.LogLevel, allLogLevelsAsStrings()...)

	app.Flag("webhook-provider-read-timeout", "The read timeout for the webhook provider in duration format (default: 5s)").
		Default((time.Second * 5).String()).DurationVar(&cfg.webhookProviderReadTimeout)
	app.Flag("webhook-provider-write-timeout", "The write timeout for the webhook provider in duration format (default: 5s)").
		Default((time.Second * 5).String()).DurationVar(&cfg.webhookProviderWriteTimeout)

	app.Flag("prefix", "Specify the prefix name").
		Default("/skydns/").StringVar(&cfg.coreDNSPrefix)
	app.Flag("txt-owner-id", "When using the TXT registry, a name that identifies this instance of ExternalDNS (default: default)").
		Default("default").StringVar(&cfg.ownerID)
	app.Flag("pre-filter-external-owned-records", "Services are pre filter based on the txt-owner-id (default: false)").
		BoolVar(&cfg.preFilterExternalOwnedRecords)

	_, err := app.Parse(args)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	// instantiate the config
	cfg := Config{
		CoreDNSConfig: CoreDNSConfig{
			domainFilter: endpoint.DomainFilter{},
		},
	}
	if err := cfg.ParseFlags(os.Args[1:]); err != nil {
		log.Fatalf("flag parsing error: %v", err)
	}
	log.Infof("config: %v", cfg)

	if cfg.logFormat == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	}
	if cfg.dryRun {
		log.Info("running in dry-run mode. No changes to DNS records will be made.")
	}

	ll, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to parse log level: %v", err)
	}
	log.SetLevel(ll)

	// instantiate the dns provider
	dnsProvider, err := NewCoreDNSProvider(cfg.CoreDNSConfig, cfg.dryRun)
	if err != nil {
		log.Fatalf("listen failed error: %v", err)
	}
	log.Info("start ExternalDNS coreDNS webhook")
	startedChan := make(chan struct{})
	go webhook.StartHTTPApi(dnsProvider, startedChan, cfg.webhookProviderReadTimeout, cfg.webhookProviderWriteTimeout, "0.0.0.0:8888")
	<-startedChan
}
