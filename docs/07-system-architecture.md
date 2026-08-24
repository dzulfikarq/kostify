# 7. System Architecture

## 7.1 Diagram

```
                        ┌────────────────────────────────────────────────┐
                        │                    Browser                     │
                        │   React SPA (Tailwind, Axios, React Query)     │
                        └───────┬───────────────────────────▲────────────┘
                                │ HTTPS (cookie HttpOnly)   │ JSON + static
                        ┌───────▼───────────────────────────┴────────────┐
                        │                  Nginx (reverse proxy)         │
                        │  /        → frontend build (static)            │
                        │  /api/v1 → backend:8080                        │
                        └───────┬────────────────────────────────────────┘
                                │
                ┌───────────────▼───────────────────────────────┐
                │            Backend API (Go + Gin)             │
                │                                               │
                │  interfaces/   handler · middleware · dto     │
                │      │ auth · rbac · csrf · rate_limit        │
                │      ▼                                        │
                │  application/  use case (business logic)      │
                │      │                                        │
                │      ▼                                        │
                │  domain/       entity + repository interface  │
                │      │                                        │
                │      ▼                                        │
                │  infrastructure/  implementasi adapter        │
                └───┬──────────┬──────────────┬────────────┬────┘
                    │          │              │            │
             ┌──────▼───┐ ┌────▼────┐  ┌──────▼─────┐ ┌───▼────────────┐
             │PostgreSQL│ │  Redis  │  │   MinIO    │ │ SMTP (Mailhog) │
             │ data utama│ │ cache + │  │ foto kost  │ │ email via      │
             │           │ │ queue   │  │ (S3 API)   │ │ worker         │
             └──────────┘ └────┬────┘  └────────────┘ └───▲────────────┘
                               │ enqueue job               │ consume
                        ┌──────▼───────────────────────────┴────┐
                        │        Worker (Go, proses sama)        │
                        │  • asynq: kirim email notifikasi       │
                        │  • ticker 60s: auto-expire booking     │
                        └────────────────────────────────────────┘
```

## 7.2 Penjelasan Layanan

| Layanan | Peran | Catatan |
| ------- | ----- | ------- |
| **Frontend** | SPA React; render semua UI; menyimpan TIDAK ADA token di storage | Static file dilayani Nginx |
| **Backend API** | REST `/api/v1`; validasi, otorisasi, business rules; satu-satunya yang bicara ke DB | Cookie access+refresh diterbitkan di sini |
| **Worker** | Proses Go terpisah (binary sama, mode `--worker`): konsumsi queue Redis untuk email; cron auto-expiry booking | Memisahkan kerja lambat dari request cycle |
| **PostgreSQL** | Source of truth seluruh entitas | Satu DB, skema sesuai docs/03 |
| **Redis** | (a) Cache listing publik & dashboard; (b) rate limit login; (c) queue asynq | Persistence `appendonly yes` agar job tidak hilang saat restart |
| **MinIO** | Object storage foto kost (S3-compatible) | Bucket private; akses foto via presigned URL berumur pendek atau public-read bucket tergantung kebutuhan demo — default: URL publik read-only per objek |
| **Nginx** | Reverse proxy + serve static + TLS termination (prod) | Dev juga bisa pakai Vite proxy |
| **Mailhog** | SMTP catch-all dev | Ganti SMTP provider di prod via env |

## 7.3 Arsitektur Backend (Clean/Layered)

Alur dependensi satu arah:

```
interfaces/handler → application/usecase → domain (interface) ← infrastructure (implement)
```

