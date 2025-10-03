# Docker Setup Guide

This guide explains how to build and run the ERP API using Docker and Docker Compose.

## Prerequisites

- Docker Engine 20.10+
- Docker Compose 2.0+

## Quick Start

### Development Environment

1. **Start all services with hot reload:**
   ```bash
   docker-compose -f docker-compose.dev.yml up
   ```

2. **Access the services:**
   - API: http://localhost:8081
   - PostgreSQL: localhost:5432
   - Redis: localhost:6379
   - PgAdmin: http://localhost:5050

### Production Environment

1. **Build and start all services:**
   ```bash
   docker-compose up -d --build
   ```

2. **View logs:**
   ```bash
   docker-compose logs -f api
   ```

## Available Services

### Main Services

- **api**: Go web API application
- **postgres**: PostgreSQL 15 database
- **redis**: Redis cache server

### Optional Services

- **pgadmin**: Web-based PostgreSQL management tool (profile: tools)

To start with optional tools:
```bash
docker-compose --profile tools up -d
```

## Environment Configuration

### Development (.env.dev)

Copy and modify `.env.dev` for local development:

```bash
cp .env.dev .env
# Edit .env with your preferences
```

### Production (.env.production)

For production, update `.env.production` with secure credentials:

```bash
# Update database passwords
# Update Redis password
# Update PgAdmin credentials
```

## Docker Commands

### Build and Run

```bash
# Build the API image
docker build -t erp-api .

# Build with docker-compose
docker-compose build

# Start all services
docker-compose up -d

# Start specific services
docker-compose up -d postgres redis
```

### Manage Services

```bash
# Stop all services
docker-compose down

# Stop and remove volumes (WARNING: deletes data)
docker-compose down -v

# Restart a service
docker-compose restart api

# View service logs
docker-compose logs -f api

# Execute commands in running container
docker-compose exec api sh
```

### Database Management

```bash
# Connect to PostgreSQL
docker-compose exec postgres psql -U erp_user -d erp_database

# Backup database
docker-compose exec postgres pg_dump -U erp_user erp_database > backup.sql

# Restore database
docker-compose exec -T postgres psql -U erp_user erp_database < backup.sql
```

### Development Workflow

```bash
# Development with hot reload
docker-compose -f docker-compose.dev.yml up

# Run tests
docker-compose exec api go test ./...

# Format code
docker-compose exec api go fmt ./...

# View API logs
docker-compose logs -f api
```

## Database Initialization

Create `scripts/init.sql` for initial database setup:

```sql
-- Example: Create tables, initial data, etc.
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

This file will be automatically executed when PostgreSQL starts for the first time.

## Health Checks

All services include health checks:

- **API**: `http://localhost:8081/health`
- **PostgreSQL**: `pg_isready` command
- **Redis**: `redis-cli ping`

Check service health:
```bash
docker-compose ps
```

## Troubleshooting

### Service Won't Start

1. Check logs:
   ```bash
   docker-compose logs service_name
   ```

2. Check if ports are already in use:
   ```bash
   netstat -tulpn | grep :8081
   ```

3. Remove old containers and volumes:
   ```bash
   docker-compose down -v
   docker-compose up -d
   ```

### Database Connection Issues

1. Verify PostgreSQL is healthy:
   ```bash
   docker-compose ps postgres
   ```

2. Check database credentials in `config.yml` and `.env`

3. Test connection:
   ```bash
   docker-compose exec postgres psql -U erp_user -d erp_database
   ```

### Performance Issues

1. Increase Docker resources (CPU, Memory) in Docker Desktop settings

2. Use production build instead of development:
   ```bash
   docker-compose up -d
   ```

## Production Deployment

### Security Checklist

- [ ] Change all default passwords in `.env.production`
- [ ] Use strong secrets for JWT tokens
- [ ] Enable SSL for database connections (`DB_SSLMODE=require`)
- [ ] Set API mode to `release` in `config.yml`
- [ ] Restrict CORS origins
- [ ] Use Docker secrets for sensitive data
- [ ] Enable firewall rules
- [ ] Regular security updates

### Deployment Steps

1. Update environment variables:
   ```bash
   cp .env.production .env
   # Edit .env with production values
   ```

2. Build and deploy:
   ```bash
   docker-compose up -d --build
   ```

3. Verify deployment:
   ```bash
   docker-compose ps
   docker-compose logs -f api
   ```

4. Setup monitoring and backups

## Additional Resources

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [PostgreSQL Docker Hub](https://hub.docker.com/_/postgres)
- [Go Docker Best Practices](https://docs.docker.com/language/golang/)

## Support

For issues or questions, please check the main README.md or create an issue in the repository.
