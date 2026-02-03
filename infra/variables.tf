# Input Variables

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name (e.g., production, staging)"
  type        = string
  default     = "production"
}

# Database
variable "db_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.micro" # ~$15/month, upgrade to db.t3.small for more capacity
}

variable "db_allocated_storage" {
  description = "RDS allocated storage in GB"
  type        = number
  default     = 20
}

variable "db_name" {
  description = "Database name"
  type        = string
  default     = "lingapp"
}

variable "db_username" {
  description = "Database master username"
  type        = string
  default     = "lingapp"
}

# ECS
variable "api_cpu" {
  description = "API task CPU units (1024 = 1 vCPU)"
  type        = number
  default     = 256
}

variable "api_memory" {
  description = "API task memory in MB"
  type        = number
  default     = 512
}

variable "web_cpu" {
  description = "Web task CPU units"
  type        = number
  default     = 256
}

variable "web_memory" {
  description = "Web task memory in MB"
  type        = number
  default     = 512
}

variable "ml_cpu" {
  description = "ML task CPU units (higher for inference)"
  type        = number
  default     = 1024 # 1 vCPU for ML inference
}

variable "ml_memory" {
  description = "ML task memory in MB (higher for models)"
  type        = number
  default     = 4096 # 4GB for ML models (Whisper IPA + faster-whisper, TTS via OpenAI)
}

# Application secrets (passed via environment or tfvars)
variable "session_secret" {
  description = "Session encryption secret"
  type        = string
  sensitive   = true
}

variable "openai_api_key" {
  description = "OpenAI API key"
  type        = string
  sensitive   = true
}

variable "google_client_id" {
  description = "Google OAuth client ID"
  type        = string
  default     = ""
}

variable "google_client_secret" {
  description = "Google OAuth client secret"
  type        = string
  sensitive   = true
  default     = ""
}

variable "github_oauth_client_id" {
  description = "GitHub OAuth client ID"
  type        = string
  default     = ""
}

variable "github_oauth_client_secret" {
  description = "GitHub OAuth client secret"
  type        = string
  sensitive   = true
  default     = ""
}

variable "stripe_secret_key" {
  description = "Stripe secret key"
  type        = string
  sensitive   = true
  default     = ""
}

variable "stripe_webhook_secret" {
  description = "Stripe webhook secret"
  type        = string
  sensitive   = true
  default     = ""
}

variable "stripe_price_basic" {
  description = "Stripe price ID for basic tier"
  type        = string
  default     = ""
}

variable "stripe_price_pro" {
  description = "Stripe price ID for pro tier"
  type        = string
  default     = ""
}

# Domain
variable "domain_name" {
  description = "Custom domain name (e.g., example.com)"
  type        = string
}
