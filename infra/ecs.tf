# ECS Cluster, Services, and Task Definitions

# ECS Cluster
resource "aws_ecs_cluster" "main" {
  name = "${local.name_prefix}-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }

  tags = {
    Name = "${local.name_prefix}-cluster"
  }
}

# ECS Cluster Capacity Providers (Fargate)
resource "aws_ecs_cluster_capacity_providers" "main" {
  cluster_name = aws_ecs_cluster.main.name

  capacity_providers = ["FARGATE", "FARGATE_SPOT"]

  default_capacity_provider_strategy {
    base              = 1
    weight            = 100
    capacity_provider = "FARGATE"
  }
}

# CloudWatch Log Groups
resource "aws_cloudwatch_log_group" "api" {
  name              = "/ecs/${local.name_prefix}-api"
  retention_in_days = 30

  tags = {
    Name = "${local.name_prefix}-api-logs"
  }
}

resource "aws_cloudwatch_log_group" "web" {
  name              = "/ecs/${local.name_prefix}-web"
  retention_in_days = 30

  tags = {
    Name = "${local.name_prefix}-web-logs"
  }
}

resource "aws_cloudwatch_log_group" "ml" {
  name              = "/ecs/${local.name_prefix}-ml"
  retention_in_days = 30

  tags = {
    Name = "${local.name_prefix}-ml-logs"
  }
}

# IAM Role for ECS Task Execution
resource "aws_iam_role" "ecs_execution" {
  name = "${local.name_prefix}-ecs-execution"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "ecs_execution" {
  role       = aws_iam_role.ecs_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# IAM Role for ECS Tasks (application permissions)
resource "aws_iam_role" "ecs_task" {
  name = "${local.name_prefix}-ecs-task"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })
}

# Attach S3 access policy to task role
resource "aws_iam_role_policy_attachment" "ecs_task_s3" {
  role       = aws_iam_role.ecs_task.name
  policy_arn = aws_iam_policy.s3_access.arn
}

# Service Discovery Namespace (for internal DNS)
resource "aws_service_discovery_private_dns_namespace" "main" {
  name        = "ling.local"
  description = "Private DNS namespace for Ling App services"
  vpc         = aws_vpc.main.id
}

# Service Discovery for API
resource "aws_service_discovery_service" "api" {
  name = "api"

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.main.id

    dns_records {
      ttl  = 10
      type = "A"
    }

    routing_policy = "MULTIVALUE"
  }

  health_check_custom_config {
    failure_threshold = 1
  }
}

# Service Discovery for ML
resource "aws_service_discovery_service" "ml" {
  name = "ml"

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.main.id

    dns_records {
      ttl  = 10
      type = "A"
    }

    routing_policy = "MULTIVALUE"
  }

  health_check_custom_config {
    failure_threshold = 1
  }
}

# ============================================
# API Service
# ============================================

resource "aws_ecs_task_definition" "api" {
  family                   = "${local.name_prefix}-api"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.api_cpu
  memory                   = var.api_memory
  execution_role_arn       = aws_iam_role.ecs_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([
    {
      name  = "api"
      image = "${aws_ecr_repository.api.repository_url}:latest"

      portMappings = [
        {
          containerPort = 8080
          hostPort      = 8080
          protocol      = "tcp"
        }
      ]

      environment = [
        { name = "PORT", value = "8080" },
        { name = "HOST", value = "0.0.0.0" },
        { name = "GIN_MODE", value = "release" },
        { name = "ENVIRONMENT", value = var.environment },
        { name = "ML_SERVICE_URL", value = "http://ml.ling.local:8001" },
        { name = "S3_BUCKET", value = aws_s3_bucket.audio.id },
        { name = "S3_REGION", value = var.aws_region },
        { name = "CORS_ALLOWED_ORIGINS", value = "https://${aws_lb.main.dns_name}" },
      ]

      secrets = [
        { name = "DATABASE_URL", valueFrom = aws_ssm_parameter.database_url.arn },
        { name = "SESSION_SECRET", valueFrom = aws_ssm_parameter.session_secret.arn },
        { name = "OPENAI_API_KEY", valueFrom = aws_ssm_parameter.openai_api_key.arn },
        { name = "GOOGLE_CLIENT_ID", valueFrom = aws_ssm_parameter.google_client_id.arn },
        { name = "GOOGLE_CLIENT_SECRET", valueFrom = aws_ssm_parameter.google_client_secret.arn },
        { name = "GITHUB_CLIENT_ID", valueFrom = aws_ssm_parameter.github_client_id.arn },
        { name = "GITHUB_CLIENT_SECRET", valueFrom = aws_ssm_parameter.github_client_secret.arn },
        { name = "STRIPE_SECRET_KEY", valueFrom = aws_ssm_parameter.stripe_secret_key.arn },
        { name = "STRIPE_WEBHOOK_SECRET", valueFrom = aws_ssm_parameter.stripe_webhook_secret.arn },
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = aws_cloudwatch_log_group.api.name
          awslogs-region        = var.aws_region
          awslogs-stream-prefix = "ecs"
        }
      }

      healthCheck = {
        command     = ["CMD-SHELL", "wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1"]
        interval    = 30
        timeout     = 5
        retries     = 3
        startPeriod = 60
      }

      essential = true
    }
  ])

  tags = {
    Name = "${local.name_prefix}-api"
  }
}

