# Dokumentasi API Kickoff

Sistem manajemen tim sepak bola berbasis REST API. Dibangun dengan Go, Gin, PostgreSQL, dan JWT Authentication.

---

## Setup Lokal

### Prasyarat

- Go 1.25+
- PostgreSQL 17+
- Docker & Docker Compose

```bash
# Clone repositori
git clone https://github.com/corvus-rex/kickoff.git
cd kickoff

# Jalankan dengan Docker Compose
docker compose up --build
```
---

## Arsitektur

### Komponen Utama

- **Gin Framework** - HTTP router dan middleware. Menangani request/response JSON.
- **JWT Middleware** - Memvalidasi token Bearer pada setiap request ke `/api/*`. Menyimpan `userID`, `userRole`, dan `userEmail` ke dalam konteks Gin.
- **Role Middleware** - Memeriksa role pengguna sebelum mengizinkan akses ke endpoint tertentu.
- **Handler Layer** - Menerima request HTTP, mem-parsing parameter dan body JSON, memanggil service, mengembalikan response.
- **Service Layer** - Business logic, validasi domain, pengecekan otorisasi berbasis kepemilikan tim.
- **Repository Layer** - Akses data melalui GORM ORM. Memisahkan query database dari logika bisnis.
- **GORM** - ORM Go dengan dukungan soft delete (`deleted_at`), migrasi otomatis, dan transaction.
- **PostgreSQL** - Database relasional.

### Pola Desain

Layered architecture (Handler -> Service -> Repository) dengan dependency injection melalui constructor

---


## Struktur Folder

```
cmd/
  api/
    main.go                  - Entry point aplikasi

internal/
  auth/
    model.go                 - Struct User, Role constants
    handler.go               - LoginHandler, RegisterRoutes
    service.go               - (tidak ada, logika sederhana di handler)
    repository.go            - FindUserByEmail
    password.go              - HashPassword, ComparePassword (bcrypt)
    jwt.go                   - GenerateToken, ParseToken, Claims struct
    middleware.go            - Middleware (JWT), RequireRole, context keys
    seed.go                  - Seed default users
    auth_test.go             - Unit test autentikasi

  config/
    config.go                - Konfigurasi dari environment variables

  database/
    connection.go            - Koneksi PostgreSQL via GORM
    migrate.go               - Registrasi model dan auto-migrate

  team/
    model.go                 - Struct Team
    handler.go               - CRUD handler, RegisterRoutes
    service.go               - Business logic + otorisasi kepemilikan
    repository.go            - Akses data tim
    team_test.go             - Unit test tim

  player/
    model.go                 - Struct Player, Position enum
    handler.go               - CRUD handler, RegisterRoutes
    service.go               - Business logic + verifikasi akses tim
    repository.go            - Akses data pemain + cek jersey unik
    player_test.go           - Unit test pemain

  match/
    model.go                 - Struct Match, MatchStatus enum
    handler.go               - CRUD handler, FinishMatch, RegisterRoutes
    service.go               - Business logic + validasi jadwal
    repository.go            - Akses data pertandingan
    match_test.go            - Unit test pertandingan

  goal/
    model.go                 - Struct Goal
    handler.go               - CRUD handler, RegisterRoutes
    service.go               - Business logic + validasi pencetak gol
    repository.go            - Akses data gol
    goal_test.go             - Unit test gol

  report/
    model.go                 - Struct MatchReport, TeamInfo, Scorer
    handler.go               - GetReport handler, RegisterRoutes
    service.go               - Logika perhitungan skor, top scorer, cumulative wins
    report_test.go           - Unit test laporan

  seed/
    seed.go                  - Seed data domain (tim, pemain, pertandingan, gol)

  testutil/
    testutil.go              - Helper function untuk integration test

Dockerfile                   - Multi-stage build
docker-compose.yml           - PostgreSQL + API services
.env                         - Environment variables (lokal)
.dockerignore                - Exclude file dari Docker build context
docs.md                      - Dokumentasi (file ini)
README.md                    - Panduan cepat
api-test.md                  - Contoh curl command
```

---

## Skema Database


### Tabel: `users`

| Kolom          | Tipe                | Keterangan                                  |
|----------------|---------------------|---------------------------------------------|
| id             | BIGSERIAL (PK)      | Primary key                                 |
| name           | VARCHAR, NOT NULL   | Nama lengkap                                |
| email          | VARCHAR, UNIQUE, NN | Email (case-insensitive saat login)         |
| password_hash  | VARCHAR, NOT NULL   | Hash bcrypt                                 |
| role           | VARCHAR(20), NOT NULL | ADMIN / MANAGER / USER                    |
| created_at     | TIMESTAMP           |                                             |
| updated_at     | TIMESTAMP           |                                             |
| deleted_at     | TIMESTAMP (soft delete) | Index                                  |

