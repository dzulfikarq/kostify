import { useQuery } from '@tanstack/react-query'
import { Link, useParams } from 'react-router-dom'
import { api } from '../lib/api'
import { StatusBadge } from '../components/StatusBadge'
import type { PropertyDetail } from '../lib/types'
import { formatIDR } from './PropertiesPage'

export function PropertyDetailPage() {
  const { id } = useParams()
  const { data: p, isLoading, isError } = useQuery({
    queryKey: ['property', id],
    queryFn: async () => (await api.get(`/properties/${id}`)).data.data as PropertyDetail,
  })

  if (isLoading) return <div className="mx-auto max-w-6xl px-4 py-10 text-slate-500">Memuat…</div>
  if (isError || !p)
    return (
      <div className="mx-auto max-w-6xl px-4 py-10 text-center">
        <p className="text-lg font-semibold">Kost tidak ditemukan</p>
        <Link to="/properties" className="text-teal-700 underline">Kembali ke daftar</Link>
      </div>
    )

  return (
    <div className="mx-auto max-w-6xl px-4 py-6">
      <div className="grid gap-2 sm:grid-cols-3">
        {p.photos.length > 0 ? (
          p.photos.map((ph) => (
            <img
              key={ph.id}
              src={ph.url}
              alt={p.name}
              className={`h-56 w-full rounded-xl object-cover ${ph.is_primary ? 'sm:col-span-2 sm:row-span-2 h-[464px]' : ''}`}
            />
          ))
        ) : (
          <div className="col-span-3 flex h-56 items-center justify-center rounded-xl bg-slate-100 text-slate-400">
            Tanpa foto
          </div>
        )}
      </div>

      <div className="mt-6 flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold">{p.name}</h1>
          <p className="text-slate-500">{p.address}, {p.city}</p>
          <div className="mt-1 flex items-center gap-2">
            <StatusBadge status={p.status} />
            <span className="text-sm text-slate-500">★ {p.reviews_summary.avg.toFixed(1)} ({p.reviews_summary.count} ulasan)</span>
          </div>
        </div>
      </div>

      {p.description && <p className="mt-4 max-w-3xl whitespace-pre-line text-slate-700">{p.description}</p>}

      <h2 className="mt-8 mb-3 text-xl font-bold">Kamar Tersedia</h2>
      <div className="grid gap-4 md:grid-cols-2">
        {p.rooms.map((r) => (
          <div key={r.id} className="rounded-xl border bg-white p-4 shadow-sm">
            <div className="flex items-center justify-between">
              <h3 className="font-semibold">Kamar {r.room_number}</h3>
              <StatusBadge status={r.status} />
            </div>
            <p className="mt-1 font-bold text-teal-700">{formatIDR(r.price_per_month)}/bln</p>
            {r.area_m2 && <p className="text-xs text-slate-500">Luas {r.area_m2} m²</p>}
            {r.description && <p className="mt-1 text-sm text-slate-600">{r.description}</p>}
            {r.facilities.length > 0 && (
              <div className="mt-2 flex flex-wrap gap-1">
                {r.facilities.map((f) => (
                  <span key={f} className="rounded-full bg-teal-50 px-2 py-0.5 text-xs text-teal-800">{f}</span>
                ))}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
