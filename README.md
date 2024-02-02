# TFC Workspace Checker

Checks workspaces in a TFC org to validate that they have Auto Destroy configured

## Script

The `script` is written in Go, and requires a config file similar to the one in
`config.hcl`. Only `tfe_org` is mandatory.

You also need a TFE_TOKEN environment variable set.

The easiest way to run this is with HCP Vault Secrets:

```
vlt run -c "go run . --config config.hcl"
```

(you first need to configure HCP Vault secrets, as below)

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

## HCP Vault Secrets Setup

In `terraform`, there is example config to create a TFC team with sufficient
permission to access workspaces.

TF will also create an API token for this team, and store it in a new HCP Vault
Secrets application.

Once this is set up, you can initialise HCP Vault Secrets in the `script` dir with:

```
vlt config init --local
```
