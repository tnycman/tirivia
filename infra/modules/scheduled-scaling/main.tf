# Pre-show scaling for an ECS service (Games Industry Lens GAMEOPS03-BP05 "Pre-scale").
#
# Show times are NOT fixed in infrastructure: producers schedule each show in the studio
# console (decision UF-1). Terraform owns only the scalable target and the always-on target
# tracking. The backend's pre-show orchestrator (studio module) books one-time scheduled
# actions per show via the Application Auto Scaling API, using `at(...)` expressions:
#   T-4 h  ALB LCU reservation (can take hours to fulfil)
#   T-60   MinCapacity -> show_min, KDS re-shard, Valkey scale-out, warm standby in Ireland
#   T+end  back to baseline after the show is SETTLED
# Rescheduling or cancelling a show deletes and re-books those actions.

terraform {
  required_version = "~> 1.17"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.74"
    }
  }
}

variable "cluster_name" { type = string }
variable "service_name" { type = string }

variable "baseline_min" {
  description = "Tasks outside show windows."
  type        = number
  default     = 2
}

variable "show_min" {
  description = "Floor during a show window; size from the 600k load test. Passed to the orchestrator."
  type        = number
}

variable "max" {
  type = number
}

variable "cpu_target" {
  type    = number
  default = 50
}

resource "aws_appautoscaling_target" "this" {
  service_namespace  = "ecs"
  scalable_dimension = "ecs:service:DesiredCount"
  resource_id        = "service/${var.cluster_name}/${var.service_name}"
  min_capacity       = var.baseline_min
  max_capacity       = var.max

  lifecycle {
    # The orchestrator raises/lowers MinCapacity around each show; don't fight it.
    ignore_changes = [min_capacity]
  }
}

resource "aws_appautoscaling_policy" "cpu" {
  name               = "${var.service_name}-cpu"
  policy_type        = "TargetTrackingScaling"
  service_namespace  = aws_appautoscaling_target.this.service_namespace
  resource_id        = aws_appautoscaling_target.this.resource_id
  scalable_dimension = aws_appautoscaling_target.this.scalable_dimension

  target_tracking_scaling_policy_configuration {
    target_value       = var.cpu_target
    scale_in_cooldown  = 300
    scale_out_cooldown = 30
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
  }
}

output "scaling_config" {
  description = "Handed to the backend pre-show orchestrator (via SSM parameter) to book per-show actions."
  value = {
    service_namespace  = aws_appautoscaling_target.this.service_namespace
    resource_id        = aws_appautoscaling_target.this.resource_id
    scalable_dimension = aws_appautoscaling_target.this.scalable_dimension
    baseline_min       = var.baseline_min
    show_min           = var.show_min
    max                = var.max
  }
}
