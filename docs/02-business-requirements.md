# 2. Business Requirements

## 2.1 Role & Permission Matrix (RBAC)

| Aksi                                  | Guest | Tenant | Owner | Super Admin |
| ------------------------------------- | :---: | :----: | :---: | :---------: |
| Lihat daftar/detail kost PUBLISHED    |  ✅   |   ✅   |  ✅   |     ✅      |
| Register / Login                      |  ✅   |   ✅   |  ✅   |     ✅      |
| Booking kamar                         |  ❌   |   ✅   |  ❌   |     ❌      |
| Review kost (setelah COMPLETED)       |  ❌   |   ✅   |  ❌   |     ❌      |
| Wishlist                              |  ❌   |   ✅   |  ✅*  |     ❌      |
| CRUD kost milik sendiri               |  ❌   |   ❌   |  ✅   |     ❌      |
| Upload foto kost                      |  ❌   |   ❌   |  ✅   |     ❌      |
| CRUD kamar                            |  ❌   |   ❌   |  ✅   |     ❌      |
| Approve/reject/confirm booking        |  ❌   |   ❌   |  ✅   |     ❌      |
| Submit kost untuk verifikasi          |  ❌   |   ❌   |  ✅   |     ❌      |
| Verifikasi kost (approve/reject)      |  ❌   |   ❌   |  ❌   |     ✅      |
| Lihat semua user                      |  ❌   |   ❌   |  ❌   |     ✅      |
| Delete kost                           |  ❌   |   ❌   |  ❌   |     ✅      |

\* Wishlist dibuka juga untuk owner agar bisa memantau kost lain; implementasi tetap sama.

> **Prinsip:** seluruh otorisasi dievaluasi di **backend middleware + service layer**.
> Frontend hanya menyembunyikan/menampilkan UI sesuai role (UX layer), bukan pengaman.

## 2.2 Workflow Properti (Verifikasi)

```
                 submit (≥1 kamar & ≥1 foto)
  ┌───────┐  ─────────────────────────────►  ┌──────────────────────┐
  │ DRAFT │                                   │ PENDING_VERIFICATION │
  └───────┘  ◄─────────────────────────────  └──────────────────────┘
        edit ulang setelah ditolak / ditarik             │
              ▲                              approve ────┴──── reject (+reason wajib)
              │                                  ▼              ▼
       ┌──────────┐ ◄── owner aktifkan    ┌──────────┐    ┌──────────┐
       │ INACTIVE │                       │PUBLISHED │    │ REJECTED │
       └──────────┘ ────────────────────► └──────────┘    └──────────┘
```

**Aturan bisnis properti:**

1. Owner membuat kost sebagai `DRAFT`; draft tidak tampil publik.
2. Submit verifikasi mensyaratkan: minimal **1 kamar**, **1 foto**, dan field wajib terisi
   (nama, alamat, kota, deskripsi).
3. Hanya `SUPER_ADMIN` dapat approve (`PUBLISHED`) atau reject (`REJECTED`, wajib reason).
4. Kost `REJECTED` dapat diedit owner lalu disubmit ulang → kembali ke `PENDING_VERIFICATION`.
5. Owner dapat menonaktifkan kost `PUBLISHED` menjadi `INACTIVE`
   (hilang dari listing publik; booking existing tidak terpengaruh).
6. Setiap aksi verifikasi tercatat di tabel `verification_logs`.

## 2.3 Workflow Booking

```
 tenant create ──► ┌─────────┐  72 jam berlalu ┌─────────┐
                   │ PENDING │ ──────────────► │ EXPIRED │
                   └─────────┘                 └─────────┘
                     │        │
     tenant cancel   │        │ owner approve
                     ▼        ▼
               ┌───────────┐ ┌────────┐
               │ CANCELLED │ │ SURVEY │◄─── survey jadwal oleh owner & tenant
               └───────────┘ └────────┘
                               │        │
                owner reject   │        │ owner confirm
                     ▼         ▼        ▼
               ┌───────────┐        ┌────────┐
               │ REJECTED/ │        │ BOOKED │
               │ CANCELLED │        └────────┘
               └───────────┘             │ tenant check-in
                                         ▼
                                    ┌────────┐  tenant check-out  ┌───────────┐
                                    │ ACTIVE │ ─────────────────► │ COMPLETED │
                                    └────────┘                    └───────────┘
                                                                      │
                                                            tenant boleh review
```

