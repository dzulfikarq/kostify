import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { api, errMessage } from '../lib/api'
import { useAuth } from '../context/AuthContext'
import { StatusBadge } from '../components/StatusBadge'

export function DashboardPage() {
  const { user } = useAuth()
  const { data, isLoading, isError, error } = useQuery({
    queryKey: ['dashboard'],
    queryFn: async () => (await api.get('/dashboard')).data.data as Record<string, any>,
  })

  if (isLoading) return <div className="mx-auto max-w-5xl px-4 py-6 text-slate-500">Memuat…</div>
  if (isError) return <div className="mx-auto max-w-5xl px-4 py-6 text-red-600">{errMessage(error)}</div>
  if (!data) return null

  const title =
    user?.role === 'owner' ? 'Dashboard Pemilik' :
    user?.role === 'super_admin' ? 'Dashboard Admin' : 'Dashboard Penyewa'

  return (
    <div className="mx-auto max-w-5xl px-4 py-6">
      <h1 className="mb-4 text-2xl font-bold">{title}</h1>

      {(data.role === 'owner' || data.role === 'super_admin') && (
        <div className="grid gap-3 sm:grid-cols-3">
          {data.role === 'owner' ? (
            <>
              <Stat label="Total Kost" value={data.total_properties} />
              <Stat label="Tayang" value={data.published} />
              <Stat label="Menunggu Verifikasi" value={data.pending_verification} />
              <Stat label="Total Kamar" value={data.total_rooms} />
              <Stat label="Okupansi" value={`${Number(data.occupancy_rate).toFixed(1)}%`} />
              <Stat label="Estimasi Revenue/Bln" value={`Rp ${Number(data.revenue_estimation_monthly ?? 0).toLocaleString('id-ID')}`} />
            </>
          ) : (
            <>
              <Stat label="Total Users" value={data.users_total} />
              <Stat label="Total Kost" value={data.properties_total} />
              <Stat label="Menunggu Verifikasi" value={data.waiting_verification} />
              <Stat label="Booking Aktif" value={data.bookings_active} />
              <Stat label="Booking Bulan Ini" value={data.bookings_this_month} />
            </>
          )}
        </div>
      )}

      {data.role === 'tenant' && (
        <div className="grid gap-3 sm:grid-cols-2">
          <Stat label="Booking Pending" value={data.pending_bookings} />
          <Stat label="Wishlist" value={data.wishlist_count} />
        </div>
      )}

      {data.role === 'owner' && Array.isArray(data.recent_bookings) && data.recent_bookings.length > 0 && (
        <>
          <h2 className="mt-8 mb-3 text-xl font-bold">Booking Terbaru</h2>
          <div className="overflow-hidden rounded-xl border bg-white shadow-sm">
            <table className="w-full text-sm">
              <thead className="bg-slate-50 text-left text-xs uppercase text-slate-500">
                <tr><th className="px-4 py-2">Kamar</th><th className="px-4 py-2">Harga</th><th className="px-4 py-2">Status</th></tr>
              </thead>
              <tbody>
                {(data.recent_bookings as Record<string, unknown>[]).map((b) => (
                  <tr key={String(b.id)} className="border-t">
                    <td className="px-4 py-2">{String(b.room_number)}</td>
                    <td className="px-4 py-2">Rp {Number(b.price_per_month).toLocaleString('id-ID')}</td>
                    <td className="px-4 py-2"><StatusBadge status={String(b.status) as never} /></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      )}

      {data.role === 'super_admin' && (
        <p className="mt-6">
          <Link to="/admin/verifications" className="text-teal-700 underline">Lihat antrean verifikasi →</Link>
        </p>
      )}
    </div>
  )
}

function Stat({ label, value }: { label: string; value: unknown }) {
  return (
    <div className="rounded-xl border bg-white p-4 shadow-sm">
      <p className="text-xs uppercase tracking-wide text-slate-400">{label}</p>
      <p className="mt-1 text-2xl font-bold text-slate-800">{String(value ?? '-')}</p>
    </div>
  )
}
