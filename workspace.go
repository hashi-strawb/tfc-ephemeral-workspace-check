package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/go-tfe"
	log "github.com/sirupsen/logrus"
)

// CheckWorkspace checks a single workspace for compliance
// and returns true if compliant, false if not, and an error if
// something goes wrong checking
func (c *Config) CheckWorkspace(workspace *tfe.Workspace) (bool, error) {
	if contains(config.IgnoredProjects, workspace.Project.ID) {
		log.WithFields(log.Fields{
			"project-id":     workspace.Project.ID,
			"workspace-id":   workspace.ID,
			"workspace-name": workspace.Name,
		}).Debugf("Skipped (project is ignored)")

		return true, nil
	}

	if contains(config.IgnoredWorkspaces, workspace.ID) {
		log.WithFields(log.Fields{
			"project-id":     workspace.Project.ID,
			"workspace-id":   workspace.ID,
			"workspace-name": workspace.Name,
		}).Debugf("Skipped (workspace is ignored)")

		return true, nil
	}

	wsDestroy, err := config.GetWorkspaceAutoDestroyDetails(workspace.ID)
	if err != nil {
		log.WithFields(log.Fields{
			"project-id":     workspace.Project.ID,
			"workspace-id":   workspace.ID,
			"workspace-name": workspace.Name,
		}).Errorf("Failed to parse auto-destroy time: %v", err)

		return false, fmt.Errorf("failed get workspace auto-destroy details: %v", err)
	}

	// If we have a duration set, we're compliant.
	if wsDestroy.Data.Attributes.AutoDestroyActivityDuration != "" {

		// TODO: Check that the duration is sufficiently short
		//
		// This would work...
		// duration, err := time.ParseDuration(wsDestroy.Data.Attributes.AutoDestroyActivityDuration)
		//
		// if not for the fact that Go's time package deliberately does not
		// understand durations like "1d" (due to edge-cases in how long a
		// day could be)
		//
		// There are community packages that do understand this, but I'm
		// not sure I want to add a dependency on one of them yet.
		//
		// The mere existence of an activity duration is enough to be
		// compliant for now.
		// Once this is included in the go-tfe library, I would expect that
		// parsing to be done at that point

		log.WithFields(log.Fields{
			"project-id":     workspace.Project.ID,
			"workspace-id":   workspace.ID,
			"workspace-name": workspace.Name,

			"auto-destroy-activity-duration": wsDestroy.Data.Attributes.AutoDestroyActivityDuration,
			"auto-destroy-at":                wsDestroy.Data.Attributes.AutoDestroyAt,
			"auto-destroy-status":            wsDestroy.Data.Attributes.AutoDestroyStatus,
		}).Infof("Compliant (activity duration set)")

		return true, nil
	}

	// if not... if we have a fixed time
	if wsDestroy.Data.Attributes.AutoDestroyAt != "" {
		// Parse the fixed time
		autoDestroyTime, err := time.Parse(time.RFC3339, wsDestroy.Data.Attributes.AutoDestroyAt)
		if err != nil {
			log.WithFields(log.Fields{
				"project-id":     workspace.Project.ID,
				"workspace-id":   workspace.ID,
				"workspace-name": workspace.Name,
			}).Errorf("Failed to parse auto-destroy time: %v", err)

			return false, fmt.Errorf("failed to parse auto-destroy time: %v", err)
		}

		// Now check that the fixed time is <= the max TTL
		if autoDestroyTime.After(time.Now().Add(config.maxTTLDuration)) {
			log.WithFields(log.Fields{
				"project-id":     workspace.Project.ID,
				"workspace-id":   workspace.ID,
				"workspace-name": workspace.Name,

				"auto-destroy-activity-duration": wsDestroy.Data.Attributes.AutoDestroyActivityDuration,
				"auto-destroy-at":                wsDestroy.Data.Attributes.AutoDestroyAt,
				"auto-destroy-status":            wsDestroy.Data.Attributes.AutoDestroyStatus,
			}).Warnf("Non-compliant (auto-destroy time > max TTL)")

			return false, nil
		}

		log.WithFields(log.Fields{
			"project-id":     workspace.Project.ID,
			"workspace-id":   workspace.ID,
			"workspace-name": workspace.Name,

			"auto-destroy-activity-duration": wsDestroy.Data.Attributes.AutoDestroyActivityDuration,
			"auto-destroy-at":                wsDestroy.Data.Attributes.AutoDestroyAt,
			"auto-destroy-status":            wsDestroy.Data.Attributes.AutoDestroyStatus,
		}).Infof("Compliant (auto-destroy time <= max TTL)")

		return true, nil
	}

	log.WithFields(log.Fields{
		"project-id":     workspace.Project.ID,
		"workspace-id":   workspace.ID,
		"workspace-name": workspace.Name,

		"auto-destroy-activity-duration": wsDestroy.Data.Attributes.AutoDestroyActivityDuration,
		"auto-destroy-at":                wsDestroy.Data.Attributes.AutoDestroyAt,
		"auto-destroy-status":            wsDestroy.Data.Attributes.AutoDestroyStatus,
	}).Warnf("Non-compliant (auto-destroy not set)")

	return false, nil
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
	Data WorkspaceAutoDestroyData `json:"data"`
}
type WorkspaceAutoDestroyData struct {
	ID         string                             `json:"id,omitempty"`
	Attributes WorkspaceAutoDestroyDataAttributes `json:"attributes"`
}
type WorkspaceAutoDestroyDataAttributes struct {
	AutoDestroyActivityDuration string `json:"auto-destroy-activity-duration,omitempty"`
	AutoDestroyAt               string `json:"auto-destroy-at,omitempty"`
	AutoDestroyStatus           string `json:"auto-destroy-status,omitempty"`
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

	body, err := io.ReadAll(io.Reader(resp.Body))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Errorf("Error from TFC: %s", body)
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var payload WorkspaceAutoDestroy
	err = json.Unmarshal(body, &payload)
	if err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *Config) UpdateWorkspaceTTL(workspace *tfe.Workspace) error {
	err := c.SetWorkspaceAutoDestroyDetails(workspace.ID)
	if err != nil {
		log.WithFields(log.Fields{
			"project-id":     workspace.Project.ID,
			"workspace-id":   workspace.ID,
			"workspace-name": workspace.Name,

			"auto-destroy-activity-duration": config.defaultTTLHoursOrDays,
		}).Errorf("Failed to update workspace TTL: %v", err)
		return err
	}

	log.WithFields(log.Fields{
		"project-id":     workspace.Project.ID,
		"workspace-id":   workspace.ID,
		"workspace-name": workspace.Name,

		"auto-destroy-activity-duration": config.defaultTTLHoursOrDays,
	}).Infof("Updated workspace TTL")

	return nil
}

func (c *Config) SetWorkspaceAutoDestroyDetails(workspaceID string) error {
	autoDestroyConfig := WorkspaceAutoDestroy{
		Data: WorkspaceAutoDestroyData{
			Attributes: WorkspaceAutoDestroyDataAttributes{
				AutoDestroyActivityDuration: config.defaultTTLHoursOrDays,
			},
		},
	}
	jsonData, _ := json.Marshal(autoDestroyConfig)

	req, err := http.NewRequest("PATCH", "https://app.terraform.io/api/v2/workspaces/"+workspaceID, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.TFEToken)
	req.Header.Set("Content-Type", "application/vnd.api+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.Reader(resp.Body))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		log.Errorf("Error from TFC: %s", body)

		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// don't bother parsing the body; assume it's fine

	return nil
}
