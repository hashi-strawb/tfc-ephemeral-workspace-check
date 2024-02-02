package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/hcl/v2/hclsimple"
	log "github.com/sirupsen/logrus"
)

var (
	config Config
	client *tfe.Client
)

type Config struct {
	TFEToken          string
	TFEOrg            string   `hcl:"tfe_org"`
	IgnoredWorkspaces []string `hcl:"ignored_workspaces,optional"`
	IgnoredProjects   []string `hcl:"ignored_projects,optional"`

	DefaultTTL            string `hcl:"default_ttl,optional"`
	defaultTTLDuration    time.Duration
	defaultTTLHoursOrDays string
	MaxTTL                string `hcl:"max_ttl,optional"`
	maxTTLDuration        time.Duration

	DryRun bool `hcl:"dry_run,optional"`
	// logFormat string `hcl:"log_format"`
	LogLevel string `hcl:"log_level,optional"`
}

var defaultConfig = Config{
	DefaultTTL: "24h",
	MaxTTL:     "168h",

	DryRun:   true,
	LogLevel: "info",
}

func mustGetEnvVar(varName string) string {
	value := os.Getenv(varName)
	if value == "" {
		log.Fatalf("%s not set. Set it, or run:\n\tvlt run -c \"go run .\"", varName)
	}
	return value
}

func init() {
	// initialise config with defaults
	config = defaultConfig

	// load config from file... but which one?
	var configFile string
	flag.StringVar(&configFile, "config", "config.hcl", "config file to load")
	flag.Parse()

	config.TFEToken = mustGetEnvVar("TFE_TOKEN")

	err := hclsimple.DecodeFile(configFile, nil, &config)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	level, err := log.ParseLevel(config.LogLevel)
	if err != nil {
		level = log.InfoLevel
	}
	log.SetLevel(level)

	// Validate that TFE Org is set to a non-empty string
	if config.TFEOrg == "" {
		log.Fatalf("tfe_org not set in config.hcl")
	}

	// Validate that TTL strings can be converted to Durations
	config.defaultTTLDuration, err = time.ParseDuration(config.DefaultTTL)
	if err != nil {
		log.Fatalf("Failed to parse default_ttl: %v", err)
	}
	config.maxTTLDuration, err = time.ParseDuration(config.MaxTTL)
	if err != nil {
		log.Fatalf("Failed to parse max_ttl: %v", err)
	}

	// TODO: check defaultTTL <= maxTTL

	config.defaultTTLHoursOrDays = roundDurationToHoursOrDays(config.defaultTTLDuration)

	client, err = tfe.NewClient(&tfe.Config{
		Token:             config.TFEToken,
		RetryServerErrors: true,
	})
	if err != nil {
		log.Fatal(err)
	}
}

// round the duration down to the nearest number of hours, and return suffixed with "h"
// if a multiple of 24, return the number of days instead, suffixed with "d"
func roundDurationToHoursOrDays(d time.Duration) string {
	hours := int(d.Round(time.Hour).Hours())

	if hours%24 == 0 {
		return fmt.Sprintf("%vd", hours/24)
	}

	return fmt.Sprintf("%vh", hours)
}
