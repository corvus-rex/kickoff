# API Testing — Milestone 4: Team Management

## Prerequisites

```bash
# Start Postgres (Docker)
docker run -d --name kickoff-pg \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=kickoff_db \
  -p 5432:5432 postgres:16-alpine

# Start the Go server
go run ./cmd/api
```

## Login

```bash
ADMIN_TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@xyz-football.local","password":"ChangeMe123!"}' | jq -r '.token')

MANAGER_TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"manager@xyz-football.local","password":"ChangeMe123!"}' | jq -r '.token')

USER_TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@xyz-football.local","password":"ChangeMe123!"}' | jq -r '.token')
```

Seeded user IDs: Admin=1, Manager=2, User=3.

## Health check

```bash
curl -s http://localhost:8080/health | jq .
```

## Create teams (ADMIN only)

```bash
# Create team with a manager assigned
curl -s -X POST http://localhost:8080/api/teams \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"name":"Persija Jakarta","founded_year":1928,"headquarters_city":"Jakarta","manager_user_id":2}' | jq .

# Create team without a manager
curl -s -X POST http://localhost:8080/api/teams \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"name":"Bali United","founded_year":2014,"headquarters_city":"Bali"}' | jq .
```

## List all teams

```bash
curl -s http://localhost:8080/api/teams \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
```

## Get team by ID

```bash
curl -s http://localhost:8080/api/teams/1 \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
```

## Update team (ADMIN or MANAGER of that team)

```bash
# Manager updates their own team (team 1) — should return 200
curl -s -X PUT http://localhost:8080/api/teams/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $MANAGER_TOKEN" \
  -d '{"name":"Persija Jakarta - Updated","founded_year":1928,"headquarters_city":"Jakarta"}' | jq .

# Manager updates a team they don't manage (team 2) — should return 403
curl -s -X PUT http://localhost:8080/api/teams/2 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $MANAGER_TOKEN" \
  -d '{"name":"Hacked!","founded_year":2020}' | jq .
```

## Authorization boundary tests

```bash
# Manager tries to create a team — should return 403
curl -s -X POST http://localhost:8080/api/teams \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $MANAGER_TOKEN" \
  -d '{"name":"MANAGER Team","founded_year":2020}' | jq .

# User reads all teams — should return 200
curl -s http://localhost:8080/api/teams \
  -H "Authorization: Bearer $USER_TOKEN" | jq .

# User reads single team — should return 200
curl -s http://localhost:8080/api/teams/1 \
  -H "Authorization: Bearer $USER_TOKEN" | jq .

# User tries to create a team — should return 403
curl -s -X POST http://localhost:8080/api/teams \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $USER_TOKEN" \
  -d '{"name":"USER Team","founded_year":2020}' | jq .
```

## Delete team (ADMIN only)

```bash
# Delete team 2 — should return 200
curl -s -X DELETE http://localhost:8080/api/teams/2 \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

# Verify team 2 is gone (soft delete) — should return 404
curl -s http://localhost:8080/api/teams/2 \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

# List should show only team 1
curl -s http://localhost:8080/api/teams \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
```
