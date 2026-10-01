# Chirpy

A lightweight RESTful social media API built with Go, PostgreSQL, and standard HTTP libraries.

## Overview

Chirpy is a backend service demonstrating core REST API principles, database persistence, authentication, and webhook handling. Users can register, log in, post "chirps", and upgrade to premium membership tiers.

## Features

- **User Authentication:** Secure password hashing (bcrypt) and JWT-based session tokens.
- **Refresh Tokens:** Long-lived tokens stored in the database for issuing new access tokens.
- **Chirps CRUD:** Endpoints to create, read, and delete short text posts.
- **Content Filtering:** Automatic censorship of profanity.
- **Webhooks:** Secure endpoint accepting simulated external payment notifications to upgrade user accounts.

## Tech Stack

- **Language:** Go
- **Database:** PostgreSQL
- **Migrations & Queries:** Goose / SQLC
- **Authentication:** `golang-jwt/jwt`, `golang.org/x/crypto/bcrypt`

## Prerequisites

- [Go](https://go.dev/) (v1.22+)
- [PostgreSQL](https://www.postgresql.org/)

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/paugomez86/chirpy
cd chirpy
```

### 2. Configure Environment Variables

Create a `.env` file in the root directory:

```env
PORT=8080
DB_URL=postgres://user:password@localhost:5432/chirpy?sslmode=disable
JWT_SECRET=your_secret_jwt_key
POLKA_KEY=your_polka_webhook_key
PLATFORM=dev
```

### 3. Run Database Migrations

```bash
goose postgres "postgres://user:password@localhost:5432/chirpy?sslmode=disable" up
```

### 4. Run the Server

```bash
go run .
```

The server will start listening on `http://localhost:8080`.

## API Documentation

| Method    | Endpoint              | Description                   | Auth Required     | Additional Params |
| ---       | ---                   | ---                           | ---               | ---               |
| `GET`     | `/api/healthz`        | Health check                  | No                | None              |
| `POST`    | `/api/users`          | Register a new user           | No                | None              |
| `GET`     | `/api/users`          | List users                    | No                | None              |
| `PUT`     | `/api/users`          | Updates a user                | Bearer Token      | None              |
| `POST`    | `/api/login`          | Log in and receive tokens     | No                | None              |
| `POST`    | `/api/chirps`         | Create a new chirp            | Bearer Token      | None              |
| `GET`     | `/api/chirps`         | List chirps                   | No                | author_id<br>sort |
| `GET`     | `/api/chirps/{id}`    | Get chirp by ID               | No                | None              |
| `DELETE`  | `/api/chirps/{id}`    | Delete a chirp                | Bearer Token      | None              |
| `POST`    | `/api/polka/webhooks` | Handle upgrade webhook        | API Key           | None              |
| `POST`    | `/api/refresh`        | Receive a new access token    | Bearer Token      | None              |
| `POST`    | `/api/revoke`         | Revokes a refresh token       | Bearer Token      | None              |


## License

This project is licensed under the MIT License.