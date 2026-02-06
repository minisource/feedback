# Minisource Feedback

Multi-tenant feedback and feature request management platform. Enables users to submit feedback, vote on features, and track product improvements.

## Repository Structure

This is a monorepo containing three projects:

```
feedback/
├── backend/      # Go API server (Fiber + MongoDB)
├── admin/        # Admin dashboard (Planned)
└── user/         # User-facing feedback portal (Next.js)
```

## Projects

### Backend (API Server)

The backend service provides REST APIs for feedback management.

- **Tech Stack**: Go 1.24, Fiber v2, MongoDB
- **Port**: 5012
- **Features**: Multi-tenant, OAuth2, voting, subscriptions, moderation

```bash
cd backend
cp .env.example .env
make run
```

See [backend/README.md](backend/README.md) for detailed documentation.

### User Portal (Frontend)

User-facing web application for submitting and viewing feedback.

- **Tech Stack**: Next.js 15, React, TypeScript, Tailwind CSS
- **Port**: 3002

```bash
cd user
cp .env.example .env
npm install
npm run dev
```

### Admin Dashboard (Planned)

Administration dashboard for managing feedback and moderation.

- **Status**: 🚧 Not yet implemented
- **Planned Tech**: Next.js, React Admin

## Quick Start

### With Docker Compose

```bash
# Start all services
docker-compose -f backend/docker-compose.yml up -d

# Start user portal
cd user && docker-compose up -d
```

### Development Setup

```bash
# 1. Start backend
cd backend
cp .env.example .env
make docker-up

# 2. Start user portal
cd ../user
cp .env.example .env
npm install
npm run dev
```

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    User Portal (:3002)                   │
│                     (Next.js)                            │
└─────────────────────────┬───────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                 Feedback Backend (:5012)                 │
│                    (Go + Fiber)                          │
├─────────────────────────────────────────────────────────┤
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐    │
│  │Feedback │  │ Voting  │  │Category │  │Moderate │    │
│  │ CRUD    │  │ System  │  │ Mgmt    │  │  Queue  │    │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘    │
└─────────────────────────┬───────────────────────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        ▼                 ▼                 ▼
  ┌───────────┐    ┌───────────┐     ┌───────────┐
  │  MongoDB  │    │   Auth    │     │  Notifier │
  │           │    │  Service  │     │  Service  │
  └───────────┘    └───────────┘     └───────────┘
```

## Features

### User Features
- Submit feedback with categories
- Vote on feedback (upvote/downvote)
- Subscribe to feedback updates
- View public roadmap
- Comment on feedback

### Admin Features
- Moderate feedback submissions
- Manage categories
- Official responses
- Analytics dashboard
- Export functionality

## API Overview

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/feedback` | GET | List feedback |
| `/api/v1/feedback` | POST | Create feedback |
| `/api/v1/feedback/:id` | GET | Get feedback |
| `/api/v1/feedback/:id/vote` | POST | Vote on feedback |
| `/api/v1/feedback/:id/subscribe` | POST | Subscribe |
| `/api/v1/categories` | GET | List categories |
| `/api/v1/admin/moderate` | POST | Moderate feedback |

## Configuration

### Backend Environment Variables

| Variable | Description |
|----------|-------------|
| `MONGODB_URI` | MongoDB connection string |
| `AUTH_URL` | Auth service URL |
| `NOTIFIER_URL` | Notifier service URL |
| `SERVER_PORT` | API server port (5012) |

### User Portal Environment Variables

| Variable | Description |
|----------|-------------|
| `NEXT_PUBLIC_API_URL` | Backend API URL |
| `NEXT_PUBLIC_AUTH_URL` | Auth service URL |

## Development

### Prerequisites

- Go 1.24+ (for backend)
- Node.js 20+ (for frontends)
- MongoDB 7+
- Docker & Docker Compose

### Running Tests

```bash
# Backend tests
cd backend
make test

# User portal tests
cd user
npm test
```

## License

MIT