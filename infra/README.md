# Infrastructure Deployment

This directory contains Terraform configuration to deploy Ling App to AWS.

## Prerequisites

1. **AWS CLI** installed and configured
2. **Terraform** >= 1.0 installed
3. **AWS Account** with admin access

## Quick Start

### 1. Configure Variables

```bash
cp terraform.tfvars.example terraform.tfvars
```

Edit `terraform.tfvars` with your secrets:
- `session_secret`: Generate with `openssl rand -hex 32`
- `openai_api_key`: From OpenAI dashboard
- OAuth and Stripe credentials (optional)

### 2. Initialize Terraform

```bash
cd infra
terraform init
```

### 3. Review the Plan

```bash
terraform plan
```

### 4. Deploy

```bash
terraform apply
```

This creates:
- VPC with public/private subnets
- RDS PostgreSQL database
- S3 bucket for audio storage
- ECR repositories for container images
- ECS Fargate cluster with 3 services (API, Web, ML)
- Application Load Balancer

### 5. Get Outputs

```bash
terraform output
```

This shows the ALB URL and GitHub secrets you need to configure.

### 6. Configure GitHub Secrets

Go to your GitHub repository → Settings → Secrets and variables → Actions

Add these secrets (values from `terraform output`):
- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`
- `AWS_REGION` (us-east-1)

### 7. Push Initial Images

The first deploy requires images in ECR. Either:
- Push to `main` branch to trigger CI/CD, OR
- Manually build and push:

```bash
# Login to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin YOUR_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com

# Build and push API
cd api
docker build -t YOUR_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/ling-prod-api:latest .
docker push YOUR_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/ling-prod-api:latest

# Repeat for web and ml
```

## Architecture

```
Internet → ALB → Web (nginx, proxies /api/* to API)
                  ↓
                 API (Go) → ML (Python)
                  ↓
                 RDS + S3
```

**Security:**
- Only Web is publicly accessible
- API only accepts traffic from Web (security groups)
- ML only accepts traffic from API
- RDS only accepts traffic from API

## Costs

Estimated monthly cost: **$80-150**
- RDS db.t3.micro: ~$15
- ECS Fargate (3 services): ~$40-80
- NAT Gateway: ~$30
- ALB: ~$20
- S3/ECR: ~$5

## Adding a Domain

1. Get a domain and create a hosted zone in Route 53
2. Request an ACM certificate for your domain
3. Uncomment the HTTPS listener in `alb.tf`
4. Add a Route 53 record pointing to the ALB

## Destroying

```bash
terraform destroy
```

**Warning:** This deletes all data including the database!
