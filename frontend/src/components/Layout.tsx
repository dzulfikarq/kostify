import { Link, Outlet, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { NotificationBell } from '../pages/NotificationsPage'

export function Layout() {
  const { user, logout } = useAuth()
  const nav = useNavigate()

  async function onLogout() {
    await logout()
    nav('/')
  }

  return (
    <div className="flex min-h-screen flex-col bg-slate-50">
      <header className="border-b bg-white">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-4 py-3">
          <Link to="/" className="text-xl font-bold text-teal-700">Kostify</Link>
          <nav className="flex items-center gap-4 text-sm">
            <Link to="/properties" className="hover:text-teal-700">Cari Kost</Link>
            {user?.role === 'tenant' && (
              <>
                <Link to="/tenant/bookings" className="hover:text-teal-700">Booking Saya</Link>
                <Link to="/tenant/wishlist" className="hover:text-teal-700">Wishlist</Link>
              </>
            )}
            {user?.role === 'owner' && (
              <>
                <Link to="/owner/properties" className="hover:text-teal-700">Kost Saya</Link>
                <Link to="/owner/bookings" className="hover:text-teal-700">Booking Masuk</Link>
                <Link to="/dashboard" className="hover:text-teal-700">Dashboard</Link>
              </>
            )}
            {user?.role === 'super_admin' && (
              <>
                <Link to="/admin/verifications" className="hover:text-teal-700">Verifikasi</Link>
                <Link to="/dashboard" className="hover:text-teal-700">Dashboard</Link>
              </>
            )}
            {user && <NotificationBell />}
            {user?.role === 'super_admin' && (
              <Link to="/admin/verifications" className="hover:text-teal-700">Verifikasi</Link>
            )}
            {user ? (
              <>
                <span className="text-slate-500">{user.name}</span>
                <button onClick={onLogout} className="rounded-md border px-3 py-1.5 hover:bg-slate-50">
                  Keluar
                </button>
              </>
            ) : (
              <>
                <Link to="/login" className="rounded-md border px-3 py-1.5 hover:bg-slate-50">Masuk</Link>
                <Link to="/register" className="rounded-md bg-teal-600 px-3 py-1.5 text-white hover:bg-teal-700">
                  Daftar
                </Link>
              </>
            )}
          </nav>
        </div>
      </header>
      <main className="flex-1">
        <Outlet />
      </main>
      <footer className="border-t bg-white py-4 text-center text-xs text-slate-400">
        Kostify — Take Home Test Fullstack Engineer
      </footer>
    </div>
  )
}
