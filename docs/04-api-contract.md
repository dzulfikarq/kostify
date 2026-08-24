# 4. API Contract (Bagian 1 — Konvensi, Auth, Properti, Kamar)

Base URL: `/api/v1`

## 4.1 Konvensi Umum

### Format response sukses

```json
{
  "success": true,
  "data": {},
  "message": "Success"
}
```

Untuk listing, ada tambahan `meta`:

```json
{
  "success": true,
  "data": [],
  "meta": { "page": 1, "limit": 12, "total": 87, "total_pages": 8 },
  "message": "Success"
}
```

### Format response error (konsisten untuk semua error)

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Input tidak valid",
    "details": [
      { "field": "email", "message": "harus berupa email yang valid" }
    ]
  },
  "request_id": "b3c9d1e0-4f2a-48c1-9e55-a1b2c3d4e5f6"
}
```

### Katalog error

| HTTP | Code               | Kapan                                                       |
| ---- | ------------------ | ------------------------------------------------------------ |
| 400  | `BAD_REQUEST`      | Body JSON rusak / format parameter salah                     |
| 401  | `UNAUTHORIZED`     | Token hilang / expired / invalid                             |
| 403  | `FORBIDDEN`        | Role tidak berizin, bukan pemilik resource, CSRF token salah |
| 404  | `NOT_FOUND`        | Resource tidak ada atau bukan milik requester                |
| 409  | `CONFLICT`         | Transisi status ilegal, duplikat, kamar tidak available      |
| 422  | `VALIDATION_ERROR` | Validasi input gagal                                         |
| 429  | `RATE_LIMITED`     | Melewati batas rate limit                                    |
| 500  | `INTERNAL_ERROR`   | Kesalahan server (detail internal TIDAK diekspos)            |

### Autentikasi & CSRF

- Cookie HttpOnly: `access_token` (15 menit), `refresh_token` (7 hari, dirotasi).
- Cookie non-HttpOnly `csrf_token`: dibaca JS, dikirim balik via header `X-CSRF-Token`
  pada setiap request mutasi (`POST/PUT/PATCH/DELETE`). Server mencocokkan cookie vs header.
- Semua cookie: `HttpOnly` (kecuali csrf), `Secure`, `SameSite=Lax`, `Path=/`, `Max-Age` eksplisit.

### Query standar listing

```
?page=1&limit=12&search=kost%20mewah&sort=created_at&order=desc
```

Validasi: `page ≥ 1`, `1 ≤ limit ≤ 100` (default 12), `order ∈ {asc, desc}`,
`sort` hanya boleh dari whitelist kolom (anti SQL injection).

---

## 4.2 Auth

### `POST /api/v1/auth/register` — Publik

Request:

```json
{
  "name": "Budi Santoso",
  "email": "budi@mail.com",
  "password": "Rahasia123",
  "role": "tenant"
}
```

Validasi: `name` required 2–100 char; `email` required format email; `password` required
min 8 char, mengandung huruf & angka; `role` enum `owner|tenant` (guest tidak bisa
mendaftar sebagai `super_admin`).

Response `201 Created`:

```json
{
  "success": true,
  "data": {
    "id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
    "name": "Budi Santoso",
    "email": "budi@mail.com",
    "role": "tenant"
  },
  "message": "Registrasi berhasil"
}
```

Error `422`:

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Input tidak valid",
    "details": [
      { "field": "password", "message": "minimal 8 karakter dan mengandung huruf serta angka" }
    ]
  }
}
```

Error `409` jika email sudah terdaftar:

```json
{
  "success": false,
  "error": { "code": "CONFLICT", "message": "Email sudah terdaftar" }
}
```

### `POST /api/v1/auth/login` — Publik, rate-limited 5/menit/IP

Request:

```json
{ "email": "budi@mail.com", "password": "Rahasia123" }
```

Response `200 OK` — set cookie `access_token`, `refresh_token`, `csrf_token`:

