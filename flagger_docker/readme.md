# Docker Containerization Guide

## Table of Contents
1. [Introduction to Docker](#introduction-to-docker)
2. [Docker Basics](#docker-basics)
3. [Dockerfile Creation](#dockerfile-creation)
4. [Docker Commands](#docker-commands)
5. [Docker Compose](#docker-compose)
6. [AWS Deployment](#aws-deployment)
7. [Best Practices](#best-practices)

## Introduction to Docker

Docker is a platform for developing, shipping, and running applications in containers. Containers are lightweight, portable, and self-contained units that can run anywhere Docker is installed.

### Why Docker?
- **Consistency**: Ensures applications run the same way across different environments
- **Isolation**: Each container runs in isolation with its own resources
- **Portability**: Easy to move applications between different environments
- **Scalability**: Simple to scale applications up or down
- **Version Control**: Easy to track and manage different versions of your application

## Docker Basics

### Installation
```bash
# For Ubuntu
sudo apt-get update
sudo apt-get install docker-ce docker-ce-cli containerd.io

# For macOS
brew install docker
```

### Basic Docker Concepts
- **Image**: A template for creating containers
- **Container**: A running instance of an image
- **Registry**: A repository for Docker images (e.g., Docker Hub)
- **Dockerfile**: A text file with instructions to build an image

## Dockerfile Creation

Here's an example Dockerfile for a Go application:

```dockerfile
# Use official Go runtime as base image
FROM golang:1.21-alpine

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o main .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]
```

## Docker Commands

### Basic Commands
```bash
# Build an image
docker build -t myapp:1.0 .

# Run a container
docker run -p 8080:8080 myapp:1.0

# List running containers
docker ps

# Stop a container
docker stop <container_id>

# Remove a container
docker rm <container_id>

# List images
docker images

# Remove an image
docker rmi <image_id>
```

## Docker Compose

Docker Compose is a tool for defining and running multi-container applications. Here's an example `docker-compose.yml`:

```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=db
    depends_on:
      - db

  db:
    image: postgres:13
    environment:
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=mydb
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

## AWS Deployment

### 1. Install AWS CLI and Configure
```bash
# Install AWS CLI
curl "https://awscli.amazonaws.com/AWSCLIV2.pkg" -o "AWSCLIV2.pkg"
sudo installer -pkg AWSCLIV2.pkg -target /

# Configure AWS CLI
aws configure
```

### 2. Create an ECR Repository
```bash
# Create repository
aws ecr create-repository --repository-name myapp

# Login to ECR
aws ecr get-login-password --region region | docker login --username AWS --password-stdin account.dkr.ecr.region.amazonaws.com
```

### 3. Build and Push Docker Image
```bash
# Build image
docker build -t myapp:1.0 .

# Tag image
docker tag myapp:1.0 account.dkr.ecr.region.amazonaws.com/myapp:1.0

# Push image
docker push account.dkr.ecr.region.amazonaws.com/myapp:1.0
```

### 4. Deploy to EC2
```bash
# Create EC2 instance
aws ec2 run-instances \
    --image-id ami-0c55b159cbfafe1f0 \
    --instance-type t2.micro \
    --key-name my-key-pair \
    --security-group-ids sg-xxxxxxxx

# Install Docker on EC2
sudo yum update -y
sudo yum install -y docker
sudo service docker start
sudo usermod -a -G docker ec2-user
```

### 5. Deploy to ECS (Elastic Container Service)
```yaml
# task-definition.json
{
    "family": "myapp",
    "containerDefinitions": [
        {
            "name": "myapp",
            "image": "account.dkr.ecr.region.amazonaws.com/myapp:1.0",
            "cpu": 256,
            "memory": 512,
            "portMappings": [
                {
                    "containerPort": 8080,
                    "hostPort": 8080
                }
            ]
        }
    ]
}
```

## Best Practices

1. **Use Multi-stage Builds**
```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main .

# Final stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
CMD ["./main"]
```

2. **Security Best Practices**
- Use non-root users
- Scan images for vulnerabilities
- Keep base images updated
- Use specific version tags
- Implement resource limits

3. **Performance Optimization**
- Use .dockerignore
- Leverage build cache
- Minimize layers
- Use appropriate base images

4. **Monitoring and Logging**
```bash
# View container logs
docker logs <container_id>

# Monitor container stats
docker stats

# Set up logging driver
docker run --log-driver=json-file --log-opt max-size=10m myapp:1.0
```

## Additional Resources

- [Docker Official Documentation](https://docs.docker.com/)
- [AWS Container Services](https://aws.amazon.com/containers/)
- [Docker Hub](https://hub.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/) 