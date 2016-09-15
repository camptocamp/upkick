package config

import (
	"bytes"
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
)

// Config stores the handler's configuration and UI interface parameters
type Config struct {
	Version             bool   `short:"V" long:"version" description:"Display version."`
	Loglevel            string `short:"l" long:"loglevel" description:"Set loglevel ('debug', 'info', 'warn', 'error', 'fatal', 'panic')." env:"UPKICK_LOG_LEVEL" default:"info"`
	Manpage             bool   `short:"m" long:"manpage" description:"Output manpage."`
	JSON                bool   `short:"j" long:"json" description:"Log as JSON (to stderr)." env:"UPKICK_JSON_OUTPUT"`
	Warn                bool   `short:"w" long:"warn-only" description:"Only warn, do not kick out-of-date containers." env:"UPKICK_WARN_ONLY"`
	HostnameFromRancher bool   `short:"H" long:"hostname-from-rancher" description:"Retrieve hostname from Rancher metadata." env:"CONPLICITY_HOSTNAME_FROM_RANCHER"`

	Docker struct {
		Endpoint string `short:"e" long:"docker-endpoint" description:"The Docker endpoint." env:"DOCKER_ENDPOINT" default:"unix:///var/run/docker.sock"`
	} `group:"Docker Options"`

	Metrics struct {
		PushgatewayURL string `short:"g" long:"gateway-url" description:"The prometheus push gateway URL to use." env:"PUSHGATEWAY_URL"`
	} `group:"Metrics Options"`
}

// LoadConfig loads the config from flags & environment
func LoadConfig(version string) *Config {
	var c Config
	parser := flags.NewParser(&c, flags.Default)
	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}

	if c.Version {
		fmt.Printf("Upkick v%v\n", version)
		os.Exit(0)
	}

	if c.Manpage {
		var buf bytes.Buffer
		parser.ShortDescription = "Unattended upgrades for Docker containers"
		parser.LongDescription = `Upkick pulls Docker images and removes containers using obsolete images.

Make sure your Docker orchestrator is set to recreate the containers.
`
		parser.WriteManPage(&buf)
		fmt.Printf(buf.String())
		os.Exit(0)
	}
	return &c
}
