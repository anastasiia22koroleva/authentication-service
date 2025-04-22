# authentication-service

Simple Go-based authentication service implementing JWT access tokens and bcrypt-hashed refresh tokens with IP binding for enhanced security.

## introduction

This service provides two REST endpoints:

1. **giveTokens**: issues a pair of access and refresh tokens for a user by `user_id` (GUID).
2. **refreshTokens**: refreshes the pair of tokens, validates the provided refresh token, binds tokens to client IP, and logs warnings if IP changes.

## requirements

- Go 1.20+
- PostgreSQL 12+
- GNU Make (optional)
- Docker & docker-compose (optional)

## setup instructions

1. clone the repository:
   ```bash
   git clone https://github.com/your-username/authentication-service.git
   cd authentication-service
   ```
2. create a `.env` file in project root:
   ```dotenv
   DATABASE_URL=postgres://postgres:password@localhost:5432/authdb?sslmode=disable
   JWT_SECRET=your_jwt_secret_key
   ```
3. initialize database schema:
   ```bash
   psql "$DATABASE_URL" -f migrations/schema.sql
   ```

## usage examples

### give tokens

Issue tokens for a user:
```bash
curl -X POST "http://localhost:8080/giveTokens?user_id=3fa85f64-5717-4562-b3fc-2c963f66afa6"   -H "Content-Type: application/json"
```
Response:
```json
{
  "a_token": "<ACCESS_TOKEN>",
  "r_token": "<REFRESH_TOKEN>"
}
```

### refresh tokens

Refresh an existing pair:
```bash
curl -X POST http://localhost:8080/refreshTokens   -H "Content-Type: application/json"   -d '{
    "a_token": "<ACCESS_TOKEN>",
    "r_token": "<REFRESH_TOKEN>"
}'
```
Response:
```json
{
  "a_token": "<NEW_ACCESS_TOKEN>",
  "r_token": "<NEW_REFRESH_TOKEN>"
}
```

## environment variables

| name         | description                        |
|--------------|------------------------------------|
| DATABASE_URL | PostgreSQL connection string       |
| JWT_SECRET   | secret key for signing JWT tokens  |

## docker

Build and run with Docker:
```bash
# build image
docker build -t auth-service .
# run container
docker run --env-file .env -p 8080:8080 auth-service
```

Or with docker-compose:
```yaml
services:
  db:
    image: postgres:13
    environment:
      POSTGRES_DB: authdb
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
  web:
    build: .
    env_file: .env
    ports:
      - "8080:8080"
    depends_on:
      - db
```

Run:
```bash
docker-compose up -d
```

## running tests

Unit tests cover token generation and validation:
```bash
go test ./... -v
```