```json
{
  "success": true,
  "data": {
    "user": {
      "id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
      "name": "Budi Santoso",
      "email": "budi@mail.com",
      "role": "tenant"
    },
    "csrf_token": "d4f5e6a7-b8c9-0123-4567-89abcdefabcd"
  },
  "message": "Login berhasil"
}
```

> `csrf_token` juga dikembalikan di body agar client mudah membacanya;
> sumber kebenaran tetap cookie (double submit).

Error `401` (pesan generik, tidak membocorkan field mana yang salah):

```json
{ "success": false, "error": { "code": "UNAUTHORIZED", "message": "Email atau password salah" } }
```

Error `429`:

```json
{ "success": false, "error": { "code": "RATE_LIMITED", "message": "Terlalu banyak percobaan. Coba lagi dalam 60 detik." } }
```

### `POST /api/v1/auth/refresh` — Publik (cookie refresh_token wajib)

Rotasi: refresh token lama dinonaktifkan, token baru diterbitkan. Reuse dari token yang
sudah dipakai → seluruh sesi dicabut (deteksi pencurian token). Anti-race: operasi atomik di Redis/DB.

Response `200`: sama seperti login (cookie baru + csrf baru).

### `POST /api/v1/auth/logout` — Authenticated + CSRF

Hapus cookie, cabang refresh token di Redis. Response `200`:

```json
{ "success": true, "data": null, "message": "Logout berhasil" }
```

### `GET /api/v1/auth/me` — Authenticated

Dipakai frontend untuk restore auth state saat load halaman.

Response `200`:

```json
{
  "success": true,
  "data": {
    "id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
    "name": "Budi Santoso",
    "email": "budi@mail.com",
    "role": "tenant"
  }
}
```

Token expired → `401`; frontend interceptor memanggil `/auth/refresh` sekali lalu retry.

---

## 4.3 Properties

### `GET /api/v1/properties` — Publik (hanya PUBLISHED)

Query didukung:

| Param        | Contoh                          | Keterangan                              |
| ------------ | ------------------------------- | ---------------------------------------- |
| `page`       | `1`                             | default 1                                |
| `limit`      | `12`                            | default 12, max 100                      |
| `search`     | `kost putri bandung`            | cari di nama/deskripsi/alamat (ILIKE)    |
| `city`       | `Bandung`                       | filter persamaan kota                    |
| `min_price`  | `500000`                        | harga kamar termurah ≥                   |
| `max_price`  | `2000000`                       | harga termurah ≤                         |
| `facilities` | `ac,wifi`                       | kamar minimal punya fasilitas tsb (JSONB)|
| `min_rating` | `4`                             | rating rata-rata ≥                       |
| `sort`       | `created_at\|price\|rating`     | whitelist kolom                          |
| `order`      | `asc\|desc`                     | default `desc`                           |

Request contoh:

```
GET /api/v1/properties?page=1&limit=12&city=Bandung&min_price=500000&sort=rating&order=desc
```

Response `200`:

```json
{
  "success": true,
  "data": [
    {
      "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
      "name": "Kost Putra Mekar Jaya",
      "description": "Kost nyaman dekat kampus, full furnitur.",
      "address": "Jl. Mawar No. 12",
      "city": "Bandung",
      "status": "published",
      "rating_avg": "4.75",
      "rating_count": 8,
      "photo_url": "https://minio.local/kostify/properties/f47ac10b/cover.webp",
      "starting_price": 750000,
      "available_rooms": 3,
      "created_at": "2026-08-01T03:00:00Z"
    }
  ],
  "meta": { "page": 1, "limit": 12, "total": 25, "total_pages": 3 }
}
```

> `starting_price` = MIN(price_per_month) kamar available; `available_rooms` = COUNT kamar
> status `available`. Dihitung dengan satu query agregasi (subquery terindeks).

