package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/hcl/v2/hclsimple"
	log "github.com/sirupsen/logrus"
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
		log.Fatalf("%s not set. Set it, or run:\n\tvlt run -c \"go run main.go\"", varName)
	}
	return value
}

func main() {
	var config Config
	config.TFEToken = mustGetEnvVar("TFE_TOKEN")
	err := hclsimple.DecodeFile("config.hcl", nil, &config)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	if config.TFEOrg == "" {
		log.Fatalf("tfe_org not set in config.hcl")
	}

	client, err := tfe.NewClient(&tfe.Config{
		Token:             config.TFEToken,
		RetryServerErrors: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	workspaces, err := client.Workspaces.List(ctx, config.TFEOrg, nil)
	if err != nil {
		log.Fatal(err)
	}

	for _, workspace := range workspaces.Items {
		if contains(config.IgnoredWorkspaces, workspace.ID) || contains(config.IgnoredProjects, workspace.Project.ID) {
			continue
		}

		wsDestroy, _ := config.GetWorkspaceAutoDestroyDetails(workspace.ID)

		// If we have a duration set, we're compliant.
		if wsDestroy.Data.Attributes.AutoDestroyActivityDuration != "" {
			log.WithFields(log.Fields{
				"project-id":     workspace.Project.ID,
				"workspace-id":   workspace.ID,
				"workspace-name": workspace.Name,

				"auto-destroy-activity-duration": wsDestroy.Data.Attributes.AutoDestroyActivityDuration,
				"auto-destroy-at":                wsDestroy.Data.Attributes.AutoDestroyAt,
				"auto-destroy-status":            wsDestroy.Data.Attributes.AutoDestroyStatus,
			}).Infof("Compliant! (activity duration set)")

			continue
		}

		// if not... if we have a fixed time
		if wsDestroy.Data.Attributes.AutoDestroyAt != "" {
			// TODO: we should check that the date is sufficiently soon

			log.WithFields(log.Fields{
				"project-id":     workspace.Project.ID,
				"workspace-id":   workspace.ID,
				"workspace-name": workspace.Name,

				"auto-destroy-activity-duration": wsDestroy.Data.Attributes.AutoDestroyActivityDuration,
				"auto-destroy-at":                wsDestroy.Data.Attributes.AutoDestroyAt,
				"auto-destroy-status":            wsDestroy.Data.Attributes.AutoDestroyStatus,
			}).Infof("Compliant! (fixed time set)")

			continue
		}

		log.WithFields(log.Fields{
			"project-id":     workspace.Project.ID,
			"workspace-id":   workspace.ID,
			"workspace-name": workspace.Name,

			"auto-destroy-activity-duration": wsDestroy.Data.Attributes.AutoDestroyActivityDuration,
			"auto-destroy-at":                wsDestroy.Data.Attributes.AutoDestroyAt,
			"auto-destroy-status":            wsDestroy.Data.Attributes.AutoDestroyStatus,
		}).Warnf("Non-compliant (auto-destroy not set)")

	}
}

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

// Specific fields we care about from the API response, which are not present
// yet in go-tfe's Workspace struct.
type WorkspaceAutoDestroy struct {
	Data struct {
		ID         string `json:"id"`
		Attributes struct {
			AutoDestroyActivityDuration string `json:"auto-destroy-activity-duration"`
			AutoDestroyAt               string `json:"auto-destroy-at"`
			AutoDestroyStatus           string `json:"auto-destroy-status"`
		} `json:"attributes"`
	} `json:"data"`
}

func (c *Config) GetWorkspaceAutoDestroyDetails(workspaceID string) (*WorkspaceAutoDestroy, error) {
	req, err := http.NewRequest("GET", "https://app.terraform.io/api/v2/workspaces/"+workspaceID, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.TFEToken)
	req.Header.Set("Content-Type", "application/vnd.api+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.Reader(resp.Body))
	if err != nil {
		return nil, err
	}

	var payload WorkspaceAutoDestroy
	err = json.Unmarshal(body, &payload)
	if err != nil {
		return nil, err
	}
	return &payload, nil
}
