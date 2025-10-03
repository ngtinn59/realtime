# 💬 Chat Realtime API - Golang WebSocket Chat System

A modern real-time chat system built with Go, Gin framework, WebSocket, GORM ORM, Redis, and PostgreSQL. Supports private messaging, group chat, file sharing, and real-time presence.

## ✨ Features

### Core Features
- 💬 **Private Messaging** - One-on-one real-time chat
- 👥 **Group Chat** - Create and manage group conversations
- 📎 **File Sharing** - Upload and share files (images, documents)
- 🔔 **Real-time Notifications** - Instant message delivery via WebSocket
- ✅ **Read Receipts** - Track message read status
- 👀 **Typing Indicators** - See when others are typing
- 🟢 **Online Presence** - Real-time user online/offline status
- 🔍 **User Search** - Find users by username or email
- 📊 **Conversation List** - View all active conversations

### Technical Features
- 🔐 **JWT Authentication** - Secure token-based auth
- 🔌 **WebSocket Support** - Real-time bidirectional communication
- 📦 **Redis Caching** - Fast online user tracking
- 🗄️ **PostgreSQL Database** - Reliable data persistence
- 🐳 **Docker Support** - Easy deployment with Docker Compose
- 🔥 **Hot Reload** - Fast development with Air
- 📝 **API Documentation** - Comprehensive API docs
- 🧪 **Clean Architecture** - Maintainable and scalable code

## 🚀 Quick Start

### Using Docker (Recommended)

```bash
# Development with hot reload
make dev-up

# Production
make docker-up
```

For detailed Docker instructions, see [DOCKER_README.md](./DOCKER_README.md)

### Local Development

1. **Install dependencies:**
   ```bash
   go mod download
   ```

2. **Run the project:**
   ```bash
   make run
   ```

3. **Auto reload with Air:**
   ```bash
   air
   ```

## 📋 Prerequisites

- Go 1.19 or higher
- Docker & Docker Compose (for containerized deployment)
- Make (recommended)

[Make for Windows](https://gnuwin32.sourceforge.net/packages/make.htm)

## 🛠️ Available Commands

Run `make help` to see all available commands:

### Local Development
- `make build` - Build the Go application
- `make run` - Run the application locally
- `make format` - Format Go code
- `make test` - Run tests

### Docker Commands
- `make docker-up` - Start all services
- `make docker-down` - Stop all services
- `make docker-logs` - View API logs
- `make dev-up` - Start dev environment with hot reload

### Database Commands
- `make db-connect` - Connect to PostgreSQL
- `make db-backup` - Backup database
- `make db-restore FILE=backup.sql` - Restore database

## 🗂️ Project Structure

```
.
├── cmd/                          # Application entrypoint
│   └── main.go                  # Main application file
├── internal/                     # Internal application code
│   ├── api/                     # API layer
│   │   ├── controllers/         # Request handlers
│   │   ├── middlewares/         # Middleware functions
│   │   ├── routers/            # Route definitions
│   │   ├── services/           # Business logic
│   │   └── startup.go          # Application startup
│   └── pkg/                    # Internal packages
│       ├── config/             # Configuration management
│       ├── database/           # Database connection
│       ├── models/             # Data models
│       ├── redis/              # Redis client
│       ├── utils/              # Utility functions
│       └── websocket/          # WebSocket hub & client
├── pkg/                         # Public packages
│   └── logger/                 # Logging utilities
├── data/                        # Configuration files
│   └── config.yml              # App configuration
├── scripts/                     # Database scripts
│   └── init.sql                # Initial database setup
├── log/                         # Application logs
├── uploads/                     # Uploaded files
├── Dockerfile                   # Production Dockerfile
├── docker-compose.yml          # Production compose
├── docker-compose.dev.yml      # Development compose
└── API_DOCUMENTATION.md        # API documentation
```

## 🔧 Configuration

### Environment Files

- `.env.dev` - Development environment variables
- `.env.production` - Production environment variables
- `.env.example` - Template for environment variables

### Database Configuration

Edit `data/config.yml` to configure:
- Server settings (port, secret, mode)
- CORS settings
- Database connection (driver, host, credentials)

Supported database drivers:
- PostgreSQL (default)
- MySQL
- SQLite
- SQL Server

## 🐳 Docker Services

The docker-compose setup includes:
- **API**: Go web application (port 8081)
- **PostgreSQL**: Database server (port 5432)
- **Redis**: Cache server (port 6379)
- **PgAdmin**: Database management tool (port 5050, optional)

## 📚 Documentation

### Complete Documentation
- 📖 [**API Documentation**](./API_DOCUMENTATION.md) - Complete REST API reference with examples
- 🗄️ [**Database Schema**](./DATABASE_SCHEMA.md) - Database tables, relationships, and queries
- 🔌 [**WebSocket Events**](./WEBSOCKET_EVENTS.md) - Real-time WebSocket protocol and events
- 🐳 [**Docker Setup Guide**](./DOCKER_README.md) - Docker deployment instructions

### Quick Links
- **API Base URL**: `http://localhost:8081/api`
- **WebSocket URL**: `ws://localhost:8081/ws?token=<JWT_TOKEN>`
- **File Uploads**: `http://localhost:8081/uploads/`

### Key Endpoints
```bash
# Authentication
POST   /api/register          # Register new user
POST   /api/login             # Login user
GET    /api/profile           # Get user profile

# Private Messages
POST   /api/messages/private  # Send private message
GET    /api/messages/private/:userID  # Get conversation
GET    /api/conversations     # List all conversations

# Group Chat
POST   /api/groups/create     # Create new group
POST   /api/messages/group    # Send group message
GET    /api/groups            # List user groups

# Files
POST   /api/files/upload      # Upload file
GET    /api/files             # List user files

# WebSocket
GET    /ws?token=<JWT>        # Connect to WebSocket
```

## 🔒 Security

- Change default passwords in production
- Use environment variables for sensitive data
- Enable SSL for database connections in production
- Set API mode to "release" in production

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Submit a pull request

## 📝 License

This project is licensed under the MIT License.