### `GET /api/v1/properties/:id` — Publik jika PUBLISHED; owner lihat miliknya apa pun statusnya

Response `200`:

```json
{
  "success": true,
  "data": {
    "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "owner_id": "6f9619ff-8b86-d011-b42d-00c04fc964ff",
    "name": "Kost Putra Mekar Jaya",
    "description": "Kost nyaman dekat kampus.",
    "address": "Jl. Mawar No. 12",
    "city": "Bandung",
    "status": "published",
    "rejection_reason": null,
    "rating_avg": "4.75",
    "rating_count": 8,
    "photos": [
      { "id": "...", "url": "https://...", "is_primary": true, "sort_order": 0 }
    ],
    "rooms": [
      {
        "id": "c9bf9e57-1685-4c89-bafb-ff5af830be8a",
        "room_number": "A01",
        "price_per_month": 750000,
        "area_m2": 16,
        "facilities": ["ac", "wifi", "private_bathroom"],
        "status": "available"
      }
    ],
    "reviews_summary": { "avg": 4.75, "count": 8 }
  }
}
```

Kost belum published diminta guest → `404` (jangan bocorkan keberadaan).

### `GET /api/v1/properties/owner` — Role owner

Listing milik owner sendiri, semua status. Query: `page, limit, search, status`.

Response `meta` + array seperti listing publik, plus `rejection_reason` dan jumlah kamar.

### `POST /api/v1/properties` — Role owner, CSRF

Membuat draft.

Request:

```json
{
  "name": "Kost Putri Melati",
  "description": "Kost putri aman dan bersih.",
  "address": "Jl. Kenanga No. 8",
  "city": "Depok"
}
```

Response `201`:

```json
{
  "success": true,
  "data": { "id": "e8b7a6c5-1234-4c89-bafb-ff5af830be8a", "status": "draft" },
  "message": "Draft kost berhasil dibuat"
}
```

### `PUT /api/v1/properties/:id` — Owner pemilik kost, CSRF

Update sebagian field (partial update). Ditolak `409` jika kost sedang
`PENDING_VERIFICATION` (harus tunggu keputusan admin atau batalkan submit).

Response `200` dengan data terbaru.

### `DELETE /api/v1/properties/:id` — Super Admin saja, CSRF

Hard delete (cascade ke foto, kamar, review). Booking histori tetap disimpan via
ON DELETE RESTRICT pada room_id booking → jika ada booking aktif, kembalikan
`409 CONFLICT`. Response `204 No Content`.

### `POST /api/v1/properties/:id/submit` — Owner pemilik, CSRF

Ajukan verifikasi. Validasi pre-condition:

- Status saat ini `DRAFT` atau `REJECTED`;
- Minimal **1 kamar**;
- Minimal **1 foto**;
- Field wajib lengkap (nama, alamat, kota).

Sukses: status → `PENDING_VERIFICATION`, catat `verification_logs(action=submitted)`,
notifikasi ke semua super admin.

Response `200`:

```json
{ "success": true, "data": { "status": "pending_verification" }, "message": "Kost diajukan untuk verifikasi" }
```

Gagal pre-condition `422`:

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Syarat submit belum terpenuhi",
    "details": [{ "field": "photos", "message": "minimal 1 foto harus diunggah" }]
  }
}
```

### `POST /api/v1/properties/:id/approve` — Super Admin, CSRF

Status → `PUBLISHED`, catat log `approved`, notifikasi ke owner. Response `200`.

### `POST /api/v1/properties/:id/reject` — Super Admin, CSRF

Request:

```json
{ "reason": "Foto tidak jelas, mohon unggah ulang foto depan bangunan" }
```

`reason` wajib (min 10 char) → validasi `422` jika kosong.
Status → `REJECTED`, log `rejected`, notifikasi ke owner. Response `200`.

---

## 4.4 Property Photos (upload)

### `POST /api/v1/properties/:id/photos` — Owner pemilik, CSRF, multipart/form-data

Form field: `file` (binary). Validasi file:

- MIME sniffing (bukan cuma ekstensi): `image/jpeg|png|webp`;
- Ukuran maksimal 5 MB;
- Max 10 foto per kost (`409` jika lebih);
- Nama objek di MinIO: `properties/{propertyId}/{uuid}.webp` (filename asli tidak dipakai — cegah path traversal);
- Upload langsung ke backend → backend stream ke MinIO (presigned URL tidak dipakai agar validasi tetap di backend).

Request contoh (multipart):

```
Content-Type: multipart/form-data; boundary=----Boundary

