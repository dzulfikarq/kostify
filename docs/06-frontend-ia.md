# 6. Frontend Information Architecture

Stack: **React + TypeScript + Vite + Tailwind CSS + Axios + React Router**.
State server: React Query (cache, loading/error state otomatis). State auth: Context API.

## 6.1 Route Map

```
/                               → HomePage (landing + search hero)
/properties                     → PropertiesPage (listing publik: search, filter, sort, pagination)
/properties/:id                 → PropertyDetailPage (galeri foto, kamar, review)
/login                          → LoginPage
/register                       → RegisterPage (pilih role owner/tenant)

/dashboard                      → redirect per role:
  /tenant/dashboard             → TenantDashboard
  /tenant/bookings              → TenantBookings (tab per status; aksi cancel/checkin/checkout/review)
  /tenant/wishlist              → TenantWishlist

  /owner/dashboard              → OwnerDashboard (okupansi, pending booking, revenue estimasi)
  /owner/properties             → OwnerProperties (tabel semua status + badge status)
  /owner/properties/new         → PropertyFormPage (create draft)
  /owner/properties/:id/edit    → PropertyFormPage (edit + kelola foto + kelola kamar + submit verifikasi)
  /owner/properties/:id/rooms   → RoomManagerPage (CRUD kamar, toggle maintenance)
  /owner/bookings               → OwnerBookings (approve/reject/confirm)

  /admin/dashboard              → AdminDashboard (statistik platform)
  /admin/verifications          → VerificationQueuePage (detail kost → approve/reject+reason)
  /admin/users                  → UsersPage (daftar user, filter role)

/notifications                  → NotificationsPage (semua role)
/403                            → ForbiddenPage
/*                              → NotFoundPage
```

## 6.2 Protected Routes

```
PublicRoute (guest only)   : /login, /register
ProtectedRoute (any auth)  : /dashboard/*, /notifications
TenantOnly                 : /tenant/*
OwnerOnly                  : /owner/*
AdminOnly                  : /super admin/*
```

Implementasi: komponen `<RequireAuth roles={[...]}>` membaca `AuthContext`.
Belum login → redirect `/login?next=...`. Login tapi role salah → render `403`.

> Frontend guard hanya UX. Keamanan sesungguhnya di middleware backend.

## 6.3 Layout

- **Public layout**: Header (logo, nav, tombol login/register atau avatar menu) + Footer.
- **Dashboard layout**: Sidebar (menu sesuai role) + Topbar (notifikasi bell dengan
  unread count, avatar dropdown profil/logout) + content area. Mobile: sidebar jadi drawer.

Menu sidebar per role:

| Tenant            | Owner                 | Super Admin       |
| ----------------- | --------------------- | ------------------ |
| Dashboard         | Dashboard             | Dashboard          |
| Booking Saya      | Kost Saya             | Verifikasi         |
| Wishlist          | Kamar (per kost)      | Users              |
| Jelajah Kost      | Booking Masuk         |                    |

## 6.4 Halaman Kunci — Struktur UI

### PropertiesPage (listing publik)

```
┌──────────────────────────────────────────────────────────┐
│ [Search bar besar..............................] [Cari] │
│ ┌─Filter Sidebar─────┐  ┌─Grid hasil (3 kolom desktop)─┐│
│ │ Kota   [select]    │  │ [Card] [Card] [Card]        ││
│ │ Harga min/max      │  │ [Card] [Card] [Card]        ││
│ │ Fasilitas ☑AC ☑Wifi│  │ ...                         ││
│ │ Rating ★★★★☆+     │  ├─────────────────────────────┤│
│ │ [Reset]            │  │ Pagination ‹ 1 2 3 ›        ││
│ └────────────────────┘  └─────────────────────────────┘│
```

State filter di URL query string (`useSearchParams`) → shareable link + back button works.
Skeleton card saat loading; EmptyState "Tidak ada kost ditemukan" + tombol reset filter;
Error state + tombol coba lagi.

### PropertyDetailPage

Galeri foto (lightbox), info utama + badge PUBLISHED, daftar kamar (card per kamar,
badge status available/pending/booked/maintenance, tombol "Booking" disabled bila bukan
available/bukan tenant), section review (rata-rata + list + pagination),
tombol wishlist (hati), sticky panel harga di desktop.

### OwnerPropertyEditPage

Tab: **Info** (form) · **Foto** (upload multi, drag reorder sederhana, set cover, delete,
progress bar) · **Kamar** (tabel CRUD inline/modal) · **Verifikasi** (status badge +
alasan penolakan bila REJECTED + tombol "Ajukan Verifikasi" disabled sampai syarat terpenuhi).

