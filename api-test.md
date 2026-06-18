# API Testing — Milestone 4: Team Management

## Login

```bash
ADMIN_TOKEN=$(curl -s -X POST http://localhost:8081/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@xyz-football.local","password":"ChangeMe123!"}' | jq -r '.token')

MANAGER_TOKEN=$(curl -s -X POST http://localhost:8081/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"manager@xyz-football.local","password":"ChangeMe123!"}' | jq -r '.token')

USER_TOKEN=$(curl -s -X POST http://localhost:8081/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@xyz-football.local","password":"ChangeMe123!"}' | jq -r '.token')
```

Seeded user IDs: Admin=1, Manager=2, User=3.

## Health check

```bash
curl -s http://localhost:8081/health | jq .
```

## Create teams (ADMIN only)

```bash
# Create team with a manager assigned
curl -s -X POST http://localhost:8081/api/teams \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"name":"Persija Jakarta","founded_year":1928,"headquarters_city":"Jakarta","manager_user_id":2}' | jq .

# Create team without a manager
curl -s -X POST http://localhost:8081/api/teams \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"name":"Bali United","founded_year":2014,"headquarters_city":"Bali"}' | jq .
```

## List all teams

```bash
curl -s http://localhost:8081/api/teams \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
```

## Get team by ID

```bash
curl -s http://localhost:8081/api/teams/1 \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
```

## Update team (ADMIN or MANAGER of that team)

```bash
# Manager updates their own team (team 1) — should return 200
curl -s -X PUT http://localhost:8081/api/teams/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $MANAGER_TOKEN" \
  -d '{"name":"Persija Jakarta - Updated","founded_year":1928,"headquarters_city":"Jakarta"}' | jq .

# Manager updates a team they don't manage (team 2) — should return 403
curl -s -X PUT http://localhost:8081/api/teams/2 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $MANAGER_TOKEN" \
  -d '{"name":"Hacked!","founded_year":2020}' | jq .
```

## Authorization boundary tests

```bash
# Manager tries to create a team — should return 403
curl -s -X POST http://localhost:8081/api/teams \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $MANAGER_TOKEN" \
  -d '{"name":"MANAGER Team","founded_year":2020}' | jq .

# User reads all teams — should return 200
curl -s http://localhost:8081/api/teams \
  -H "Authorization: Bearer $USER_TOKEN" | jq .

# User reads single team — should return 200
curl -s http://localhost:8081/api/teams/1 \
  -H "Authorization: Bearer $USER_TOKEN" | jq .

# User tries to create a team — should return 403
curl -s -X POST http://localhost:8081/api/teams \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $USER_TOKEN" \
  -d '{"name":"USER Team","founded_year":2020}' | jq .
```

## Delete team (ADMIN only)

```bash
# Delete team 2 — should return 200
curl -s -X DELETE http://localhost:8081/api/teams/2 \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

# Verify team 2 is gone (soft delete) — should return 404
curl -s http://localhost:8081/api/teams/2 \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

# List should show only team 1
curl -s http://localhost:8081/api/teams \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
```

---

# Player Management (Milestone 5)

## Setup: ensure team exists

```bash
# Re-create team 2 if you deleted it earlier
curl -s -X POST http://localhost:8081/api/teams \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"name":"Bali United","founded_year":2014,"headquarters_city":"Bali","manager_user_id":2}' | jq .
```

## Create players (ADMIN or MANAGER of the team)

```bash
# Admin creates players in team 1
curl -s -X POST http://localhost:8081/api/teams/1/players \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"name":"Player A","height_cm":178,"weight_kg":72,"position":"STRIKER","jersey_number":10}' | jq .

curl -s -X POST http://localhost:8081/api/teams/1/players \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"name":"Player B","height_cm":172,"weight_kg":68,"position":"DEFENDER","jersey_number":14}' | jq .

# Admin creates players in team 2
curl -s -X POST http://localhost:8081/api/teams/4/players \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"name":"Player C","height_cm":185,"weight_kg":78,"position":"STRIKER","jersey_number":9}' | jq .

# Manager creates a player in their own team (team 1) — should succeed
curl -s -X POST http://localhost:8081/api/teams/1/players \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $MANAGER_TOKEN" \
  -d '{"name":"Player D","height_cm":180,"weight_kg":76,"position":"GOALKEEPER","jersey_number":1}' | jq .

# Manager creates a player in a team they don't manage (team 2) — should return 403
curl -s -X POST http://localhost:8081/api/teams/4/players \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $MANAGER_TOKEN" \
  -d '{"name":"Hacker Player","height_cm":170,"weight_kg":65,"position":"MIDFIELDER","jersey_number":99}' | jq .
```

