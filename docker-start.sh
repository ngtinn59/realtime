#!/bin/bash

# ERP API Docker Quick Start Script
# This script helps you quickly start the ERP API with Docker

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  ERP API Docker Quick Start${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Error: Docker is not installed${NC}"
    echo "Please install Docker from https://docs.docker.com/get-docker/"
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo -e "${RED}Error: Docker Compose is not installed${NC}"
    echo "Please install Docker Compose from https://docs.docker.com/compose/install/"
    exit 1
fi

# Check if .env file exists
if [ ! -f .env ]; then
    echo -e "${YELLOW}No .env file found. Creating from .env.dev...${NC}"
    cp .env.dev .env
    echo -e "${GREEN}✓ Created .env file${NC}"
fi

# Prompt user for environment
echo ""
echo "Select environment:"
echo "  1) Development (with hot reload)"
echo "  2) Production"
echo ""
read -p "Enter choice [1-2]: " choice

case $choice in
    1)
        echo -e "${GREEN}Starting development environment...${NC}"
        COMPOSE_FILE="docker-compose.dev.yml"
        ;;
    2)
        echo -e "${GREEN}Starting production environment...${NC}"
        COMPOSE_FILE="docker-compose.yml"
        
        # Check if .env.production exists
        if [ -f .env.production ]; then
            read -p "Use .env.production? [Y/n]: " use_prod_env
            if [[ $use_prod_env =~ ^[Yy]$ ]] || [ -z "$use_prod_env" ]; then
                cp .env.production .env
                echo -e "${GREEN}✓ Using .env.production${NC}"
            fi
        fi
        ;;
    *)
        echo -e "${RED}Invalid choice. Exiting.${NC}"
        exit 1
        ;;
esac

echo ""
echo -e "${YELLOW}Building and starting services...${NC}"
echo ""

# Build and start services
if [ "$choice" = "1" ]; then
    docker-compose -f $COMPOSE_FILE up
else
    docker-compose -f $COMPOSE_FILE up -d --build
    
    echo ""
    echo -e "${GREEN}✓ Services started successfully!${NC}"
    echo ""
    echo "Access your services at:"
    echo "  - API:      http://localhost:8081"
    echo "  - PgAdmin:  http://localhost:5050 (if tools profile enabled)"
    echo ""
    echo "Useful commands:"
    echo "  - View logs:     docker-compose logs -f api"
    echo "  - Stop services: docker-compose down"
    echo "  - View status:   docker-compose ps"
    echo ""
fi
