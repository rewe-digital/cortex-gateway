package gateway

import (
	"flag"
)

// Config for a gateway
type Config struct {
	DistributorAddress   string
	QueryFrontendAddress string
}

// RegisterFlags adds the flags required to config this package's Config struct
func (cfg *Config) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&cfg.DistributorAddress, "gateway.distributor.address", "", "Upstream HTTP URL for Cortex Distributor")
	f.StringVar(&cfg.QueryFrontendAddress, "gateway.query-frontend.address", "", "Upstream HTTP URL for Cortex Query Frontend")
}
