package main

import (
	"context"
	"os"

	log "github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()

	workspaces, err := client.Workspaces.List(ctx, config.TFEOrg, nil)
	if err != nil {
		log.Fatal(err)
	}

	hadNonCompliant := false
	hadErrors := false

	for _, workspace := range workspaces.Items {
		compliant, err := config.CheckWorkspace(workspace)
		if err != nil {
			hadErrors = true
		}

		if !compliant {

			// TODO: if dryrun, set hadNonCompliant and continue

			// Now fix the non-compliant workspace
			err = config.UpdateWorkspaceTTL(workspace)
			if err != nil {
				hadErrors = true

				// Only set hadNonCompliant if we had workspaces we were unable to update
				hadNonCompliant = true
			}
		}
	}

	exit(hadNonCompliant, hadErrors)
}

// exit is a helper function to exit the program with the correct exit code
//
// if there were any errors, end with a non-zero exit code
// if there were any non-compliant workspaces, end with a non-zero exit code
// e.g. 1 = errors, 2 = non-compliant workspaces, 3 = both
func exit(hadNonCompliant, hadErrors bool) {
	if hadNonCompliant && hadErrors {
		os.Exit(3)
	} else if hadNonCompliant {
		os.Exit(2)
	} else if hadErrors {
		os.Exit(1)
	}
}
