# 1. Product Foundation

## 1.1 Problem Statement

Pengelolaan kost di Indonesia masih sangat manual:

- **Pemilik kost** mengelola kamar, penyewa, dan pembayaran lewat catatan buku / WhatsApp.
  Tidak ada satu sumber data untuk status kamar dan riwayat penyewaan.
- **Calon penghuni** sulit menemukan kost yang terpercaya — informasi tersebar di grup
  Facebook/Telegram tanpa verifikasi, sering sudah penuh ketika dihubungi.
- **Proses booking** tidak terstruktur: nego via chat, tidak ada jejak kesepakatan,
  pemilik bisa tidak sengaja menerima dua orang untuk kamar yang sama.

## 1.2 Solusi

Kostify adalah platform web yang:

1. Memungkinkan **pemilik kost** mendaftarkan properti beserta kamar dan foto.
2. Menjamin **setiap kost diverifikasi Super Admin** sebelum tampil publik → kepercayaan calon penghuni.
3. Menyediakan **alur booking terstruktur** dengan status jelas dan expiry otomatis (72 jam),
   sehingga kamar tidak "terkunci" oleh booking yang tidak kunjung diproses.
4. Mencatat seluruh perubahan status pada **audit trail** (booking history, verification logs).

## 1.3 Target User

| Peran       | Deskripsi                   | Nilai utama                                         |
| ----------- | --------------------------- | --------------------------------------------------- |
| Super Admin | Pengelola platform          | Verifikasi kost, kelola user, pantau aktivitas      |
| Owner       | Pemilik kost (1–N properti) | Kelola kamar, proses booking, lihat statistik       |
| Tenant      | Calon penghuni              | Cari kost terverifikasi, booking, review, wishlist  |

## 1.4 Value Proposition

- **Tenant**: kost tampil publik hanya jika sudah terverifikasi → mengurangi risiko penipuan;
  status booking transparan dengan tenggat 72 jam.
- **Owner**: alur booking otomatis (expiry, notifikasi) → kamar tidak double-booked;
  dashboard ringkasan okupansi.
- **Platform**: audit trail penuh atas setiap transaksi dan keputusan verifikasi.

## 1.5 Scope

**In scope:**

- Auth cookie-based + CSRF, RBAC 3 role.
- CRUD kost & kamar + upload foto (MinIO).
- Workflow verifikasi kost.
- Workflow booking lengkap dengan auto-expiry 72 jam.
- Review & rating pasca sewa.
- Wishlist, notifikasi in-app + email via queue, dashboard per role.
- Search/filter/sort/pagination pada listing.

**Out of scope (future):**

- Payment gateway / tagihan bulanan otomatis.
- Chat realtime owner–tenant.
- Mobile app native.