## Jersey uniqueness test

```bash
# Try to create a player with the same jersey number in the same team — should return 400
curl -s -X POST http://localhost:8081/api/teams/1/players \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"name":"Duplicate Jersey","height_cm":175,"weight_kg":70,"position":"MIDFIELDER","jersey_number":10}' | jq .

# Same jersey number in a different team — should succeed
curl -s -X POST http://localhost:8081/api/teams/2/players \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"name":"Different Team Same Jersey","height_cm":175,"weight_kg":70,"position":"MIDFIELDER","jersey_number":10}' | jq .
```

## List players by team

```bash
# List players in team 1
curl -s http://localhost:8081/api/teams/1/players \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

# List players in team 2
curl -s http://localhost:8081/api/teams/2/players \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
```

## Get player by ID

```bash
# Get player 1
curl -s http://localhost:8081/api/players/1 \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
```

## Update player (ADMIN or MANAGER of that team)

```bash
# Update player's position — should return 200
curl -s -X PUT http://localhost:8081/api/players/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"name":"Bambang Pamungkas","height_cm":178,"weight_kg":72,"position":"MIDFIELDER","jersey_number":10}' | jq .

# Manager updates player in their team (team 1) — should succeed
curl -s -X PUT http://localhost:8081/api/players/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $MANAGER_TOKEN" \
  -d '{"name":"Bambang Pamungkas","position":"STRIKER","jersey_number":10}' | jq .

# Manager updates player in a team they don't manage (team 2) — should return 403
curl -s -X PUT http://localhost:8081/api/players/3 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $MANAGER_TOKEN" \
  -d '{"name":"Ilija Spasojevic","position":"STRIKER","jersey_number":9}' | jq .
```

## Validation tests

```bash
# Invalid position — should return 400
curl -s -X POST http://localhost:8081/api/teams/1/players \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"name":"Invalid Player","position":"COACH","jersey_number":99}' | jq .

# Empty name — should return 400
curl -s -X POST http://localhost:8081/api/teams/1/players \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"name":"","position":"STRIKER","jersey_number":99}' | jq .
```

## User read-only test

```bash
# User reads players — should return 200
curl -s http://localhost:8081/api/teams/1/players \
  -H "Authorization: Bearer $USER_TOKEN" | jq .

# User tries to create a player — should return 403
curl -s -X POST http://localhost:8081/api/teams/1/players \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $USER_TOKEN" \
  -d '{"name":"User Player","position":"STRIKER","jersey_number":50}' | jq .
```

## Delete player (ADMIN or MANAGER of that team)

```bash
# ADMIN deletes player 4 (the one in team 2 with jersey 10) — should return 200
curl -s -X DELETE http://localhost:8081/api/players/4 \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

# Verify player 4 is gone
curl -s http://localhost:8081/api/players/4 \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

# List should show only remaining players
curl -s http://localhost:8081/api/teams/1/players \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
curl -s http://localhost:8081/api/teams/2/players \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
```

---

# Match Scheduling (Milestone 6)

Seeded team IDs: Mavericks=1, Dragon=2, Giants=3.

## Get seeded matches (any authenticated user)

```bash
# List all matches
curl -s http://localhost:8081/api/matches \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

# Get match by ID
curl -s http://localhost:8081/api/matches/1 \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
```

## Create a match (ADMIN only)

```bash
curl -s -X POST http://localhost:8081/api/matches \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"match_date":"2026-07-15","match_time":"20:00","home_team_id":1,"away_team_id":3}' | jq .
```

## Validation — same team

```bash
# Should return 400
curl -s -X POST http://localhost:8081/api/matches \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"match_date":"2026-07-15","match_time":"20:00","home_team_id":1,"away_team_id":1}' | jq .
```

## Validation — nonexistent team

```bash
# Should return 400
curl -s -X POST http://localhost:8081/api/matches \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"match_date":"2026-07-15","match_time":"20:00","home_team_id":1,"away_team_id":999}' | jq .
```

## Update a match (ADMIN only)

```bash
# Change match time
curl -s -X PUT http://localhost:8081/api/matches/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"match_date":"2026-06-20","match_time":"18:00"}' | jq .
```

## Authorization — MANAGER read only

