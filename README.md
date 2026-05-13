# Collab API


Real-time Collaboration Platform built with **Go** – entwickelt als Senior-Level Referenzprojekt.

Modernes Backend mit Fokus auf **Clean Architecture**, hoher Performance und Production-Readiness.

## Features

- **Real-time WebSocket** Kommunikation
- Chat-Nachrichten mit Persistenz
- Live Cursor Position Sharing
- Typing Indicator
- Online User Presence (Redis)
- JWT Authentication + Refresh Tokens
- Room Management (CRUD)
- Message History
- Rate Limiting
- Graceful Shutdown
- Strukturiertes Logging (Zap)

## Technologie-Stack

- **Go 1.26**
- **Gorilla Mux** + **Gorilla WebSocket**
- **PostgreSQL** + **pgx**
- **Redis** (Presence & Rate Limiting)
- **JWT** (golang-jwt)
- **Clean Architecture** (Domain-Driven Design)
- **Docker** + **docker-compose**
- **Goose** (Migrations)
- **Zap** Logging

## Projektstruktur

```bash
.
├── cmd/api/                # Einstiegspunkt
├── internal/
│   ├── domain/             # Business Entities & Interfaces
│   ├── application/        # Use Cases / Services
│   ├── adapter/            # Handler, Repository, WebSocket, Middleware
│   └── infrastructure/     # Config, DB, Redis, Logger
├── migrations/             # Datenbank-Migrationen
├── static/                 # Statische Dateien (Test-Client)
├── docker-compose.yml
├── Dockerfile
└── Makefile

```

## API Endpunkte


| Methode | Endpoint | Beschreibung |
| --- | --- | --- |
| POST | /api/v1/auth/login | Login (JWT Token) |
| POST | /api/v1/rooms | Neuen Room erstellen |
| GET | /api/v1/rooms | Alle Rooms abrufen |
| GET | /api/v1/rooms/{room_id} | Room Details |
| GET | /api/v1/rooms/{room_id}/members | Online User im Room |
| GET | /api/v1/rooms/{room_id}/messages | Message History |
| GET | /ws?token=...&room_id=... | WebSocket Verbindung |