resource "aws_ecs_service" "api" {
  name            = "${local.name_prefix}-api"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.api.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = aws_subnet.private[*].id
    security_groups  = [aws_security_group.api.id]
    assign_public_ip = false
  }

  service_registries {
    registry_arn = aws_service_discovery_service.api.arn
  }

  depends_on = [aws_iam_role_policy_attachment.ecs_execution]

  tags = {
    Name = "${local.name_prefix}-api"
  }

  lifecycle {
    ignore_changes = [task_definition] # Allow CI/CD to update
  }
}

# ============================================
# Web Service
# ============================================

resource "aws_ecs_task_definition" "web" {
  family                   = "${local.name_prefix}-web"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.web_cpu
  memory                   = var.web_memory
  execution_role_arn       = aws_iam_role.ecs_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([
    {
      name  = "web"
      image = "${aws_ecr_repository.web.repository_url}:latest"

      portMappings = [
        {
          containerPort = 80
          hostPort      = 80
          protocol      = "tcp"
        }
      ]

      environment = [
        # nginx uses envsubst to inject this into the config for proxying /api/* to API service
        { name = "API_URL", value = "http://api.ling.local:8080" },
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = aws_cloudwatch_log_group.web.name
          awslogs-region        = var.aws_region
          awslogs-stream-prefix = "ecs"
        }
      }

      healthCheck = {
        command     = ["CMD-SHELL", "wget --no-verbose --tries=1 --spider http://localhost:80/ || exit 1"]
        interval    = 30
        timeout     = 5
        retries     = 3
        startPeriod = 30
      }

      essential = true
    }
  ])

  tags = {
    Name = "${local.name_prefix}-web"
  }
}

resource "aws_ecs_service" "web" {
  name            = "${local.name_prefix}-web"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.web.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = aws_subnet.public[*].id
    security_groups  = [aws_security_group.web.id]
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.web.arn
    container_name   = "web"
    container_port   = 80
  }

  depends_on = [
    aws_lb_listener.http,
    aws_iam_role_policy_attachment.ecs_execution
  ]

  tags = {
    Name = "${local.name_prefix}-web"
  }

  lifecycle {
    ignore_changes = [task_definition] # Allow CI/CD to update
  }
}

# ============================================
# ML Service
# ============================================

resource "aws_ecs_task_definition" "ml" {
  family                   = "${local.name_prefix}-ml"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.ml_cpu
  memory                   = var.ml_memory
  execution_role_arn       = aws_iam_role.ecs_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([
    {
      name  = "ml"
      image = "${aws_ecr_repository.ml.repository_url}:latest"

      portMappings = [
        {
          containerPort = 8001
          hostPort      = 8001
          protocol      = "tcp"
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = aws_cloudwatch_log_group.ml.name
          awslogs-region        = var.aws_region
          awslogs-stream-prefix = "ecs"
        }
      }

      healthCheck = {
        command     = ["CMD-SHELL", "curl -f http://localhost:8001/health || exit 1"]
        interval    = 30
        timeout     = 10
        retries     = 3
        startPeriod = 120 # ML service needs longer to start (downloads models)
      }

      essential = true
    }
  ])

  tags = {
    Name = "${local.name_prefix}-ml"
  }
}