Constraint: `CHECK (role IN ('ADMIN','MANAGER','USER'))`

### Tabel: `teams`

| Kolom                  | Tipe              | Keterangan                              |
|------------------------|-------------------|-----------------------------------------|
| id                     | BIGSERIAL (PK)    |                                         |
| name                   | VARCHAR, UNIQUE   | Nama tim (unik)                         |
| logo_url               | TEXT              | URL logo                                |
| founded_year           | INTEGER, NOT NULL | Tahun berdiri                           |
| headquarters_address   | TEXT              | Alamat markas                           |
| headquarters_city      | VARCHAR           | Kota markas                             |
| manager_user_id        | BIGINT (FK→users) | Manager tim (nullable)                  |
| created_at             | TIMESTAMP         |                                         |
| updated_at             | TIMESTAMP         |                                         |
| deleted_at             | TIMESTAMP (soft delete) | Index                              |

### Tabel: `players`

| Kolom          | Tipe                | Keterangan                                  |
|----------------|---------------------|---------------------------------------------|
| id             | BIGSERIAL (PK)      |                                             |
| team_id        | BIGINT (FK→teams)   | Tim tempat pemain bermain                   |
| name           | VARCHAR, NOT NULL   | Nama pemain                                 |
| height_cm      | DOUBLE PRECISION    | Tinggi badan (cm)                           |
| weight_kg      | DOUBLE PRECISION    | Berat badan (kg)                            |
| position       | VARCHAR(20), NOT NULL | STRIKER / MIDFIELDER / DEFENDER / GOALKEEPER / FLEX |
| jersey_number  | INTEGER, NOT NULL   | Nomor punggung, unik                              |
| created_at     | TIMESTAMP           |                                             |
| updated_at     | TIMESTAMP           |                                             |
| deleted_at     | TIMESTAMP (soft delete) | Index                                  |

Constraint: `UNIQUE (team_id, jersey_number)` - kombinasi `idx_team_jersey`

### Tabel: `matches`

| Kolom          | Tipe                | Keterangan                                  |
|----------------|---------------------|---------------------------------------------|
| id             | BIGSERIAL (PK)      |                                             |
| match_date     | DATE, NOT NULL      | Tanggal pertandingan (YYYY-MM-DD)           |
| match_time     | VARCHAR(5), NOT NULL | Waktu pertandingan (HH:MM)                 |
| home_team_id   | BIGINT (FK→teams)   | Tim tuan rumah                              |
| away_team_id   | BIGINT (FK→teams)   | Tim tamu                                    |
| status         | VARCHAR(20), NOT NULL | SCHEDULED (default) / FINISHED            |
| created_at     | TIMESTAMP           |                                             |
| updated_at     | TIMESTAMP           |                                             |
| deleted_at     | TIMESTAMP (soft delete) | Index                                  |

### Tabel: `goals`

| Kolom          | Tipe                | Keterangan                                  |
|----------------|---------------------|---------------------------------------------|
| id             | BIGSERIAL (PK)      |                                             |
| match_id       | BIGINT (FK→matches) | Pertandingan                                |
| player_id      | BIGINT (FK→players) | Pencetak gol                                |
| goal_minute    | INTEGER, NOT NULL   | Menit terjadinya gol                        |
| created_at     | TIMESTAMP           |                                             |
| updated_at     | TIMESTAMP           |                                             |
| deleted_at     | TIMESTAMP (soft delete) | Index                                  |

Semua tabel menggunakan soft delete (`deleted_at`). Query GORM secara otomatis menambahkan `WHERE deleted_at IS NULL` pada operasi normal.

---

## Aturan Otorisasi

Tiga role: **ADMIN**, **MANAGER**, **USER**.

### Ringkasan Izin per Endpoint

| Endpoint | Method | ADMIN | MANAGER | USER |
|---|---|---|---|---|
| `/auth/login` | POST | ✅ | ✅ | ✅ |
| `GET /api/teams` | GET | ✅ | ✅ | ✅ |
| `GET /api/teams/:team_id` | GET | ✅ | ✅ | ✅ |
| `POST /api/teams` | POST | ✅ | ❌ | ❌ |
| `PUT /api/teams/:team_id` | PUT | ✅ | ✅ (own) | ❌ |
| `DELETE /api/teams/:team_id` | DELETE | ✅ | ❌ | ❌ |
| `GET /api/teams/:team_id/players` | GET | ✅ | ✅ | ✅ |
| `POST /api/teams/:team_id/players` | POST | ✅ | ✅ (own team) | ❌ |
| `GET /api/players/:id` | GET | ✅ | ✅ | ✅ |
| `PUT /api/players/:id` | PUT | ✅ | ✅ (own team) | ❌ |
| `DELETE /api/players/:id` | DELETE | ✅ | ✅ (own team) | ❌ |
| `GET /api/matches` | GET | ✅ | ✅ | ✅ |
| `GET /api/matches/:match_id` | GET | ✅ | ✅ | ✅ |
| `POST /api/matches` | POST | ✅ | ❌ | ❌ |
| `PUT /api/matches/:match_id` | PUT | ✅ | ❌ | ❌ |
| `PUT /api/matches/:match_id/finish` | PUT | ✅ | ❌ | ❌ |
| `DELETE /api/matches/:match_id` | DELETE | ✅ | ❌ | ❌ |
| `GET /api/matches/:match_id/goals` | GET | ✅ | ✅ | ✅ |
| `POST /api/matches/:match_id/goals` | POST | ✅ | ❌ | ❌ |
| `DELETE /api/goals/:id` | DELETE | ✅ | ❌ | ❌ |
| `GET /api/matches/:match_id/report` | GET | ✅ | ✅ | ✅ |

