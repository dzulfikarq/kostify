# 5. API Contract (Bagian 2 — Booking, Review, Wishlist, Notifikasi)

## 5.1 Bookings

### State machine (referensi)

```
PENDING ──approve──► SURVEY ──confirm──► BOOKED ──checkin──► ACTIVE ──checkout──► COMPLETED
   │ │                  │
   │ └──cancel──────────┴──cancel──► CANCELLED
   └──expire (72 jam)──► EXPIRED      SURVEY/PENDING ──reject(owner)──► REJECTED
```

### `POST /api/v1/bookings` — Role tenant, CSRF

Request:

```json
{
  "room_id": "c9bf9e57-1685-4c89-bafb-ff5af830be8a",
  "lease_duration_months": 12,
  "note": "Apakah bisa survey akhir pekan ini?"
}
```

Validasi & aturan bisnis (dieksekusi berurutan di service layer):

1. `room_id` UUID valid; kamar ada → `404`;
2. Kamar status `AVAILABLE` → `409` bila bukan;
3. Kost induk berstatus `PUBLISHED` → `409`;
4. Tenant bukan pemilik kost tersebut → `403`;
5. Belum ada booking `PENDING` milik user ini pada kamar tsb (`uq_one_pending_per_room`);
6. `lease_duration_months` integer 1–36.

Semua dalam **satu transaksi DB**:

- INSERT booking: `status=pending`, `price_per_month = rooms.price_per_month` (snapshot),
  `expires_at = now() + interval '72 hours'`;
- UPDATE kamar: `available → pending`;
- INSERT `booking_history(pending)`;
- INSERT notifikasi untuk owner;
- ENQUEUE job email ke Redis (asynq).

Response `201 Created`:

```json
{
  "success": true,
  "data": {
    "id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
    "room_id": "c9bf9e57-1685-4c89-bafb-ff5af830be8a",
    "property": { "id": "f47ac10b...", "name": "Kost Putra Mekar Jaya" },
    "tenant_id": "9b1deb4d...",
    "owner_id": "6f9619ff...",
    "status": "pending",
    "price_per_month": 750000,
    "lease_duration_months": 12,
    "note": "Apakah bisa survey akhir pekan ini?",
    "expires_at": "2026-08-27T10:00:00Z",
    "created_at": "2026-08-24T10:00:00Z"
  },
  "message": "Booking berhasil dibuat. Menunggu konfirmasi pemilik maksimal 3 hari."
}
```

Kamar tidak available → `409`:

```json
{ "success": false, "error": { "code": "CONFLICT", "message": "Kamar sedang tidak tersedia" } }
```

### `GET /api/v1/bookings/me` — Role tenant

Daftar booking milik tenant. Query: `page`, `limit`, `status`, `sort=created_at|expires_at`, `order`.

Response `200`: array booking + info property/room + meta pagination.

### `GET /api/v1/bookings/owner` — Role owner

Daftar booking masuk ke kost milik owner. Query sama + filter `property_id`.

### Detail transisi (endpoint owner)

| Endpoint                              | Dari              | Ke        | Ekstra                                        |
| ------------------------------------- | ----------------- | --------- | ---------------------------------------------- |
| `PUT /api/v1/bookings/:id/approve`    | `pending`         | `survey`  | body opsional `{ "survey_at": "...ISO..." }`  |
| `PUT /api/v1/bookings/:id/reject`     | `pending/survey`  | `rejected`| body wajib `{ "reason": "..." }` (min 10 char)|
| `PUT /api/v1/bookings/:id/confirm`    | `survey`          | `booked`  | body `{ "start_date": "2026-09-01" }` wajib   |

Semua hanya oleh owner pemilik kost (cek `bookings.owner_id == user.id`) — role check di
middleware, ownership check di service. Transisi ilegal → `409`.

Contoh request approve:

```json
{ "survey_at": "2026-08-26T14:00:00Z" }
```

Contoh response `200`:

```json
{
  "success": true,
  "data": { "id": "7c9e6679...", "status": "survey" },
  "message": "Booking disetujui, menunggu jadwal survey"
}
```

Contoh reject request/response:

```json
// request
{ "reason": "Kamar sudah disewakan sampai Desember" }

// response 200
{ "success": true, "data": { "id": "7c9e6679...", "status": "rejected" }, "message": "Booking ditolak" }
```

Setiap aksi: update booking + room status + insert history + notifikasi tenant + enqueue email.

### Check-in / Check-out (endpoint tenant)

| Endpoint                            | Dari     | Ke         | Efek kamar            |
| ------------------------------------ | -------- | ---------- | ---------------------- |
| `PUT /api/v1/bookings/:id/checkin`   | `booked` | `active`   | `booked → active`; set `checked_in_at` |
| `PUT /api/v1/bookings/:id/checkout`  | `active` | `completed`| `active → available`; set `checked_out_at` |

Hanya oleh tenant pemilik booking. Checkout → kamar otomatis available lagi.

Response checkin `200`:

```json
{ "success": true, "data": { "id": "7c9e6679...", "status": "active", "checked_in_at": "2026-09-01T02:15:00Z" }, "message": "Check-in berhasil. Selamat menempati!" }
```

Transisi ilegal → `409 CONFLICT`.

### Cancel (tenant atau owner sesuai konteks)

`PUT /api/v1/bookings/:id/cancel` — hanya dari `pending|survey`, hanya tenant pemilik
booking (owner memakai `reject`). Body opsional `{ "reason": "..." }`.
Efek: booking → `cancelled`, kamar → `available`, notifikasi ke lawan pihak.

### Auto-expiry worker (bukan endpoint)

Worker tiap 60 detik menjalankan (idempotent, aman terhadap instance ganda via
`FOR UPDATE SKIP LOCKED`):

```sql
UPDATE bookings SET status='expired', updated_at=now()
WHERE id IN (
  SELECT id FROM bookings
  WHERE status='pending' AND expires_at < now()
  LIMIT 100
  FOR UPDATE SKIP LOCKED
)
RETURNING id, tenant_id, owner_id, room_id;
```

Untuk tiap baris: kamar `pending → available`, insert history, notifikasi tenant+owner,
enqueue email. Log `worker.expired_bookings` terstruktur.

---

## 5.2 Reviews

### `POST /api/v1/bookings/:bookingId/reviews` — Tenant pemilik booking, CSRF

Request:

```json
{ "rating": 5, "comment": "Kost bersih, owner ramah, wifi kencang." }
```

Aturan bisnis:

1. Booking milik tenant yang login → `403` selain itu;
2. Status booking harus `COMPLETED` → `409` bila belum;
3. Satu booking satu review (unique constraint `booking_id`) → `409` duplikat;
4. `rating` integer 1–5; `comment` max 1000 char opsional.

Dalam satu transaksi: insert review + recompute agregat properti
(`rating_avg = AVG(rating)`, `rating_count = COUNT(*)`) → update tabel properties.

Response `201 Created`:

```json
{
  "success": true,
  "data": {
    "id": "d1b6a7c8-9999-8888-7777-666655554444",
    "booking_id": "7c9e6679...",
    "property_id": "f47ac10b...",
    "rating": 5,
    "comment": "Kost bersih, owner ramah, wifi kencang.",
    "created_at": "2026-12-01T08:00:00Z"
  },
  "message": "Review berhasil dikirim"
}
```

Belum completed → `409`:

```json
{ "success": false, "error": { "code": "CONFLICT", "message": "Review hanya dapat dibuat setelah masa sewa selesai" } }
```

### `GET /api/v1/properties/:propertyId/reviews` — Publik

Query: `page`, `limit`, `sort=created_at|rating`, `order`.

Response `200`:

```json
{
  "success": true,
  "data": [
    {
      "id": "d1b6a7c8...",
      "rating": 5,
      "comment": "Kost bersih, owner ramah.",
      "tenant_name": "Budi Santoso",
      "created_at": "2026-12-01T08:00:00Z"
    }
  ],
  "meta": { "page": 1, "limit": 10, "total": 8, "total_pages": 1 }
}
```

---

## 5.3 Wishlist

### `POST /api/v1/wishlist` — Authenticated (tenant/owner), CSRF

Request:

```json
{ "property_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479" }
```

Validasi: property ada dan `PUBLISHED`. Duplikat tetap idempotent → `200`
(tidak error 409, supaya tombol hati aman dipencet berkali-kali).

Response `200`:

```json
{ "success": true, "data": { "property_id": "f47ac10b..." }, "message": "Ditambahkan ke wishlist" }
```

### `DELETE /api/v1/wishlist/:propertyId` — Authenticated, CSRF

Response `204 No Content`.

### `GET /api/v1/wishlist` — Authenticated

Query: `page`, `limit`. Join properties (nama, foto cover, harga mulai, kota).

Response `200`: array ringkas + meta pagination.

---

## 5.4 Notifications

### `GET /api/v1/notifications` — Authenticated

Query: `page`, `limit`, `is_read=true|false`, `type=booking|verification|review|system`.

Response `200`:

```json
{
  "success": true,
  "data": [
    {
      "id": "0f1e2d3c-aaaa-bbbb-cccc-ddddeeeeffff",
      "title": "Booking disetujui",
      "body": "Booking Anda untuk kamar A01 telah disetujui. Silakan atur jadwal survey.",
      "type": "booking",
      "ref_data": { "booking_id": "7c9e6679..." },
      "is_read": false,
      "created_at": "2026-08-25T04:20:00Z"
    }
  ],
  "meta": { "page": 1, "limit": 15, "total": 42, "total_pages": 3 }
}
```

### `PUT /api/v1/notifications/:id/read` — Authenticated, CSRF

Tandai satu notifikasi milik user sebagai dibaca. Response `200`.

### `PUT /api/v1/notifications/read-all` — Authenticated, CSRF

Tandai semua. Response `200` dengan `data: { "updated": 17 }`.

---

## 5.5 Dashboard (agregasi per role)

### `GET /api/v1/dashboard` — Authenticated

Respons bergantung role.

**Owner:**

```json
{
  "success": true,
  "data": {
    "role": "owner",
    "total_properties": 3,
    "published": 2,
    "pending_verification": 1,
    "total_rooms": 18,
    "occupied_rooms": 11,
    "occupancy_rate": 61.1,
    "bookings_pending": 4,
    "revenue_estimation_monthly": 8250000,
    "recent_bookings": [ /* 5 booking terbaru */ ]
  }
}
```

**Tenant:**

```json
{
  "success": true,
  "data": {
    "role": "tenant",
    "active_booking": { /* booking aktif atau null */ },
    "pending_bookings": 1,
    "wishlist_count": 6,
    "recommended_properties": [ /* 5 kost published rating tertinggi */ ]
  }
}
```

**Super Admin:**

```json
{
  "success": true,
  "data": {
    "role": "super_admin",
    "users_total": 152,
    "properties_total": 48,
    "waiting_verification": 5,
    "bookings_active": 33,
    "bookings_this_month": 21,
    "verification_queue": [ /* kost pending_verification terlama dulu */ ]
  }
}
```

---

## 5.6 Health & Observability

| Endpoint            | Keterangan                                                        |
| ------------------- | ------------------------------------------------------------------ |
| `GET /healthz`      | Liveness: proses hidup. Response `{"status":"ok"}`                 |
| `GET /readyz`       | Readiness: cek ping PostgreSQL, Redis, MinIO; `503` bila ada yang mati |

Middleware global: request ID (UUID per request, dikirim balik di header `X-Request-ID`
dan field `request_id` pada error response), structured logging JSON
(method, path, status, durasi, user_id bila ada), recovery panic → `500` generik.
