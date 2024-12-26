# Task Tracker

A lightweight HTTP server for tracking tasks and their completion status, written in Go. This project supports CRUD operations for tasks, is fully Dockerized, and is deployed on Fly.io.

---

## Features
- Create, read, update, and delete tasks.
- Persistent task storage with JSON files.
- Dockerized for easy deployment.
- Dynamic port configuration via environment variables.
- Deployed to Fly.io with logging and autoscaling support.

---

## Setup

### Prerequisites
- [Go](https://golang.org/doc/install) (for running locally)
- [Docker](https://www.docker.com/) (for containerized execution)
- [Fly.io CLI](https://fly.io/docs/getting-started/installing-flyctl/) (for deployment)

---

## Running Locally

1. Clone the repository:
   ```bash
   git clone https://github.com/sirthus/task-tracker.git
   cd task-tracker
   ```

2. Run the app (default port: `8000`):
   ```bash
   go run main.go middleware.go
   ```

3. Test endpoints:
   - Health check: `http://localhost:8000/tasks/health`
   - List tasks: `http://localhost:8000/tasks`

4. Run on a custom port:
   ```bash
   PORT=8080 go run main.go middleware.go
   ```

---

## Running with Docker

1. Build the Docker image:
   ```bash
   docker build -t task-tracker .
   ```

2. Run the Docker container:
   ```bash
   docker run -p 8000:8000 task-tracker
   ```

3. Test endpoints:
   - Health check: `http://localhost:8000/tasks/health`
   - List tasks: `http://localhost:8000/tasks`

4. Run on a custom port:
   ```bash
   docker run -e PORT=8080 -p 8080:8080 task-tracker
   ```

---

## Deployment

The app is deployed on Fly.io and accessible at:
[https://task-tracker-winter-grass-1987.fly.dev](https://task-tracker-winter-grass-1987.fly.dev)

To redeploy:
1. Build and deploy using Fly.io:
   ```bash
   flyctl deploy
   ```

2. Monitor logs:
   ```bash
   flyctl logs
   ```

---

## API Endpoints

### Base URL:
- Locally: `http://localhost:<PORT>`
- Fly.io: `https://task-tracker-winter-grass-1987.fly.dev`

### Endpoints:
| Method | Endpoint              | Description                   |
|--------|-----------------------|-------------------------------|
| GET    | `/tasks`             | Retrieve all tasks            |
| POST   | `/tasks`             | Add a new task                |
| PUT    | `/tasks/{id}`        | Update an existing task       |
| DELETE | `/tasks/{id}`        | Delete a task by ID           |
| GET    | `/tasks/health`      | Health check for the app      |

---

## Monitoring & Logs

### Fly.io Logs
View real-time logs using:
```bash
flyctl logs
```

---

## Repository

GitHub: [Task Tracker](https://github.com/sirthus/task-tracker)

---
