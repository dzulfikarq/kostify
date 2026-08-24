# 9. Implementation Milestones (5 Minggu)

Prinsip: vertikal slice — tiap minggu berakhir dengan fitur yang jalan end-to-end
(DB → API → UI), bukan lapisan terpisah.

## Minggu 1 — Fondasi: Infra, DB, Auth, RBAC

| Hari | Output                                                                                     |
| ---- | ------------------------------------------------------------------------------------------- |
| 1    | Skeleton repo monorepo, docker-compose (postgres, redis, minio, mailhog), .env.example      |
| 2    | Migration 000001 up/down, Makefile (`migrate-up/down`, `seed`), koneksi GORM                 |
| 3    | Domain entities + interfaces, error domain, response envelope + centralized error handler    |
| 4    | Auth: register, login (cookie HttpOnly), me, logout; middleware Auth                         |
| 5    | Refresh token rotasi di Redis, CSRF double-submit middleware, RBAC middleware, rate limit login |

**Exit criteria:** `docker compose up` hidup; register→login→me→refresh→logout via curl/Postman;
CSRF ditolak tanpa header; endpoint owner-only menolak tenant (403).

## Minggu 2 — Properti, Kamar, Foto, Verifikasi

| Hari | Output                                                                                       |
| ---- | ---------------------------------------------------------------------------------------------- |
| 1-2  | CRUD property (owner) + listing publik PUBLISHED dengan search/filter/sort/pagination          |
| 3    | Upload foto MinIO (validasi MIME/size/count) + delete foto; detail property (photos+rooms)     |
| 4    | CRUD room + toggle maintenance + validasi unik nomor kamar                                     |
| 5    | Submit verifikasi (pre-condition ≥1 kamar ≥1 foto) + approve/reject admin (+reason, log)       |
|      | Frontend: OwnerProperties, PropertyForm, RoomManager, VerificationQueue                        |

**Exit criteria:** alur lengkap owner buat kost → upload foto → tambah kamar → submit →
admin approve/reject → tampil/hilang dari listing publik.

## Minggu 3 — Booking System

| Hari | Output                                                                                          |
| ---- | ------------------------------------------------------------------------------------------------ |
| 1    | POST /bookings (transaksi: lock kamar, snapshot harga, expires_at=+72h, history, notifikasi)     |
| 2    | Worker auto-expiry (ticker 60s, SKIP LOCKED) + integrasi asynq queue untuk email                  |
| 3    | Approve/reject/confirm owner + cancel tenant; semua update status kamar konsisten                |
| 4    | Check-in/check-out; GET bookings/me & /owner                                                     |
| 5    | Frontend TenantBookings + OwnerBookings (aksi kontekstual, countdown expiry, confirm dialog)     |

**Exit criteria:** booking dibuat → expire otomatis saat 72h lewat (diuji dengan TTL pendek);
seluruh state machine berjalan; transisi ilegal ditolak 409; email masuk Mailhog via queue.

## Minggu 4 — Dashboard, Review, Wishlist, Notifikasi

| Hari | Output                                                                              |
| ---- | ------------------------------------------------------------------------------------- |
| 1    | Review: create (validasi COMPLETED, unique) + recompute rating; list review publik     |
| 2    | Wishlist: add/remove/list + tombol hati di UI                                         |
| 3    | Notification in-app (list, read, read-all, bell unread count) + email template rapi   |
| 4    | Dashboard agregat per role (owner okupansi & revenue, tenant ringkas, admin statistik)|
| 5    | Cache Redis listing/detail/dashboard; halaman frontend pendukung                      |

**Exit criteria:** seluruh fitur utama terpakai mulas: cari → booking → survey → booked →
checkin → checkout → review; rating properti ter-update; notifikasi muncul di kedua pihak.

## Minggu 5 — Polish, Testing, Dokumentasi, Deploy

| Hari | Output                                                                                        |
| ---- | ----------------------------------------------------------------------------------------------- |
| 1    | UI/UX audit: loading skeleton semua fetch, empty state, error state, disabled state, responsive 3 breakpoint |
| 2    | Unit test usecase (auth, booking state machine, verifikasi pre-condition, review rule)           |
| 3    | Integration test handler (httptest + testcontainers PG/Redis); component test frontend minimal   |
| 4    | Swagger docs final, README lengkap (arsitektur, keputusan, trade-off), CI workflow               |
| 5    | E2E critical flow (login → create property → submit → approve → book → checkin → checkout), bugfix|

**Exit criteria:** repo siap dikumpulkan: satu perintah jalan penuh, test hijau di CI,
dokumentasi menjawab daftar pertanyaan interview assignment.

## Buffer & Risiko

| Risiko | Mitigasi |
| ------ | -------- |
| Asynq/email delay lambat dipelajari | Fallback: worker ticker polling tabel jobs sederhana; asynq menyusul |
| Race condition booking | Partial unique index sudah melindungi; uji concurrent dengan script paralel |
| Waktu UI molor | Prioritas halaman inti dulu; halaman admin boleh fungsional-minimalis |
| MinIO presigned URL rumit | Mulai dari bucket public-read sederhana; upgrade presigned kalau sempat |