- **domain/**: struct entity (`User`, `Property`, `Room`, `Booking`, ...), enum status,
  interface repository (`UserRepository`, `BookingRepository`, ...), dan error domain
  (`ErrNotFound`, `ErrConflict`, `ErrForbidden`). Tidak import GORM/Gin.
- **application/**: use case per fitur (`auth_usecase.go`, `property_usecase.go`,
  `booking_usecase.go`, ...). Semua aturan bisnis ada di sini: state machine booking,
  syarat submit verifikasi, snapshot harga, pembuatan notifikasi + enqueue.
  Transaksi multi-tabel dibungkus `TransactionManager` (interface di domain,
  implement GORM di infrastructure).
- **infrastructure/**: `postgres/` (repository GORM), `redis/` (cache + rate limiter),
  `minio/` (uploader), `queue/` (asynq producer & handler), `smtp/` (mail sender).
- **interfaces/**: `http/handler` (parse req → DTO → panggil usecase → format response),
  `http/middleware` (`Auth`, `RequireRole`, `CSRF`, `RateLimit`, `RequestID`, `Logger`,
  `Recovery`), `http/dto` (request/response struct + binding tag), `http/routes.go`.

Kenapa layered clean: mudah di-test (usecase unit-test dengan mock repository),
mudah menjawab "kenapa" saat interview, dan boundary jelas tanpa over-engineering.

## 7.4 Alur Autentikasi (cookie-based)

```
Login:
 client → POST /auth/login → verifikasi bcrypt → generate:
   access_token  = JWT {sub, role, jti} exp 15m  → cookie HttpOnly Secure SameSite=Lax Path=/ Max-Age=900
   refresh_token = opaque random 256-bit, disimpan HASHED di Redis (key rt:<hash>)
                   exp 7d → cookie HttpOnly ... Max-Age=604800
   csrf_token    = random 128-bit → cookie SameSite=Lax TANPA HttpOnly (harus bisa dibaca JS)
 Response body memuat user + csrf_token.

Request mutasi:
 middleware CSRF: bandingkan header X-CSRF-Token vs cookie csrf_token → mismatch 403.

Request terlindungi:
 middleware Auth: verifikasi JWT access token dari cookie → set user_id+role ke context.

Access expired:
 interceptor axios → POST /auth/refresh (kirim cookie refresh):
   cek hash token ada di Redis → hapus key lama, terbit pasangan baru (ROTASI).
   Jika key sudah tidak ada = token reuse → cabut seluruh family (delete family key) → 401.

Logout:
 hapus Redis key refresh + clear cookie (Max-Age=0).
```

**Kenapa cookie bukan localStorage:** token di localStorage dapat dibaca script XSS;
HttpOnly membuat JS browser tidak bisa membaca token sama sekali.
**Kenapa SameSite=Lax:** mencegah cookie dikirim pada cross-site POST (mitigasi CSRF
lapis pertama); double submit token menjadi lapis kedua yang eksplisit.
**Kenapa double-submit untuk CSRF:** stateless (tidak butuh tabel session server-side),
sederhana, dan standar OWASP; attacker di domain lain tidak bisa MEMBACA cookie
untuk menyetel header yang cocok.

## 7.5 Alur Booking + Auto-Expiry + Queue

```
Tenant klik Booking
  → POST /bookings (CSRF + JWT)
  → BookingUsecase.Create (satu transaksi):
      lock room FOR UPDATE → cek available → insert booking(pending, expires_at=+72h)
      → update room pending → insert history → insert notification(owner)
      → asynq.Enqueue(EmailJob{type:"booking_created", to:owner})
  → 201

Worker ticker tiap 60s:
  SELECT ... WHERE status='pending' AND expires_at<now() FOR UPDATE SKIP LOCKED LIMIT 100
  → per baris: expire booking → room available → history → notifications → enqueue emails

EmailJob handler:
  render template → SMTP send → ack.
  Retry otomatis 3x backoff (asynq default) → gagal terus masuk archived queue (inspeksi manual).
```

## 7.6 Caching Strategy (bonus)

| Key                          | Isi                              | TTL    | Invalidasi                      |
| ---------------------------- | -------------------------------- | ------ | -------------------------------- |
| `props:list:{hash(query)}`   | hasil listing publik             | 60s    | TTL saja (data listing toleran stale singkat) |
| `prop:detail:{id}`           | detail property published        | 300s   | DEL saat approve/reject/update   |
| `dashboard:admin`            | statistik admin                  | 30s    | TTL                              |

Rate limit login: counter INCR `rl:login:{ip}` EXPIRE 60 → >5 = 429.

## 7.7 Keamanan Ringkas

- Bcrypt cost 12 untuk password.
- Security headers middleware: `X-Content-Type-Options`, `X-Frame-Options: DENY`,
  `Strict-Transport-Security` (prod), CSP ketat untuk frontend.
- Rate limit global ringan (100 req/menit/IP) + khusus login.
- Upload: sniff MIME magic bytes, batas ukuran, nama object UUID, ekstensi whitelisted,
  disajikan dari origin MinIO terpisah (bukan domain app) → isolasi XSS via file.
- Error internal di-log dengan request ID; response ke client generik.

## 7.8 docker-compose Services

```yaml
services:
  postgres:   # postgres:16-alpine, volume pgdata, healthcheck pg_isready
  redis:      # redis:7-alpine, appendonly yes, healthcheck
  minio:      # minio/minio, console + API port, createbuckets init
  mailhog:    # mailhog/mailhog (SMTP 1025, UI 8025)
  api:        # build backend/, depends_on healthy, migrate-up lalu serve
  worker:     # image sama dengan api, command --mode=worker
  frontend:   # build frontend/, Nginx serving dist + proxy /api
```

Satu perintah `docker compose up --build` menyalakan semuanya; migration dijalankan
otomatis oleh container `api` sebelum listen (idempotent).
