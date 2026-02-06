# Feedback Backend Microservice

A feature-rich feedback management microservice built with Go and the Fiber framework. It provides feedback collection, voting, subscriptions, and moderation features for collecting and managing user feedback.

## Features

- **Feedback Management**
  - Create, read, update, delete feedback
  - Rich text descriptions with attachments
  - Category organization
  - Tag support

- **Voting System**
  - Upvote/downvote feedback
  - Vote score calculation
  - Trending feedback algorithm

- **Subscriptions**
  - Subscribe to specific feedback items
  - Subscribe to categories
  - Email and push notification preferences

- **Moderation**
  - Pending feedback approval workflow
  - Official responses from admins
  - Status tracking (pending, approved, planned, in progress, completed)

- **Statistics & Analytics**
  - Feedback statistics
  - Category analytics
  - Top contributors

## Tech Stack

- **Language**: Go 1.24
- **Framework**: Fiber v2
- **Database**: MongoDB
- **Auth**: JWT (via auth microservice)
- **SDK**: Uses comment microservice via go-sdk

## Quick Start

### Prerequisites

- Go 1.24+
- MongoDB 6+
- Auth microservice running
- Comment microservice running

### Installation

1. Navigate to the backend directory:
```bash
cd feedback/backend
```

2. Copy the environment file:
```bash
cp .env.example .env
```

3. Configure the `.env` file with your settings.

4. Run the service:
```bash
make run
```

Or with Docker:
```bash
docker-compose up -d
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | HTTP server port | `5012` |
| `SERVER_NAME` | Service name | `feedback-service` |
| `MONGO_URI` | MongoDB connection URI | `mongodb://localhost:27017` |
| `MONGO_DATABASE` | Database name | `feedback_db` |
| `AUTH_SERVICE_URL` | Auth service URL | `http://localhost:9001` |
| `COMMENT_SERVICE_URL` | Comment service URL | `http://localhost:5010` |
| `STORAGE_SERVICE_URL` | Storage service URL | `http://localhost:5004` |

See `.env.example` for all configuration options.

## API Endpoints

### Feedback

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/v1/feedback` | No | List feedback with filters |
| GET | `/api/v1/feedback/trending` | No | Get trending feedback |
| GET | `/api/v1/feedback/stats` | No | Get feedback statistics |
| GET | `/api/v1/feedback/:id` | No | Get feedback by ID |
| POST | `/api/v1/feedback` | Yes | Create new feedback |
| PUT | `/api/v1/feedback/:id` | Yes | Update feedback |
| DELETE | `/api/v1/feedback/:id` | Yes | Delete feedback |
| POST | `/api/v1/feedback/:id/vote` | Yes | Vote on feedback |

### Comments (via Comment Microservice)

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/v1/feedback/:id/comments` | No | List comments |
| GET | `/api/v1/feedback/:id/comments/stats` | No | Get comment statistics |
| POST | `/api/v1/feedback/:id/comments` | Yes | Create comment |
| PUT | `/api/v1/feedback/:id/comments/:commentId` | Yes | Update comment |
| DELETE | `/api/v1/feedback/:id/comments/:commentId` | Yes | Delete comment |
| POST | `/api/v1/feedback/:id/comments/:commentId/reactions` | Yes | Add reaction |
| DELETE | `/api/v1/feedback/:id/comments/:commentId/reactions/:type` | Yes | Remove reaction |

### Categories

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/v1/categories` | No | List categories |
| GET | `/api/v1/categories/:id` | No | Get category by ID |
| POST | `/api/v1/admin/categories` | Admin | Create category |
| PUT | `/api/v1/admin/categories/:id` | Admin | Update category |
| DELETE | `/api/v1/admin/categories/:id` | Admin | Delete category |

### Subscriptions

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/v1/subscriptions` | Yes | List subscriptions |
| POST | `/api/v1/subscriptions` | Yes | Create subscription |
| PUT | `/api/v1/subscriptions/:id` | Yes | Update subscription |
| DELETE | `/api/v1/subscriptions/:id` | Yes | Unsubscribe |
| GET | `/api/v1/feedback/:id/subscription` | Yes | Check subscription status |
| DELETE | `/api/v1/feedback/:id/subscription` | Yes | Unsubscribe from feedback |

### Admin

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/v1/admin/feedback/pending` | Admin | Get pending feedback |
| POST | `/api/v1/admin/feedback/:id/approve` | Admin | Approve feedback |
| POST | `/api/v1/admin/feedback/:id/reject` | Admin | Reject feedback |
| PUT | `/api/v1/admin/feedback/:id/status` | Admin | Update status |
| POST | `/api/v1/admin/feedback/:id/response` | Admin | Add official response |
| GET | `/api/v1/admin/dashboard` | Admin | Get dashboard stats |
| GET | `/api/v1/admin/settings` | Admin | Get all settings |
| PUT | `/api/v1/admin/settings` | Admin | Update setting |

### Settings

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/v1/settings` | No | Get public settings |

## Project Structure

```
feedback/backend/
├── cmd/                 # Application entry points
├── config/              # Configuration management
├── internal/
│   ├── handler/         # HTTP handlers
│   ├── models/          # Database models
│   ├── repository/      # Data access layer
│   ├── router/          # Route definitions
│   └── usecase/         # Business logic
├── .env.example         # Environment template
├── Dockerfile           # Container definition
├── Makefile             # Build commands
└── go.mod               # Dependencies
```

## Feedback Statuses

| Status | Description |
|--------|-------------|
| `pending` | Awaiting moderation |
| `approved` | Approved and visible |
| `rejected` | Rejected by moderator |
| `under_review` | Being reviewed |
| `planned` | Planned for implementation |
| `in_progress` | Currently being worked on |
| `completed` | Implemented |
| `closed` | Closed without action |

## Query Parameters

### List Feedback

| Parameter | Type | Description |
|-----------|------|-------------|
| `status` | string | Filter by status |
| `category_id` | string | Filter by category |
| `author_id` | string | Filter by author |
| `tags` | string | Filter by tags (comma-separated) |
| `search` | string | Search in title/description |
| `sort_by` | string | `new`, `top`, `trending`, `most_commented` |
| `sort_order` | string | `asc`, `desc` |
| `page` | int | Page number |
| `per_page` | int | Items per page |

## Integration with Comment Microservice

The feedback service uses the comment microservice via the go-sdk for all comment-related functionality. This provides:

- Comment CRUD operations
- Nested replies
- Reactions (like, dislike, etc.)
- Comment moderation

Comments are linked to feedback items using the feedback ID as the `resource_type` and `resource_id`.

## Development

### Run Tests

```bash
make test
```

### Build

```bash
make build
```

### Run Locally

```bash
make run
```

## Docker

Build and run with Docker:

```bash
# Build image
docker build -t feedback-service .

# Run with docker-compose
docker-compose up -d
```

## Health Checks

| Endpoint | Description |
|----------|-------------|
| `/health` | Basic health check |
| `/ready` | Readiness probe |

## License

MIT License - see LICENSE file for details.
