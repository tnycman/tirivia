variable "account_id" {
  description = "Guard rail: the provider refuses to run in any other account."
  type        = string
}

variable "region" {
  type    = string
  default = "eu-central-1"
}

variable "standby_region" {
  description = "Region of the state replica; Ireland is the standby for every plane (ADR-0003)."
  type        = string
  default     = "eu-west-1"
}

variable "env" {
  description = "Account name, e.g. dev, staging, prod."
  type        = string
}

variable "abort_incomplete_upload_days" {
  description = "Days before S3 aborts an incomplete multipart upload. No published value exists, so it is set per account (OP-7)."
  type        = number
}
