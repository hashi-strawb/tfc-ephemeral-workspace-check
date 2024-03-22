# TFC Workspace Checker

Checks workspaces in a TFC org to validate that they have Auto Destroy configured

## Use as a GitHub Action

For your repo, you only need a config file and a github workflow file.

For the config file, something like this: `dryrun.hcl`:

```
tfe_org = "hashi_strawb_testing"

ignored_workspaces = [
  "ws-g8SALRtooyGs6JKH", # My bootstrap workspace for this script
]

ignored_projects = [
  "prj-5ZkQWUDRZFAUzn9Q", # Terraform Cloud Demo
]

default_ttl = "72h"  # 3 days
max_ttl     = "168h" # 7 days

log_level = "debug"
```

For the GitHub Actions file, something like this: `.github/workflows/`:

```
name: Dry Run

on:
  workflow_dispatch:
  schedule:
    - cron: '0 0/12 * * *'
  push:
    branches: [ "main" ]

jobs:

  run:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - id: dryrun
      uses: hashi-strawb/tfc-ephemeral-workspace-check@v0.1.0
      with:
        tfe-token: ${{ secrets.TFE_TOKEN }}
        config: dryrun.hcl
```


## Script

The script is written in Go, and requires a config file similar to the one mentioned above. Only `tfe_org` is mandatory.

You also need a TFE_TOKEN environment variable set.

The easiest way to run this is with HCP Vault Secrets:

```
vlt run -c "go run . --config config.hcl"
```
You should get an output similar to this:

```
$ vlt run -c "go run ."
INFO[0001] Compliant! (activity duration set)            auto-destroy-activity-duration=1d auto-destroy-at= auto-destroy-status= project-id=prj-UpKoCoERU4EmkoGV workspace-id=ws-1EA9MzkuuUn5b7UJ workspace-name=test-ws-ephemeral-inactive
WARN[0001] Non-compliant (auto-destroy not set)          auto-destroy-activity-duration= auto-destroy-at= auto-destroy-status= project-id=prj-UpKoCoERU4EmkoGV workspace-id=ws-fTiAgLDe2jmNyUPA workspace-name=nocode-test
WARN[0001] Non-compliant (auto-destroy not set)          auto-destroy-activity-duration= auto-destroy-at= auto-destroy-status= project-id=prj-UpKoCoERU4EmkoGV workspace-id=ws-ViGbWb4YTPpaEcrS workspace-name=test-ws-persistant
INFO[0002] Compliant! (fixed time set)                   auto-destroy-activity-duration= auto-destroy-at="2100-01-01T00:00:00.000Z" auto-destroy-status= project-id=prj-UpKoCoERU4EmkoGV workspace-id=ws-Zt6RJ7PxUsPScJhN workspace-name=test-ws-ephemeral-static
```

### TODO

* [X] Set a compliance TTL (workspace must auto-destroy after X duration)
* [X] Automatically set a default TTL on non-compliant workspaces
    * [X] Optionally skip this with a `--dry-run` argument
* [ ] A whole bunch of refactoring would be nice