```bash
# MANAGER reads matches — should return 200
curl -s http://localhost:8081/api/matches \
  -H "Authorization: Bearer $MANAGER_TOKEN" | jq .

# MANAGER tries to create a match — should return 403
curl -s -X POST http://localhost:8081/api/matches \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $MANAGER_TOKEN" \
  -d '{"match_date":"2026-07-15","match_time":"20:00","home_team_id":1,"away_team_id":2}' | jq .

# USER can also read — should return 200
curl -s http://localhost:8081/api/matches \
  -H "Authorization: Bearer $USER_TOKEN" | jq .
```

## Delete a match (ADMIN only)

```bash
# Delete match 4 (the one you just created)
curl -s -X DELETE http://localhost:8081/api/matches/4 \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

# Verify deletion (soft delete)
curl -s http://localhost:8081/api/matches/4 \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
```

---

# Match Result & Goals (Milestone 7)

Seeded match IDs: Mavericks vs Dragon = 1, Giants vs Mavericks = 2, Dragon vs Giants = 3.
Seeded players in match 1: Player A (ID 1, Team 1/Mavericks), Player B (ID 2, Team 1), Player E (ID 5, Team 2/Dragon).

## Check seeded goals

```bash
# List goals for match 1 (already finished with 3 seeded goals)
curl -s http://localhost:8081/api/matches/1/goals \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

# Verify match 1 is now FINISHED
curl -s http://localhost:8081/api/matches/1 \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
```

## Finish a match (ADMIN only)

```bash
# Finish match 2 (currently SCHEDULED)
curl -s -X PUT http://localhost:8081/api/matches/2/finish \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

# Try finishing it again — should return 400 (already finished)
curl -s -X PUT http://localhost:8081/api/matches/2/finish \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
```

## Record a goal (ADMIN only)

Player F (ID 6) is on Team 2 (Dragon). Match 2 is Giants (3) vs Mavericks (1).

```bash
# Record a goal for Player F (ID 6) in match 2 — should fail (wrong team)
curl -s -X POST http://localhost:8081/api/matches/2/goals \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"player_id":6,"goal_minute":15}' | jq .

# Record a goal for Player A (ID 1, Mavericks) in match 2 — should succeed
curl -s -X POST http://localhost:8081/api/matches/2/goals \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"player_id":1,"goal_minute":33}' | jq .

# Record a goal for Player I (ID 9, Giants) in match 2 — should succeed
curl -s -X POST http://localhost:8081/api/matches/2/goals \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"player_id":9,"goal_minute":78}' | jq .
```

## List goals for match 2

```bash
curl -s http://localhost:8081/api/matches/2/goals \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
```

## Validation — nonexistent player

```bash
# Player 999 doesn't exist
curl -s -X POST http://localhost:8081/api/matches/2/goals \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"player_id":999,"goal_minute":15}' | jq .
```

## Validation — invalid minute

```bash
curl -s -X POST http://localhost:8081/api/matches/2/goals \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"player_id":1,"goal_minute":0}' | jq .
```

## Authorization — MANAGER cannot record goals

```bash
curl -s -X POST http://localhost:8081/api/matches/1/goals \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $MANAGER_TOKEN" \
  -d '{"player_id":1,"goal_minute":10}' | jq .
```

## Delete a goal (ADMIN only)

```bash
# Delete goal 4 (the first goal you just added to match 2)
curl -s -X DELETE http://localhost:8081/api/matches/2/goals/4 \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

# Verify — only remaining goals for match 2
curl -s http://localhost:8081/api/matches/2/goals \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
```

---

# Match Report (Milestone 8)

Match 1 is seeded as FINISHED (Mavericks 2–1 Dragon). Match 2 was finished in the goal tests above.

## Get report for a finished match

```bash
# Report for match 1 (seeded as finished)
curl -s http://localhost:8081/api/matches/1/report \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .

# Report for match 2 (finished above)
curl -s http://localhost:8081/api/matches/2/report \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
```

## Report for an unfinished match — should return 400

```bash
# Match 3 is still SCHEDULED
curl -s http://localhost:8081/api/matches/3/report \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
```

## Report for nonexistent match — should return 404

```bash
curl -s http://localhost:8081/api/matches/999/report \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
```

## Any authenticated user can read reports

```bash
curl -s http://localhost:8081/api/matches/1/report \
  -H "Authorization: Bearer $MANAGER_TOKEN" | jq .

curl -s http://localhost:8081/api/matches/1/report \
  -H "Authorization: Bearer $USER_TOKEN" | jq .
```
