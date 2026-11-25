# IVS Low-Latency channel + auto-record to S3 (stack/02-live-video, ADR-0002, ADR-0003).
# Instantiate once per region: eu-central-1 (primary) and eu-west-1 (standby). Milan has no IVS.

terraform {
  required_version = "~> 1.17"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.74"
    }
  }
}

variable "name" {
  description = "Channel name, e.g. tirivia-prod-main."
  type        = string
}

variable "recording_bucket_name" {
  description = "S3 bucket for recordings. Must be in the same region as the channel."
  type        = string
}

variable "recording_reconnect_window_seconds" {
  description = "Merge reconnects within this window into one recording (0-300). DEFAULT 90 (SPEC §6.9)."
  type        = number
  default     = 90
  validation {
    condition     = var.recording_reconnect_window_seconds >= 0 && var.recording_reconnect_window_seconds <= 300
    error_message = "IVS allows 0-300 seconds."
  }
}

variable "tags" {
  type    = map(string)
  default = {}
}

resource "aws_ivs_recording_configuration" "this" {
  name                               = "${var.name}-rec"
  recording_reconnect_window_seconds = var.recording_reconnect_window_seconds

  destination_configuration {
    s3 {
      bucket_name = var.recording_bucket_name
    }
  }

  thumbnail_configuration {
    recording_mode          = "INTERVAL"
    target_interval_seconds = 10
  }

  tags = var.tags
}

resource "aws_ivs_channel" "this" {
  name         = var.name
  type         = "ADVANCED_HD" # 720p top rendition, server-side transcoded ladder (research 01)
  latency_mode = "LOW"

  recording_configuration_arn = aws_ivs_recording_configuration.this.arn

  tags = var.tags
}

output "channel_arn" {
  value = aws_ivs_channel.this.arn
}

output "playback_url" {
  description = "Returned to apps via JoinShow; never hard-coded in apps."
  value       = aws_ivs_channel.this.playback_url
}

output "ingest_endpoint" {
  value = aws_ivs_channel.this.ingest_endpoint
}