**Keterangan:**
- ✅ (own) - MANAGER hanya dapat mengubah tim yang memiliki `manager_user_id` sesuai dengan ID user-nya.
- ✅ (own team) - MANAGER hanya dapat mengelola pemain dalam tim yang ia manage.
- ❌ - Akses ditolak dengan `403 Forbidden`.
---

## Asumsi

Berikut adalah asumsi yang digunakan dalam pengembangan aplikasi:

### Domain

1. **Satu pemain hanya bermain untuk satu tim** - relasi `players.team_id` many-to-one ke teams.
2. **Nomor punggung unik per tim** - kombinasi `(team_id, jersey_number)` memiliki unique constraint.
3. **Pertandingan hanya mempertemukan dua tim yang berbeda** - home_team_id dan away_team_id tidak boleh sama.
4. **Tim yang bertanding harus sudah terdaftar** - validasi keberadaan kedua tim saat membuat pertandingan.
5. **Hanya ADMIN yang bisa membuat/menghapus pertandingan dan gol** - MANAGER dan USER tidak memiliki akses.
6. **Gol dapat dicatat pada pertandingan yang sudah selesai** - tidak ada larangan menambah gol ke pertandingan FINISHED.
7. **Pertandingan hanya bisa di-finish sekali** - `ErrAlreadyFinished` jika status sudah FINISHED.
8. **Skor tidak disimpan secara eksplisit** - skor dihitung dari jumlah gol per tim.
9. **Laporan mencakup akumulasi kemenangan** - `countCumulativeWins` menghitung kemenangan tim dalam pertandingan FINISHED sebelumnya termasuk pertandingan saat ini.
10. **Logo tim disimpan dalam bentuk url** - memudahkan proses penyimpanan dan integrasi dengan file storage eksternal (e.g. Amazon S3)

### Teknis

1. **Soft delete** - semua model menggunakan `gorm.DeletedAt`. Data tidak benar-benar dihapus dari database.
2. **Migrasi otomatis** - struktur tabel dibuat/diperbarui otomatis saat aplikasi startup. Tidak ada file migrasi manual.
3. **Case-insensitive email** - email dinormalisasi ke lowercase sebelum disimpan dan dicari.
4. **Seeding** - data awal hanya dimasukkan jika tabel masih kosong.
5. **Truncate + RESTART IDENTITY** - pada seeding development, semua tabel di-truncate dengan restart identity untuk memastikan ID deterministik.
6. **Environment variables** - konfigurasi sepenuhnya melalui env var (file `.env` untuk lokal, system env untuk production/Docker).
7. **Gin debug mode** - digunakan di development. Untuk production, set `GIN_MODE=release`.
8. **Hanya komponen** `/api/matches` **yang memerlukan pagination** - jumlah pertandingan merupakan data yang berpotensi memiliki kelajuan peningkatan paling besar. Oleh karena itu, by-default, response dibatasi 500 rows (angka arbitrer)

---

## Endpoint API

Semua endpoint (kecuali `/health` dan `/auth/login`) membutuhkan header:
```
Authorization: Bearer <token>
```

### Autentikasi

| Method | Path | Deskripsi |
|---|---|---|
| `POST` | `/auth/login` | Login dan dapatkan JWT token |

### Teams

| Method | Path | Role | Deskripsi |
|---|---|---|---|
| `GET` | `/api/teams` | Semua | Daftar semua tim |
| `GET` | `/api/teams/:team_id` | Semua | Detail tim |
| `POST` | `/api/teams` | ADMIN | Buat tim baru |
| `PUT` | `/api/teams/:team_id` | ADMIN / MANAGER (own) | Update tim |
| `DELETE` | `/api/teams/:team_id` | ADMIN | Hapus tim (soft delete) |

### Players

