# Kostify

Sistem manajemen kost fullstack — Go (Gin/GORM) · PostgreSQL · Redis · MinIO · React + TypeScript + Tailwind.

Take-home test **Fullstack Engineer — PT. Mumtaz Teknologi Indonesia**.

## Quick Start

```bash
docker compose up --build -d
```

Satu perintah: PostgreSQL, Redis, MinIO, Mailhog, API (auto-migrate + seed otomatis saat pertama kali), dan frontend production build.

| Layanan        | URL                            |
| -------------- | ------------------------------ |
| Frontend (SPA) | http://localhost:3000          |
| API            | http://localhost:8080/api/v1   |
| Health         | /healthz · /readyz             |
| MinIO Console  | http://localhost:9001          |
| Mailhog        | http://localhost:8025          |

Frontend dev mode (hot reload): `cd frontend && npm install && npm run dev` → http://localhost:5173 (proxy `/api` ke :8080).

### Akun Demo

| Role        | Email                | Password    |
| ----------- | -------------------- | ----------- |
| Super Admin | admin@kostify.test   | Admin123!   |
| Owner       | owner@kostify.test   | Owner123!   |
| Tenant      | tenant@kostify.test  | Tenant123!  |

## Fitur

- **Auth**: cookie HttpOnly (access JWT 15 menit + refresh opaque 7 hari dengan rotasi atomik via Lua & reuse detection), CSRF double-submit untuk mutasi logout, rate limit login.
- **RBAC**: tenant / owner / super admin (middleware role-based; ownership check di service layer).
- **Properti**: CRUD kost & kamar oleh owner, upload foto ke MinIO (MIME sniffing, maks 5MB/10 foto), state machine verifikasi `draft → pending_verification → published/rejected`.
- **Listing publik**: filter kota/search/harga/fasilitas/rating, sort, pagination.
- **Booking**: transaksi DB dengan row-lock (`FOR UPDATE`), snapshot harga, expiry 72 jam via worker `SKIP LOCKED`, state machine lengkap pending→survey→booked→active→completed + reject/cancel/expire.
- **Notifikasi**: in-app (bell unread count) + email via Mailhog.
- **Review**: hanya booking COMPLETED, satu review per booking, recompute agregat rating properti.
- **Wishlist**: idempotent add/remove.
- **Dashboard**: agregasi per role (okupansi & revenue owner, ringkasan tenant, statistik admin).

## Alur Demo End-to-End

1. Login owner → Kost Saya → Tambah Kost → unggah foto → tambah kamar → Ajukan Verifikasi.
2. Login admin → Verifikasi → Review detail → Setujui.
3. Login tenant → Cari Kost → detail → Ajukan Booking.
4. Owner: Booking Masuk → Setujui → Konfirmasi (tanggal mulai).
5. Tenant: Booking Saya → Check-in → Check-out → Beri Ulasan.
6. Notifikasi muncul di kedua pihak; email terlihat di Mailhog.

## Arsitektur Singkat

```
Browser ──► Nginx (web) ──► Go API ──► PostgreSQL (sumber kebenaran)
                          │        ├─ Redis     (refresh token store)
                          │        ├─ MinIO     (foto)
                          │        └─ Mailhog   (SMTP dev)
                          └─ Worker ticker 60s: auto-expiry booking
```

Layered backend: `domain` (entitas + kontrak) → `application` (use case, aturan bisnis) → `infrastructure` (postgres/minio/redis/mailer) → `interfaces/http` (handler/dto/middleware).

## Keputusan Teknis & Trade-off

| Keputusan | Alasan |
| --------- | ------ |
| Cookie auth (bukan localStorage) | XSS tidak bisa mencuri token; refresh rotasi + reuse detection memitigasi replay. |
| Partial unique index di DB (`uq_one_pending_per_room`, `uq_room_single_active`) | Race condition booking dilindungi di level DB, bukan hanya aplikasi. |
| Snapshot harga saat booking | Harga kamar boleh berubah; booking harus stabil. |
| Worker ticker SQL (bukan asynq) | Fallback sesuai rencana risiko docs/09 — lebih sedikit infrastruktur; upgrade path jelas. |
| Email inline goroutine (bukan queue) | Trade-off yang sama: tanpa durability; cukup untuk skala demo, antrean asynq bila perlu retry. |
| Rating denormalisasi di tabel properties | Listing butuh sort/filter by rating tanpa join agregasi. |
| Param route `:id` konsisten (deviasi kontrak `:propertyId`) | Gin menuntut nama param seragam per posisi pohon routing. |

## Testing

```bash
# backend (unit: jwt, password hashing, domain rules)
cd backend && go test ./...

# frontend (typecheck + build)
cd frontend && npm run build
```

CI: `.github/workflows/ci.yml` menjalankan backend test + vet dan frontend typecheck/build pada setiap push.

## Dokumentasi

| Dokumen | Isi |
| ------- | --- |
| [01 Product Foundation](docs/01-product-foundation.md) | Problem, solusi, target user, scope |
| [02 Business Requirements](docs/02-business-requirements.md) | RBAC matrix, workflow properti & booking, aturan bisnis |
| [03 Database & ERD](docs/03-database-erd.md) | ERD + SQL migration up/down lengkap |
| [04 API Contract](docs/04-api-contract.md) | Konvensi, auth, properti, kamar, upload foto |
| [05 API Contract Transaksi](docs/05-api-contract-transactions.md) | Booking, review, wishlist, notifikasi, dashboard |
| [06 Frontend IA](docs/06-frontend-ia.md) | Route map, layout, komponen, interceptor axios |
| [07 System Architecture](docs/07-system-architecture.md) | Diagram layanan, alur auth/booking/queue, caching, keamanan |
| [08 Repository Structure](docs/08-repository-structure.md) | Struktur folder backend & frontend |
| [09 Milestones](docs/09-milestones.md) | Jadwal 5 minggu + exit criteria per fase |
