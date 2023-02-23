package main

import (
	"context"
	"encoding/json"
	"github.com/alecthomas/kingpin"
	"log"
	"net/http"
	"os"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
	"sigs.k8s.io/external-dns/provider"
)

// Version is the current version of the app, generated at build time
var Version = "unknown"

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

func (p *CoreDNSPlugin) providerHandler(w http.ResponseWriter, req *http.Request) {
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
	log.Println("this should never happen")
}

func (p *CoreDNSPlugin) propertyValuesEquals(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet { // propertyValuesEquals
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
	}

}

func (p *CoreDNSPlugin) adjustEndpoints(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet { // propertyValuesEquals
		var pve []*endpoint.Endpoint
		if err := json.NewDecoder(req.Body).Decode(&pve); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		pve = p.provider.AdjustEndpoints(pve)
		out, _ := json.Marshal(&pve)
		w.Write(out)
	}

}

func (cfg *CoreDNSConfig) ParseFlags(args []string) error {
	app := kingpin.New("external-dns-coredns-plugin", "ExternalDNS CoreDNS plugin")
	app.Version(Version)
	app.DefaultEnvars()
	app.Flag("prefix", "Specify the prefix name").
		Default("/skydns/").StringVar(&cfg.coreDNSPrefix)
	app.Flag("dry-run", "When enabled, prints DNS record changes rather than actually performing them (default: disabled)").
		BoolVar(&cfg.dryRun)
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
	cfg := CoreDNSConfig{
		domainFilter: endpoint.DomainFilter{},
	}
	if err := cfg.ParseFlags(os.Args[1:]); err != nil {
		log.Fatalf("flag parsing error: %v", err)
	}

	// instantiate the dns provider
	provider, err := NewCoreDNSProvider(cfg)
	if err != nil {
		panic(err)
	}

	p := CoreDNSPlugin{
		provider: provider,
	}

	m := http.NewServeMux()
	m.HandleFunc("/records", p.providerHandler)
	m.HandleFunc("/propertyvaluesequals", p.propertyValuesEquals)
	m.HandleFunc("/adjustendpoints", p.adjustEndpoints)
	if err := http.ListenAndServe(":8888", m); err != nil {
		log.Fatal(err)
	}
}
