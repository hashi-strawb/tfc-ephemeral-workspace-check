package main

import (
	"os"

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

	// TODO: Default TTL
	// TODO: Compliance TTL
}

func mustGetEnvVar(varName string) string {
	value := os.Getenv(varName)
	if value == "" {
		log.Fatalf("%s not set. Set it, or run:\n\tvlt run -c \"go run .\"", varName)
	}
	return value
}

func init() {
	// Log as JSON instead of the default ASCII formatter.
	// log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	// log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	// log.SetLevel(log.WarnLevel)

	config.TFEToken = mustGetEnvVar("TFE_TOKEN")
	err := hclsimple.DecodeFile("config.hcl", nil, &config)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	if config.TFEOrg == "" {
		log.Fatalf("tfe_org not set in config.hcl")
	}

	client, err = tfe.NewClient(&tfe.Config{
		Token:             config.TFEToken,
		RetryServerErrors: true,
	})
	if err != nil {
		log.Fatal(err)
	}
}