### OwnerBookingsPage

Filter tab: Semua/Pending/Survey/Booked/Active/Completed. Tabel: tenant, kost/kamar,
harga, durasi, status badge, countdown expiry (untuk pending), aksi kontekstual:
Approve/Reject (modal alasan) saat pending; Confirm (date picker start_date) saat survey.

### VerificationQueuePage (Super Admin)

Daftar kost `pending_verification` (terlama dulu). Detail: semua field, foto gallery,
kamar, owner info. Tombol Approve (confirm dialog) dan Reject (textarea reason wajib).

## 6.5 Form UX Standard (semua form)

- Validasi inline saat blur + saat submit (zod schema dibagi antara client check).
- Error message merah di bawah field, aria-invalid untuk aksesibilitas.
- Submit button: disabled saat submitting, spinner di dalam tombol.
- Sukses → toast hijau + navigasi/update cache.
- Server validation error (422) dipetakan ke field terkait lewat `error.details[]`.

## 6.6 Reusable Components (`src/components/common`)

| Komponen    | Tanggung jawab singkat                                                     |
| ----------- | --------------------------------------------------------------------------- |
| Button      | variant primary/secondary/danger/ghost, size, loading, disabled             |
| Input       | label, error text, prefix icon                                              |
| Select      | native select styled                                                        |
| Textarea    | seperti Input                                                               |
| Modal       | overlay, esc close, focus trap dasar                                        |
| ConfirmDialog| modal khusus destructive (merah), wajib konfirmasi                        |
| Drawer      | panel slide (filter mobile, notifikasi mobile)                              |
| Table       | generik kolom + row actions                                                 |
| Pagination  | page controls + info total                                                  |
| Dropdown    | menu kecil (avatar, row actions)                                            |
| Badge       | warna per status (published=hijau, pending=kuning, rejected=merah, dst)     |
| Card        | container dasar                                                             |
| Toast       | global via context, auto dismiss                                            |
| Spinner     | loading inline                                                              |
| Skeleton    | placeholder card/table                                                      |
| EmptyState  | ikon + judul + deskripsi + CTA opsional                                     |
| StatusBadge | mapping enum status → warna & label Indonesia                               |

## 6.7 Axios Instance & Interceptor

```ts
// services/api.ts (desain)
const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL + "/api/v1",
  withCredentials: true, // kirim cookie HttpOnly
});

api.interceptors.request.use((config) => {
  const token = getCsrfCookie();
  if (token && isMutating(config.method)) config.headers["X-CSRF-Token"] = token;
  return config;
});

let refreshPromise: Promise<void> | null = null;

api.interceptors.response.use(undefined, async (error) => {
  const { response, config } = error;

  if (response?.status === 401 && !config._retried) {
    config._retried = true;
    // single-flight refresh: request concurrent menunggu promise yang sama
    refreshPromise ??= api.post("/auth/refresh").finally(() => (refreshPromise = null));
    try {
      await refreshPromise;
      return api(config); // retry sekali
    } catch {
      useAuth.getState().logout(); // refresh gagal → logout paksa + redirect /login
      return Promise.reject(error);
    }
  }

  switch (response?.status) {
    case 403: toast.error("Anda tidak memiliki akses"); break;
    case 422: throw mapValidationErrors(response); // dilempar ke form terkait
    case 429: toast.error("Terlalu banyak permintaan. Coba beberapa saat lagi."); break;
    case 500: toast.error("Terjadi kesalahan server. Coba lagi nanti."); break;
  }
  return Promise.reject(error);
});
```

Poin anti-jebakan refresh:

- **Single-flight**: satu promise refresh dibagi ke semua request concurrent;
- **Retry sekali** via flag `_retried` → tidak infinite loop;
- **Refresh gagal → logout bersih** (state dibuang, redirect login);
- Request non-GET setelah refresh memakai csrf token baru dari cookie.

## 6.8 Auth Flow Frontend

1. App mount → `GET /auth/me` (React Query). Sukses → user tersimpan di AuthContext.
2. `401` → user guest, tampilkan public shell.
3. Login sukses → set user, redirect `next` param atau dashboard per role.
4. Logout → `POST /auth/logout` → clear state → redirect home.
5. Role-based UI: `user.role === 'owner'` mengontrol menu/tombol; tetap diasumsikan
   backend adalah gate sebenarnya.