------Boundary
Content-Disposition: form-data; name="file"; filename="depan.jpg"
Content-Type: image/jpeg

<binary>
------Boundary--
```

Response `201`:

```json
{
  "success": true,
  "data": {
    "id": "aa1b2c3d-1111-2222-3333-444455556666",
    "url": "http://localhost:9000/kostify/properties/e8b7a6c5.../9f8e7d6c....jpg",
    "object_key": "properties/e8b7a6c5.../9f8e7d6c....jpg",
    "is_primary": true,
    "sort_order": 0
  },
  "message": "Foto berhasil diunggah"
}
```

### `DELETE /api/v1/properties/:id/photos/:photoId` — Owner pemilik, CSRF

Hapus metadata DB + object MinIO. Response `204`.

---

## 4.5 Rooms

### `GET /api/v1/properties/:propertyId/rooms`

Publik jika kost PUBLISHED (semua kamar); owner melihat miliknya apa pun status kost.
Query: `page`, `limit`, `status` (filter), `search` (room_number).

Response `200` (array ringkas):

```json
{
  "success": true,
  "data": [
    {
      "id": "c9bf9e57-1685-4c89-bafb-ff5af830be8a",
      "room_number": "A01",
      "price_per_month": 750000,
      "area_m2": 16,
      "description": "Full AC, kasur queen",
      "facilities": ["ac", "wifi", "private_bathroom"],
      "status": "available"
    }
  ],
  "meta": { "page": 1, "limit": 20, "total": 3, "total_pages": 1 }
}
```

### `POST /api/v1/properties/:propertyId/rooms` — Owner pemilik, CSRF

Request:

```json
{
  "room_number": "A02",
  "price_per_month": 850000,
  "area_m2": 18,
  "description": "Balkon menghadap taman",
  "facilities": ["ac", "wifi", "private_bathroom", "desk"]
}
```

Validasi: `room_number` required max 20, unik per kost (`409` bila duplikat);
`price_per_month` integer > 0 (misal 100_000–50_000_000); `area_m2` opsional > 0;
`facilities` array of enum `ac|wifi|private_bathroom|shared_bathroom|desk|wardrobe|balcony`.

Response `201` + data kamar.

### `PUT /api/v1/rooms/:id` — Owner pemilik, CSRF

Partial update field di atas. Tidak bisa mengubah `status` lewat endpoint ini
(ada endpoint khusus). Kamar yang sedang menopang booking aktif tetap boleh
update harga/deskripsi (booking lama pakai snapshot).

Response `200` + data terbaru.

### `DELETE /api/v1/rooms/:id` — Owner pemilik, CSRF

Tolak `409` bila kamar punya booking dengan status `pending|survey|booked|active`.
Response `204`.

### `PATCH /api/v1/rooms/:id/status` — Owner pemilik, CSRF

Owner hanya boleh toggle `available ↔ maintenance` (untuk renovasi).
Status lain dikelola otomatis oleh sistem booking.

Request:

```json
{ "status": "maintenance" }
```

Transisi ilegal (misal `available → booked`) → `409`:

```json
{
  "success": false,
  "error": {
    "code": "CONFLICT",
    "message": "Perubahan status manual hanya tersedia antara available dan maintenance"
  }
}
```

Response `200` + data kamar.
