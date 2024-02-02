tfe_org = "hashi_strawb_testing"

ignored_workspaces = [
  "ws-g8SALRtooyGs6JKH", # My bootstrap workspace for this script
]

ignored_projects = [
  "prj-5ZkQWUDRZFAUzn9Q", # Terraform Cloud Demo
  "prj-m31hcrTHKS1uiPAv", # Stacks pre-preview
  "prj-GiZbB7qAkHnMUkdf", # Org Management
  "prj-pcTZQ1XybSPjBScu", # Default Project
  "prj-gpiFDBZUvLnYaqJX", # Azure Dynamic Creds
]

default_ttl = "72h"  # 3 days
max_ttl     = "168h" # 7 days

log_level = "info"
dry_run   = false