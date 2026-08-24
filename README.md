# Kostify

Sistem manajemen kost fullstack — Go (Gin/GORM) · PostgreSQL · Redis · MinIO · React + Tailwind.

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

## Quick Start (setelah implementasi)

```bash
docker compose up --build
```

- Frontend: http://localhost:5173
- API: http://localhost:8080/api/v1
- Swagger: http://localhost:8080/swagger/index.html
- MinIO Console: http://localhost:9001
- Mailhog: http://localhost:8025

Akun demo (via `make seed`): admin@kostify.test / owner@kostify.test / tenant@kostify.test
