package main

import (
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
	IgnoredWorkspaces []string `hcl:"ignored_workspaces"`
	IgnoredProjects   []string `hcl:"ignored_projects"`

	DefaultTTL            string `hcl:"default_ttl"`
	defaultTTLDuration    time.Duration
	defaultTTLHoursOrDays string
	MaxTTL                string `hcl:"max_ttl"`
	maxTTLDuration        time.Duration
}

func mustGetEnvVar(varName string) string {
	value := os.Getenv(varName)
	if value == "" {
		log.Fatalf("%s not set. Set it, or run:\n\tvlt run -c \"go run .\"", varName)
	}
	return value
}

func init() {
	// TODO: flag to specify config file

	// TODO: flag to specify debug level
	// log.SetLevel(log.DebugLevel)

	// TODO: optional config for this
	// log.SetFormatter(&log.JSONFormatter{})
	// log.SetLevel(log.DebugLevel)

	config.TFEToken = mustGetEnvVar("TFE_TOKEN")
	err := hclsimple.DecodeFile("config.hcl", nil, &config)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

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

	// TOOD: check defaultTTL <= maxTTL

	config.defaultTTLHoursOrDays = roundDurationToHoursOrDays(config.defaultTTLDuration)
	log.Debugf("Using default TTL: %s, or %s for the TFC API", config.defaultTTLDuration, config.defaultTTLHoursOrDays)

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