**Aturan bisnis booking:**

1. Booking hanya boleh dibuat pada kamar berstatus `AVAILABLE`.
2. Saat booking dibuat: kamar `AVAILABLE → PENDING`, `expires_at = now() + 72 jam`.
3. **Auto-expiry:** worker scan berkala (tiap 60 detik); booking `PENDING` dengan
   `expires_at < now()` → `EXPIRED`, kamar kembali `AVAILABLE`, notifikasi dikirim.
4. Owner approve → booking `SURVEY`, kamar tetap di-hold (`SURVEY`).
5. Owner confirm (setelah survei selesai) → `BOOKED`, kamar → `BOOKED`.
6. Tenant check-in → `ACTIVE`, kamar → `ACTIVE`, tercatat `checked_in_at`.
7. Tenant check-out → `COMPLETED`, kamar → `AVAILABLE` (bisa disewa lagi),
   tercatat `checked_out_at`. Setelah status ini tenant boleh membuat review.
8. Tenant boleh cancel saat `PENDING` atau `SURVEY` → `CANCELLED`, kamar kembali `AVAILABLE`.
9. Owner reject saat `PENDING`/`SURVEY` → `REJECTED` (+reason), kamar kembali `AVAILABLE`.
10. **1 user maksimal 1 booking `PENDING` per kamar**
    (dicek service layer + partial unique index di DB).
11. Semua transisi dicatat di `booking_history` (from → to, aktor, catatan).
12. Transisi ilegal (contoh: check-out saat masih `SURVEY`) ditolak dengan `409 CONFLICT`.

## 2.4 Aturan Bisnis Lainnya

| #  | Aturan                                                                                                                              |
| -- | ------------------------------------------------------------------------------------------------------------------------------------ |
| 1  | Email unik per user. Password di-hash bcrypt (cost ≥ 10). Minimal 8 karakter, mengandung huruf & angka.                               |
| 2  | Review hanya dari tenant yang memiliki booking `COMPLETED` pada properti tersebut; 1 booking = 1 review (unique constraint).          |
| 3  | Rating bilangan bulat 1–5. Rating rata-rata properti di-update otomatis setelah review dibuat.                                        |
| 4  | Foto kost: maksimal 10 foto/kost, tiap file ≤ 5 MB, MIME `image/jpeg|image/png|image/webp`.                                          |
| 5  | Hanya `SUPER_ADMIN` dapat hard-delete kost. Owner cukup set `INACTIVE` (soft removal dari publik) agar histori booking tetap utuh.    |
| 6  | Rate limiting login: maks 5 percobaan / menit / IP → `429`.                                                                          |
| 7  | Notifikasi in-app dibuat untuk tiap event penting: booking dibuat/approve/reject/confirm/expire/checkin/checkout/cancel, hasil verifikasi, review baru. Email dikirim asinkron via queue Redis. |
| 8  | Harga kamar di-**snapshot** ke booking agar perubahan harga tidak memengaruhi transaksi lama.                                         |
| 9  | Booking hanya bisa dibuat pada kost berstatus `PUBLISHED`.                                                                            |
| 10 | Owner tidak bisa booking kost miliknya sendiri.                                                                                      |

## 2.5 Definition of Done (per fitur)

- [ ] Validasi input di backend (source of truth) + feedback inline di frontend.
- [ ] Otorisasi dievaluasi di backend.
- [ ] Status code HTTP tepat, response format konsisten.
- [ ] Unit test service layer (business rules).
- [ ] Terdokumentasi di Swagger/OpenAPI.
