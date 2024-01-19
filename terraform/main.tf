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

resource "tfe_team_token" "team" {
  team_id = tfe_team.team.id
}


#
# HVS App + Secret
#

provider "hcp" {
  # hashi_strawb_testing Project
  project_id = "fbb3e676-6ac9-46ab-9c09-a7864d33c83a"
}

resource "hcp_vault_secrets_app" "app" {
  app_name    = "ephemeral-workspace-checker"
  description = "TFC API token for ${tfe_team.team.name} in hashi_strawb_testing"
}

resource "hcp_vault_secrets_secret" "secret" {
  app_name     = hcp_vault_secrets_app.app.app_name
  secret_name  = "TFE_TOKEN"
  secret_value = tfe_team_token.team.token
}



# TODO: HCP Service Principal, IAM Policy & Binding
# TODO: HCP Workload Identity Config
