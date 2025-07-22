# Go Chat Room Application

A real-time chat room application built with Go, Gin, and PostgreSQL. This project demonstrates a scalable backend chat service with message persistence, client management, and a simple REST API. It is containerized with Docker and orchestrated using Docker Compose.

## Features
- Real-time chat room with multiple clients
- RESTful API for joining, leaving, sending, and retrieving messages
- Message persistence using PostgreSQL
- In-memory message cache for fast retrieval
- Automatic cleanup of inactive clients
- Graceful shutdown and health check endpoint
- Fully containerized for easy deployment

## Architecture
- **Go (Gin):** HTTP server and API endpoints
- **GORM:** ORM for PostgreSQL
- **PostgreSQL:** Message storage
- **Docker & Docker Compose:** Containerization and orchestration

### Main Components
- `main.go`: Entry point, server setup, API routing
- `internal/chat/`: Chat room logic, client management, message broadcasting
- `internal/models/`: Data models (Message)
- `internal/config/`: Configuration loading
- `db/init.sql`: Database schema

## Getting Started

### Prerequisites
- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/)

### Quick Start (Docker Compose)

```sh
git clone https://github.com/sreenathsvrm/chat-room.git
cd chat-room
docker-compose up --build
```

- The API will be available at `http://localhost:8080`
- PostgreSQL will be available at `localhost:5432` (default user/password: `postgres`)

### Local Development (without Docker)
1. Install Go 1.24+
2. Install and run PostgreSQL (create a database named `chat`)
3. Set environment variables (see below)
4. Run migrations: `psql -U postgres -d chat -f db/init.sql`
5. Build and run:
   ```sh
   go mod download
   go run app/main.go
   ```

## Configuration
Configuration is loaded from environment variables (or a `.env` file):

| Variable      | Default     | Description                |
|--------------|-------------|----------------------------|
| DB_HOST      | localhost   | Database host              |
| DB_PORT      | 5432        | Database port              |
| DB_USER      | postgres    | Database user              |
| DB_PASSWORD  | postgres    | Database password          |
| DB_NAME      | chat        | Database name              |

## API Endpoints

All endpoints are prefixed with `/api`.

### Join Room
- **POST** `/api/join`
- **Body:** `{ "client_id": "string" }`
- **Response:** `{ "status": "success", "client_id": "string" }`

### Leave Room
- **POST** `/api/leave`
- **Body:** `{ "client_id": "string" }`
- **Response:** `{ "status": "success" }`

### Send Message
- **POST** `/api/send`
- **Body:** `{ "client_id": "string", "message": "string" }`
- **Response:** `{ "status": "success" }`

### Get Messages (Long Polling)
- **GET** `/api/messages?client_id=string&since=unix_timestamp`
- **Response:** `{ "messages": [ { "id": 1, "sender_id": "string", "message": "string", "created_at": "timestamp" }, ... ] }`
- `since` is optional (returns all messages if omitted)

### Health Check
- **GET** `/health`
- **Response:** `{ "status": "ok" }`

## Database Schema

```sql
CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    sender_id VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
```

## Project Structure
```
├── app/
│   ├── Dockerfile
│   ├── main.go
│   └── internal/
│       ├── chat/
│       │   ├── client.go
│       │   ├── repository.go
│       │   └── room.go
│       ├── config/
│       │   └── config.go
│       └── models/
│           └── message.go
├── db/
│   └── init.sql
├── docker-compose.yml
├── go.mod
├── go.sum
```

## License
MIT 