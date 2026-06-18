# API Testing — Milestone 4: Team Management

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

---

# Player Management (Milestone 5)

## Setup: ensure team exists

```bash
# Re-create team 2 if you deleted it earlier
curl -s -X POST http://localhost:8080/api/teams \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"name":"Bali United","founded_year":2014,"headquarters_city":"Bali","manager_user_id":2}' | jq .
```

## Create players (ADMIN or MANAGER of the team)

```bash
# Admin creates players in team 1
curl -s -X POST http://localhost:8080/api/teams/1/players \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"name":"Player A","height_cm":178,"weight_kg":72,"position":"STRIKER","jersey_number":10}' | jq .

curl -s -X POST http://localhost:8080/api/teams/1/players \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"name":"Player B","height_cm":172,"weight_kg":68,"position":"DEFENDER","jersey_number":14}' | jq .

# Admin creates players in team 2
curl -s -X POST http://localhost:8080/api/teams/4/players \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"name":"Player C","height_cm":185,"weight_kg":78,"position":"STRIKER","jersey_number":9}' | jq .

# Manager creates a player in their own team (team 1) — should succeed
curl -s -X POST http://localhost:8080/api/teams/1/players \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $MANAGER_TOKEN" \
  -d '{"name":"Player D","height_cm":180,"weight_kg":76,"position":"GOALKEEPER","jersey_number":1}' | jq .

# Manager creates a player in a team they don't manage (team 2) — should return 403
curl -s -X POST http://localhost:8080/api/teams/4/players \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $MANAGER_TOKEN" \
  -d '{"name":"Hacker Player","height_cm":170,"weight_kg":65,"position":"MIDFIELDER","jersey_number":99}' | jq .
```

## Jersey uniqueness test

```bash
# Try to create a player with the same jersey number in the same team — should return 400
curl -s -X POST http://localhost:8080/api/teams/1/players \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"name":"Duplicate Jersey","height_cm":175,"weight_kg":70,"position":"MIDFIELDER","jersey_number":10}' | jq .

# Same jersey number in a different team — should succeed
curl -s -X POST http://localhost:8080/api/teams/2/players \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"name":"Different Team Same Jersey","height_cm":175,"weight_kg":70,"position":"MIDFIELDER","jersey_number":10}' | jq .
```

## List players by team

```bash
# List players in team 1
curl -s http://localhost:8080/api/teams/1/players \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

# List players in team 2
curl -s http://localhost:8080/api/teams/2/players \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
```

## Get player by ID

```bash
# Get player 1
curl -s http://localhost:8080/api/players/1 \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
```

## Update player (ADMIN or MANAGER of that team)

```bash
# Update player's position — should return 200
curl -s -X PUT http://localhost:8080/api/players/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"name":"Bambang Pamungkas","height_cm":178,"weight_kg":72,"position":"MIDFIELDER","jersey_number":10}' | jq .

# Manager updates player in their team (team 1) — should succeed
curl -s -X PUT http://localhost:8080/api/players/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $MANAGER_TOKEN" \
  -d '{"name":"Bambang Pamungkas","position":"STRIKER","jersey_number":10}' | jq .

# Manager updates player in a team they don't manage (team 2) — should return 403
curl -s -X PUT http://localhost:8080/api/players/3 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $MANAGER_TOKEN" \
  -d '{"name":"Ilija Spasojevic","position":"STRIKER","jersey_number":9}' | jq .
```

## Validation tests

```bash
# Invalid position — should return 400
curl -s -X POST http://localhost:8080/api/teams/1/players \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"name":"Invalid Player","position":"COACH","jersey_number":99}' | jq .

# Empty name — should return 400
curl -s -X POST http://localhost:8080/api/teams/1/players \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"name":"","position":"STRIKER","jersey_number":99}' | jq .
```

## User read-only test

```bash
# User reads players — should return 200
curl -s http://localhost:8080/api/teams/1/players \
  -H "Authorization: Bearer $USER_TOKEN" | jq .

# User tries to create a player — should return 403
curl -s -X POST http://localhost:8080/api/teams/1/players \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $USER_TOKEN" \
  -d '{"name":"User Player","position":"STRIKER","jersey_number":50}' | jq .
```

## Delete player (ADMIN or MANAGER of that team)

```bash
# ADMIN deletes player 4 (the one in team 2 with jersey 10) — should return 200
curl -s -X DELETE http://localhost:8080/api/players/4 \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

# Verify player 4 is gone
curl -s http://localhost:8080/api/players/4 \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

# List should show only remaining players
curl -s http://localhost:8080/api/teams/1/players \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
curl -s http://localhost:8080/api/teams/2/players \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
```
