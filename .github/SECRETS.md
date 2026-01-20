# GitHub Secrets Configuration

This document lists all the secrets required for CI/CD pipelines and deployment.

## Required Secrets

### AWS Deployment (Required for CI/CD)

These are needed to deploy to AWS ECS:

| Secret | Description | How to Get |
|--------|-------------|------------|
| `AWS_ACCESS_KEY_ID` | AWS access key for ECR/ECS deployment | Create IAM user with ECR/ECS permissions |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key | From IAM user creation |
| `AWS_REGION` | AWS region | `us-east-1` (or your chosen region) |

### Coverage Reporting (Optional)

| Secret | Description | How to Get |
|--------|-------------|------------|
| `CODECOV_TOKEN` | Token for uploading coverage reports | Get from [codecov.io](https://codecov.io) after linking your repo |

## Secrets Managed by Terraform

These secrets are stored in AWS SSM Parameter Store by Terraform. You provide them in `infra/terraform.tfvars`:

| Terraform Variable | SSM Parameter | Description |
|-------------------|---------------|-------------|
| `session_secret` | `/ling-prod/SESSION_SECRET` | Session encryption key |
| `openai_api_key` | `/ling-prod/OPENAI_API_KEY` | OpenAI API key |
| `google_client_id` | `/ling-prod/GOOGLE_CLIENT_ID` | Google OAuth client ID |
| `google_client_secret` | `/ling-prod/GOOGLE_CLIENT_SECRET` | Google OAuth secret |
| `github_oauth_client_id` | `/ling-prod/GITHUB_CLIENT_ID` | GitHub OAuth client ID |
| `github_oauth_client_secret` | `/ling-prod/GITHUB_CLIENT_SECRET` | GitHub OAuth secret |
| `stripe_secret_key` | `/ling-prod/STRIPE_SECRET_KEY` | Stripe API secret |
| `stripe_webhook_secret` | `/ling-prod/STRIPE_WEBHOOK_SECRET` | Stripe webhook secret |
| `stripe_price_basic` | `/ling-prod/STRIPE_PRICE_BASIC` | Stripe price ID for basic tier |
| `stripe_price_pro` | `/ling-prod/STRIPE_PRICE_PRO` | Stripe price ID for pro tier |
| `domain_name` | N/A | Custom domain name (e.g., example.com) |

The `DATABASE_URL` and S3-related configurations are automatically generated and stored in SSM by Terraform.

## Setting Up for Deployment

### Step 1: Create IAM User for GitHub Actions

1. Go to AWS Console → IAM → Users → Create user
2. Name: `github-actions-deploy`
3. Attach these policies:
   - `AmazonEC2ContainerRegistryFullAccess`
   - `AmazonECS_FullAccess`
4. Create access key for "Application running outside AWS"
5. Save the Access Key ID and Secret Access Key

### Step 2: Add GitHub Secrets

1. Go to your GitHub repository
2. Navigate to **Settings** → **Secrets and variables** → **Actions**
3. Add these secrets:
   - `AWS_ACCESS_KEY_ID` - from Step 1
   - `AWS_SECRET_ACCESS_KEY` - from Step 1
   - `AWS_REGION` - `us-east-1`

### Step 3: Run Terraform

```bash
cd infra
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your secrets
terraform init
terraform apply
```

### Step 4: Deploy

Push to `main` branch to trigger deployment.

## Local Development

For local development, copy the example files and fill in the values:

```bash
cp api/example.env api/.env
cp web/example.env.local web/.env.local
```

### Required Local Environment Variables

**API (`api/.env`)**:
- `DATABASE_URL` - PostgreSQL connection string (provided by Docker Compose)
- `SESSION_SECRET` - Generate with: `openssl rand -hex 32`
- `OPENAI_API_KEY` - From OpenAI dashboard

**Optional API Variables** (for specific features):
- OAuth: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`
- Stripe: `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET`, `STRIPE_PRICE_BASIC`, `STRIPE_PRICE_PRO`
- S3/MinIO: `S3_ENDPOINT`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_BUCKET`, `S3_REGION` (defaults work with Docker Compose)
- Service URLs: `ML_SERVICE_URL`, `TTS_SERVICE_URL`, `STT_SERVICE_URL`, `FRONTEND_URL`
- OAuth Redirects: `GOOGLE_REDIRECT_URL`, `GITHUB_REDIRECT_URL`

**Web (`web/.env.local`)**:
- `VITE_API_URL` - API endpoint (default: `http://localhost:8080`)

**Never commit `.env` files with real credentials to the repository.**

## Security Notes

- Rotate secrets regularly
- Use different credentials for staging vs production
- All production secrets are stored in AWS SSM Parameter Store (encrypted)
- The `.env` file is gitignored - keep it that way
- Terraform state contains sensitive data - don't commit `*.tfstate` files
