import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { api } from '../lib/api'
import type { PropertySummary } from '../lib/types'

export function formatIDR(n: number): string {
  return 'Rp ' + n.toLocaleString('id-ID')
}

function PropertyCard({ p }: { p: PropertySummary }) {
  return (
    <Link
      to={`/properties/${p.id}`}
      className="overflow-hidden rounded-xl border bg-white shadow-sm transition hover:shadow-md"
    >
      {p.photo_url ? (
        <img src={p.photo_url} alt={p.name} className="h-44 w-full object-cover" />
      ) : (
        <div className="flex h-44 items-center justify-center bg-slate-100 text-slate-400">Tanpa foto</div>
      )}
      <div className="space-y-1 p-4">
        <h3 className="font-semibold text-slate-800">{p.name}</h3>
        <p className="text-sm text-slate-500">{p.city}</p>
        <div className="flex items-center justify-between pt-1">
          <span className="text-sm font-semibold text-teal-700">
            {p.starting_price ? `${formatIDR(p.starting_price)}/bln` : '—'}
          </span>
          <span className="text-xs text-slate-500">
            ★ {p.rating_avg > 0 ? p.rating_avg.toFixed(1) : '-'} ({p.rating_count})
          </span>
        </div>
      </div>
    </Link>
  )
}

export function PropertiesPage() {
  const params = new URLSearchParams(location.search)
  const qs = params.toString()

  const { data, isLoading, isError } = useQuery({
    queryKey: ['properties', qs],
    queryFn: async () => {
      const r = await api.get(`/properties${qs ? `?${qs}` : ''}`)
      return r.data as { data: PropertySummary[]; meta: { total: number } }
    },
  })

  return (
    <div className="mx-auto max-w-6xl px-4 py-6">
      <h1 className="mb-4 text-2xl font-bold">Jelajah Kost</h1>
      {isLoading && <p className="text-slate-500">Memuat…</p>}
      {isError && <p className="text-red-600">Gagal memuat data.</p>}
      {data && (
        <>
          <p className="mb-3 text-sm text-slate-500">{data.meta.total} kost ditemukan</p>
          {data.data.length === 0 ? (
            <div className="rounded-xl border border-dashed p-10 text-center text-slate-500">
              Tidak ada kost ditemukan.
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {data.data.map((p) => (
                <PropertyCard key={p.id} p={p} />
              ))}
            </div>
          )}
        </>
      )}
    </div>
  )
}
