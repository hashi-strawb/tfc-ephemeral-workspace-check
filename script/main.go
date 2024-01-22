package main

import (
	"context"

	log "github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()

	workspaces, err := client.Workspaces.List(ctx, config.TFEOrg, nil)
	if err != nil {
		log.Fatal(err)
	}

	for _, workspace := range workspaces.Items {
		_, _ = config.CheckWorkspace(workspace)
	}

	// TODO: if there were any errors, end with a non-zero exit code
	// TODO: if there were any non-compliant workspaces, end with a non-zero exit code
	// e.g. 1 = errors, 2 = non-compliant workspaces, 3 = both
}