| Method | Path | Role | Deskripsi |
|---|---|---|---|
| `GET` | `/api/teams/:team_id/players` | Semua | Daftar pemain dalam tim |
| `POST` | `/api/teams/:team_id/players` | ADMIN / MANAGER (own) | Tambah pemain |
| `GET` | `/api/players/:id` | Semua | Detail pemain |
| `PUT` | `/api/players/:id` | ADMIN / MANAGER (own) | Update pemain |
| `DELETE` | `/api/players/:id` | ADMIN / MANAGER (own) | Hapus pemain (soft delete) |

### Matches

| Method | Path | Role | Deskripsi |
|---|---|---|---|
| `GET` | `/api/matches` | Semua | Daftar semua pertandingan (Default pagination limit: 500 pertandingan) |
| `GET` | `/api/matches/:match_id` | Semua | Detail pertandingan |
| `POST` | `/api/matches` | ADMIN | Buat pertandingan baru |
| `PUT` | `/api/matches/:match_id` | ADMIN | Update pertandingan |
| `PUT` | `/api/matches/:match_id/finish` | ADMIN | Selesaikan pertandingan |
| `DELETE` | `/api/matches/:match_id` | ADMIN | Hapus pertandingan (soft delete) |

### Goals

| Method | Path | Role | Deskripsi |
|---|---|---|---|
| `GET` | `/api/matches/:match_id/goals` | Semua | Daftar gol dalam pertandingan |
| `POST` | `/api/matches/:match_id/goals` | ADMIN | Catat gol baru |
| `DELETE` | `/api/goals/:id` | ADMIN | Hapus gol (soft delete) |

### Reports

| Method | Path | Role | Deskripsi |
|---|---|---|---|
| `GET` | `/api/matches/:match_id/report` | Semua | Laporan pertandingan |

### Health Check

| Method | Path | Role | Deskripsi |
|---|---|---|---|
| `GET` | `/health` | Publik | Cek status server |


---

## Development

### Seed Data

Saat pertama kali dijalankan dengan `APP_ENV=development`, aplikasi melakukan seeding otomatis:

**User seed** (jika tabel `users` kosong):

| Role | Email | Password |
|---|---|---|
| ADMIN | admin@xyz-football.local | ChangeMe123! |
| MANAGER | manager@xyz-football.local | ChangeMe123! |
| USER | user@xyz-football.local | ChangeMe123! |

**Domain seed** (jika `SEED_DOMAIN=true`, di-`TRUNCATE` + restart identity tiap startup development):

| Data | Detail |
|---|---|
| Tim | Jakarta Mavericks (ID=1, manager=User#2), Bali Dragon (ID=2), Bandung Giants (ID=3) |
| Pemain | 4 pemain per tim (total 12), dengan posisi STRIKER, MIDFIELDER, DEFENDER, GOALKEEPER, FLEX |
| Pertandingan | Match#1 (FINISHED, Mavericks vs Dragon), Match#2-3 (SCHEDULED) |
| Gol | 3 gol di Match#1 |

### Testing

```bash
# Auth tests
go test -v ./internal/auth/...

# Team tests
go test -v ./internal/team/...

# Player tests
go test -v ./internal/player/...

# Match tests
go test -v ./internal/match/...

# Goal tests
go test -v ./internal/goal/...

# Reporft tests
go test -v ./internal/report/...

# Test spesifik
go test ./internal/... -run TestPlayerService_JerseyUniqueness -v
```

### Environment Variables

| Variable | Default | Deskripsi |
|---|---|---|
| `PORT` | `8080` | Port HTTP server |
| `APP_ENV` | `development` | Environment (`development`/`production`) |
| `DB_HOST` | `localhost` | Host PostgreSQL |
| `DB_PORT` | `5432` | Port PostgreSQL |
| `DB_USER` | `postgres` | User PostgreSQL |
| `DB_PASSWORD` | `postgres` | Password PostgreSQL |
| `DB_NAME` | `xyz_football` | Nama database |
| `DB_SSLMODE` | `disable` | SSL mode PostgreSQL |
| `DB_MAX_OPEN_CONNS` | `25` | Maksimum koneksi terbuka |
| `DB_MAX_IDLE_CONNS` | `5` | Maksimum idle koneksi |
| `DB_CONN_MAX_LIFETIME_MIN` | `5` | Maksimum umur koneksi (menit) |
| `JWT_SECRET` | `dev-secret-change-me` | Secret key untuk JWT |
| `JWT_EXPIRY_MINUTES` | `60` | Masa berlaku token (menit) |
| `SEED_USERS` | `true` (development) | Aktifkan seeding user |
| `SEED_DOMAIN` | `true` (development) | Aktifkan seeding data domain |

> Catatan: Di luar environment `development`, `JWT_SECRET` harus diubah dari nilai default (aplikasi akan fatal error).

---
