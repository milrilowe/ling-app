# Ling App Infrastructure
# Terraform configuration for AWS deployment

terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }

  # Uncomment after first apply to enable remote state
  # backend "s3" {
  #   bucket         = "ling-app-terraform-state"
  #   key            = "terraform.tfstate"
  #   region         = "us-east-1"
  #   encrypt        = true
  #   dynamodb_table = "ling-app-terraform-locks"
  # }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "ling-app"
      Environment = var.environment
      ManagedBy   = "terraform"
    }
  }
}

# Random suffix for globally unique names
resource "random_id" "suffix" {
  byte_length = 4
}

locals {
  name_prefix = "ling-${var.environment}"
  common_tags = {
    Project     = "ling-app"
    Environment = var.environment
  }
}