resource "aws_ecs_service" "ml" {
  name            = "${local.name_prefix}-ml"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.ml.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = aws_subnet.private[*].id
    security_groups  = [aws_security_group.ml.id]
    assign_public_ip = false
  }

  service_registries {
    registry_arn = aws_service_discovery_service.ml.arn
  }

  depends_on = [aws_iam_role_policy_attachment.ecs_execution]

  tags = {
    Name = "${local.name_prefix}-ml"
  }

  lifecycle {
    ignore_changes = [task_definition] # Allow CI/CD to update
  }
}

# ============================================
# SSM Parameters for Secrets
# ============================================

resource "aws_ssm_parameter" "database_url" {
  name  = "/${local.name_prefix}/DATABASE_URL"
  type  = "SecureString"
  value = "postgresql://${var.db_username}:${random_password.db_password.result}@${aws_db_instance.main.endpoint}/${var.db_name}"

  tags = {
    Name = "${local.name_prefix}-database-url"
  }
}

resource "aws_ssm_parameter" "session_secret" {
  name  = "/${local.name_prefix}/SESSION_SECRET"
  type  = "SecureString"
  value = var.session_secret

  tags = {
    Name = "${local.name_prefix}-session-secret"
  }
}

resource "aws_ssm_parameter" "openai_api_key" {
  name  = "/${local.name_prefix}/OPENAI_API_KEY"
  type  = "SecureString"
  value = var.openai_api_key

  tags = {
    Name = "${local.name_prefix}-openai-api-key"
  }
}

resource "aws_ssm_parameter" "google_client_id" {
  name  = "/${local.name_prefix}/GOOGLE_CLIENT_ID"
  type  = "SecureString"
  value = var.google_client_id != "" ? var.google_client_id : "placeholder"

  tags = {
    Name = "${local.name_prefix}-google-client-id"
  }
}

resource "aws_ssm_parameter" "google_client_secret" {
  name  = "/${local.name_prefix}/GOOGLE_CLIENT_SECRET"
  type  = "SecureString"
  value = var.google_client_secret != "" ? var.google_client_secret : "placeholder"

  tags = {
    Name = "${local.name_prefix}-google-client-secret"
  }
}

resource "aws_ssm_parameter" "github_client_id" {
  name  = "/${local.name_prefix}/GITHUB_CLIENT_ID"
  type  = "SecureString"
  value = var.github_oauth_client_id != "" ? var.github_oauth_client_id : "placeholder"

  tags = {
    Name = "${local.name_prefix}-github-client-id"
  }
}

resource "aws_ssm_parameter" "github_client_secret" {
  name  = "/${local.name_prefix}/GITHUB_CLIENT_SECRET"
  type  = "SecureString"
  value = var.github_oauth_client_secret != "" ? var.github_oauth_client_secret : "placeholder"

  tags = {
    Name = "${local.name_prefix}-github-client-secret"
  }
}

resource "aws_ssm_parameter" "stripe_secret_key" {
  name  = "/${local.name_prefix}/STRIPE_SECRET_KEY"
  type  = "SecureString"
  value = var.stripe_secret_key != "" ? var.stripe_secret_key : "placeholder"

  tags = {
    Name = "${local.name_prefix}-stripe-secret-key"
  }
}

resource "aws_ssm_parameter" "stripe_webhook_secret" {
  name  = "/${local.name_prefix}/STRIPE_WEBHOOK_SECRET"
  type  = "SecureString"
  value = var.stripe_webhook_secret != "" ? var.stripe_webhook_secret : "placeholder"

  tags = {
    Name = "${local.name_prefix}-stripe-webhook-secret"
  }
}

# IAM policy for ECS to read SSM parameters
resource "aws_iam_policy" "ssm_read" {
  name        = "${local.name_prefix}-ssm-read"
  description = "Allow ECS tasks to read SSM parameters"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ssm:GetParameters",
          "ssm:GetParameter"
        ]
        Resource = "arn:aws:ssm:${var.aws_region}:${data.aws_caller_identity.current.account_id}:parameter/${local.name_prefix}/*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "ecs_execution_ssm" {
  role       = aws_iam_role.ecs_execution.name
  policy_arn = aws_iam_policy.ssm_read.arn
}
