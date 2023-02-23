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
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/alecthomas/kingpin"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
	"sigs.k8s.io/external-dns/provider"
)

// Version is the current version of the app, generated at build time
var Version = "unknown"

type Config struct {
	CoreDNSConfig
	dryRun    bool
	logFormat string
	LogLevel  string
}

type CoreDNSPlugin struct {
	provider provider.Provider
}

type PropertyValuesEqualsRequest struct {
	Name     string `json:"name"`
	Previous string `json:"previous"`
	Current  string `json:"current"`
}

type PropertiesValuesEqualsResponse struct {
	Equals bool `json:"equals"`
}

func (p *CoreDNSPlugin) records(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet { // records
		log.Println("get records")
		records, err := p.provider.Records(context.Background())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(records)
		return
	} else if req.Method == http.MethodPost { // applychanges
		log.Println("post applychanges")
		// extract changes from the request body
		var changes plan.Changes
		if err := json.NewDecoder(req.Body).Decode(&changes); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err := p.provider.ApplyChanges(context.Background(), &changes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	log.Println("records: this should never happen")
}

func (p *CoreDNSPlugin) propertyValuesEquals(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet { // propertyValuesEquals
		log.Println("get propertyValuesEquals")
		pve := PropertyValuesEqualsRequest{}
		if err := json.NewDecoder(req.Body).Decode(&pve); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		b := p.provider.PropertyValuesEqual(pve.Name, pve.Previous, pve.Current)
		r := PropertiesValuesEqualsResponse{
			Equals: b,
		}
		out, _ := json.Marshal(&r)
		w.Write(out)
		return
	}
	log.Println("propertyValuesEquals: this should never happen")
}

func (p *CoreDNSPlugin) adjustEndpoints(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet { // adjustEndpoints
		log.Println("get adjustEndpoints")
		var pve []*endpoint.Endpoint
		if err := json.NewDecoder(req.Body).Decode(&pve); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		pve = p.provider.AdjustEndpoints(pve)
		out, _ := json.Marshal(&pve)
		w.Write(out)
		return
	}
	log.Println("adjustEndpoints: this should never happen")
}

func (p *CoreDNSPlugin) root(w http.ResponseWriter, req *http.Request) {
	header := w.Header()
	header.Set("Vary", "Content-Type")
	header.Set("Content-Type", "application/external.dns.plugin+json;version=1")
	w.WriteHeader(200)
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
	app := kingpin.New("external-dns-coredns-plugin", "ExternalDNS CoreDNS plugin")
	app.Version(Version)
	app.DefaultEnvars()
	app.Flag("dry-run", "When enabled, prints DNS record changes rather than actually performing them (default: disabled)").
		BoolVar(&cfg.dryRun)
	app.Flag("log-format", "The format in which log messages are printed (default: text, options: text, json)").
		Default("text").EnumVar(&cfg.logFormat, "text", "json")
	app.Flag("log-level", "Set the level of logging. (default: info, options: panic, debug, info, warning, error, fatal").
		Default("info").EnumVar(&cfg.LogLevel, allLogLevelsAsStrings()...)

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
	log.Infof("config: %s", cfg)

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
	log.Info("start ExternalDNS coreDNS plugin")
	p := CoreDNSPlugin{
		provider: dnsProvider,
	}

	m := http.NewServeMux()
	m.HandleFunc("/", p.root)
	m.HandleFunc("/records", p.records)
	m.HandleFunc("/propertyvaluesequals", p.propertyValuesEquals)
	m.HandleFunc("/adjustendpoints", p.adjustEndpoints)
	if err := http.ListenAndServe(":8888", m); err != nil {
		log.Fatalf("listen failed error: %v", err)
	}
}
