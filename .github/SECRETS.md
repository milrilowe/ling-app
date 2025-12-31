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

The `DATABASE_URL` is automatically generated and stored in SSM by Terraform.

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

For local development, copy `api/example.env` to `api/.env` and fill in the values:

```bash
cp api/example.env api/.env
```

**Never commit `.env` files with real credentials to the repository.**

## Security Notes

- Rotate secrets regularly
- Use different credentials for staging vs production
- All production secrets are stored in AWS SSM Parameter Store (encrypted)
- The `.env` file is gitignored - keep it that way
- Terraform state contains sensitive data - don't commit `*.tfstate` files
