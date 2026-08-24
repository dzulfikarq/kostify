# 8. Repository Structure

```
kostify/
├── README.md                        # overview, setup, keputusan teknis (wajib assignment)
├── docker-compose.yml
├── .gitignore
├── .github/
│   └── workflows/ci.yml             # lint + test backend & frontend, build docker
├── docs/                            # dokumentasi blueprint ini
│   ├── 01-product-foundation.md
│   ├── 02-business-requirements.md
│   ├── 03-database-erd.md
│   ├── 04-api-contract.md
│   ├── 05-api-contract-transactions.md
│   ├── 06-frontend-ia.md
│   ├── 07-system-architecture.md
│   └── 08-repository-structure.md
│   └── 09-milestones.md
│
├── backend/
│   ├── cmd/
│   │   └── server/main.go           # entrypoint: load config, init dep, mode api|worker
│   ├── internal/
│   │   ├── domain/                  # ENTITIES + INTERFACE (tanpa import gin/gorm)
│   │   │   ├── user.go              # entity User, enum Role
│   │   │   ├── property.go          # entity Property, Photo, enum PropertyStatus
│   │   │   ├── room.go              # entity Room, enum RoomStatus
│   │   │   ├── booking.go           # entity Booking, History, enum BookingStatus + transisi valid
│   │   │   ├── review.go
│   │   │   ├── notification.go
│   │   │   ├── errors.go            # ErrNotFound, ErrConflict, ErrForbidden, ErrValidation...
│   │   │   └── repository.go        # interface semua repository + TransactionManager
│   │   ├── application/
│   │   │   ├── auth_usecase.go      # register/login/refresh(rotasi)/logout/me
│   │   │   ├── property_usecase.go  # CRUD, submit, approve/reject (+log)
│   │   │   ├── room_usecase.go
│   │   │   ├── booking_usecase.go   # create/approve/reject/confirm/checkin/out/cancel/expiry
│   │   │   ├── review_usecase.go    # + recompute rating properti
│   │   │   ├── wishlist_usecase.go
│   │   │   ├── notification_usecase.go
│   │   │   └── dashboard_usecase.go # agregasi per role
│   │   ├── infrastructure/
│   │   │   ├── postgres/
│   │   │   │   ├── connection.go    # gorm.Open + pool config
│   │   │   │   ├── tx_manager.go    # implementasi TransactionManager
│   │   │   │   ├── user_repo.go
│   │   │   │   ├── property_repo.go # termasuk query listing dinamis + scope filter
│   │   │   │   ├── room_repo.go
│   │   │   │   ├── booking_repo.go  # termasuk query expiry SKIP LOCKED
│   │   │   │   └── ...
│   │   │   ├── redis/client.go      # cache helper + rate limiter
│   │   │   ├── minio/uploader.go    # PutObject + validasi MIME sniff
│   │   │   ├── queue/producer.go    # asynq enqueue email job
│   │   │   ├── queue/email_handler.go
│   │   │   └── smtp/sender.go
│   │   └── interfaces/http/
│   │       ├── router.go            # /api/v1 grouping + middleware chain
│   │       ├── handler/
│   │       │   ├── auth_handler.go
│   │       │   ├── property_handler.go  # termasuk upload foto multipart
│   │       │   ├── room_handler.go
│   │       │   ├── booking_handler.go
│   │       │   ├── review_handler.go
│   │       │   ├── wishlist_handler.go
│   │       │   ├── notification_handler.go
│   │       │   └── dashboard_handler.go
│   │       ├── middleware/
│   │       │   ├── auth.go          # verifikasi JWT dari cookie → context
│   │       │   ├── rbac.go          # RequireRole("owner", "super_admin")
│   │       │   ├── csrf.go          # double submit check untuk method mutasi
│   │       │   ├── rate_limit.go    # login limiter + global limiter (Redis)
│   │       │   ├── request_id.go
│   │       │   ├── logger.go        # structured JSON log
│   │       │   ├── recovery.go
│   │       │   └── security_headers.go
│   │       ├── dto/                 # request struct + binding tags + response mapping
│   │       │   ├── auth_dto.go
│   │       │   ├── property_dto.go
│   │       │   ├── booking_dto.go
│   │       │   └── common.go        # envelope sukses/error + pagination meta
│   │       └── error_handler.go     # centralized: domain error → HTTP code konsisten
│   ├── migrations/                  # golang-migrate format
│   │   ├── 000001_init_schema.up.sql / .down.sql
│   │   └── 000002_....up.sql / .down.sql
│   ├── scripts/
│   │   └── seed/main.go             # akun demo + data contoh
│   ├── pkg/
│   │   ├── jwt/                     # sign/verify access token
│   │   └── validator/               # helper binding + pesan Indonesia
│   ├── .env.example
│   ├── Dockerfile                   # multi-stage build → distroless/alpine
│   ├── Makefile                     # migrate-up, migrate-down, seed, test, lint, run
│   └── go.mod
│
└── frontend/
    ├── src/
    │   ├── main.tsx
    │   ├── App.tsx                  # router definition + providers
    │   ├── components/
    │   │   ├── common/              # Button, Input, Select, Modal, ConfirmDialog,
    │   │   │                        # Drawer, Table, Pagination, Dropdown, Badge,
    │   │   │                        # Card, Toast, Spinner, Skeleton, EmptyState, StatusBadge
    │   │   ├── features/
    │   │   │   ├── auth/            # LoginForm, RegisterForm
    │   │   │   ├── properties/      # PropertyCard, PropertyForm, PhotoUploader,
    │   │   │   │                    # RoomTable, ReviewList, FilterSidebar
    │   │   │   ├── bookings/        # BookingCard, BookingActions, ExpiryCountdown
    │   │   │   └── notifications/   # NotificationItem, NotificationBell
    │   │   └── layout/              # PublicLayout, DashboardLayout, Header, Sidebar, Footer
    │   ├── pages/
    │   │   ├── public/              # HomePage, PropertiesPage, PropertyDetailPage,
    │   │   │                        # LoginPage, RegisterPage, NotFoundPage, ForbiddenPage
    │   │   ├── tenant/              # TenantDashboard, TenantBookings, WishlistPage
    │   │   ├── owner/               # OwnerDashboard, OwnerProperties, PropertyFormPage,
    │   │   │                        # RoomManagerPage, OwnerBookings
    │   │   ├── admin/               # AdminDashboard, VerificationQueuePage, UsersPage
    │   │   └── shared/              # NotificationsPage
    │   ├── hooks/
    │   │   ├── useAuth.ts           # context accessor + guards helpers
    │   │   ├── useProperties.ts     # React Query hooks (list/detail/mutations)
    │   │   ├── useBookings.ts
    │   │   └── useNotifications.ts
    │   ├── context/AuthContext.tsx
    │   ├── services/
    │   │   ├── api.ts               # axios instance + interceptors (lihat docs/06)
    │   │   ├── auth.service.ts
    │   │   ├── property.service.ts
    │   │   ├── room.service.ts
    │   │   ├── booking.service.ts
    │   │   ├── review.service.ts
    │   │   ├── wishlist.service.ts
    │   │   └── notification.service.ts
    │   ├── utils/
    │   │   ├── validators.ts        # zod schemas dipakai form
    │   │   ├── formatters.ts        # rupiah, tanggal, relatif waktu
    │   │   └── constants.ts         # enum status, mapping warna badge, fasilitas
    │   ├── types/                   # TS types mirror DTO backend
    │   ├── .env.example             # VITE_API_URL=http://localhost:8080
    │   ├── tailwind.config.js
    │   ├── Dockerfile               # build vite → nginx serve dist
    │   ├── nginx.conf               # try_files SPA + proxy /api → api:8080
    │   └── package.json
```

## File Kritis (urutan pengerjaan disarankan)

1. `docker-compose.yml` + `.env.example` × 2 — lingkungan siap sebelum kode.
2. `migrations/000001_*` + Makefile — skema reproducible.
3. `domain/*` — kontrak entitas & interface.
4. `auth` stack end-to-end: `jwt`, `auth_usecase`, middleware auth/csrf/rbac, `auth_handler`,
   frontend `AuthContext` + interceptor — fondasi semua fitur lain.
5. `property` + `room` stack + MinIO uploader.
6. `booking` stack + worker expiry + queue email.
7. `review`, `wishlist`, `notification`, `dashboard`.
8. Frontend pages mengikuti urutan fitur yang sama.
9. Testing & polish terus-menerus per milestone (bukan di akhir saja).
