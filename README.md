# TFC Workspace Checker

Checks workspaces in a TFC org to validate that they have Auto Destroy configured

## Script

The `script` is written in Go, and requires a config file similar to the one in
`config.hcl`. All fields in that example are mandatory, but the `ignored_*` lists can be empty (`[]`).

You also need a TFE_TOKEN environment variable set.

The easiest way to run this is with HCP Vault Secrets:

```
vlt run -c "go run main.go"
```
(you first need to configure)

### TODO

* [ ] Set a compliance TTL (workspace must auto-destroy after X duration)
* [ ] Automatically set a default TTL on non-compliant workspaces
    * [ ] Optionally skip this with a `--dry-run` argument
* [ ] A whole bunch of refactoring would be nice


Stretch Goal:
* [ ] Notify Slack when script finds non-compliant workspaces

## HCP Vault Secrets Setup

In `terraform`, there is example config to create a TFC team with sufficient
permission to access workspaces.

TF will also create an API token for this team, and store it in a new HCP Vault
Secrets application.

Once this is set up, you can initialise HCP Vault Secrets in the `script` dir with:

```
vlt config init --local
```