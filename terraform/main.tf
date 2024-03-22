terraform {
  cloud {
    organization = "hashi_strawb_testing"

    workspaces {
      name = "ephemeral_workspace_check-HCP"
    }
  }

  required_providers {
    tfe = {
      source  = "hashicorp/tfe"
      version = "~> 0.51"
    }

    hcp = {
      source  = "hashicorp/hcp"
      version = "~> 0.80"
    }
  }
}



#
# TFC Team + API Token
#

provider "tfe" {
  organization = "hashi_strawb_testing"
}

resource "tfe_team" "team" {
  name = "ephemeral-workspace-checker"
  organization_access {
    manage_workspaces = true
    manage_projects   = true
  }
}

resource "time_rotating" "rotate" {
  rotation_days = 7
}
# workaround from https://github.com/hashicorp/terraform-provider-time/issues/118#issuecomment-1316056478
resource "time_static" "rotate" {
  rfc3339 = time_rotating.rotate.rfc3339
}

resource "tfe_team_token" "team" {
  team_id = tfe_team.team.id

  lifecycle {
    replace_triggered_by = [
      time_static.rotate
    ]
  }
}


#
# HVS App + Secret
#

provider "hcp" {
  # hashi_strawb_testing Project
  project_id = "fbb3e676-6ac9-46ab-9c09-a7864d33c83a"
}

resource "hcp_vault_secrets_app" "app" {
  app_name    = "tfc-ephemeral-workspace-checker"
  description = "TFC API token for ${tfe_team.team.name} in hashi_strawb_testing"
}

resource "hcp_vault_secrets_secret" "secret" {
  app_name     = hcp_vault_secrets_app.app.app_name
  secret_name  = "TFE_TOKEN"
  secret_value = tfe_team_token.team.token
}


#
# HCP Service Principal, IAM Policy & Binding
#

/*
resource "hcp_service_principal" "sp" {
  name = "tfc-ephemeral-workspace-checker"
}

resource "hcp_project_iam_binding" "example" {
  principal_id = hcp_service_principal.sp.resource_id
  role         = "roles/viewer"
}
*/


#
# HCP Workload Identity Config
#

# https://registry.terraform.io/providers/hashicorp/hcp/latest/docs/resources/iam_workload_identity_provider
# TODO: Theoretically... we can use GitHub Actions workload identity to auth with HCP...
# but I can't find any documentation on how to do this.
# So for now, we use Secrets Sync to push the TFE_TOKEN to GitHub Actions
#
# Which... has to be done manually for now.
# https://developer.hashicorp.com/hcp/docs/vault-secrets/integrations/github-actions
