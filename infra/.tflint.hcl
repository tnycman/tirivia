# TFLint 0.66.0 with the AWS ruleset (stack/09-infrastructure-aws).
plugin "terraform" {
  enabled = true
  preset  = "recommended"
}

plugin "aws" {
  enabled = true
  version = "0.50.0" # latest release per GitHub API (published 2026-10-23)
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}
