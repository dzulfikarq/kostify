import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { api } from '../lib/api'
import type { PropertySummary } from '../lib/types'

export function HomePage() {
  const { data } = useQuery({
    queryKey: ['properties', 'featured'],
    queryFn: async () =>
      (await api.get('/properties?sort=rating&order=desc&limit=3')).data as {
        data: PropertySummary[]
      },
  })

  return (
    <div>
      <section className="bg-teal-700 py-20 text-center text-white">
        <h1 className="text-4xl font-bold">Cari Kost Idamanmu</h1>
        <p className="mt-2 text-teal-100">Ribuan pilihan kost nyaman di seluruh Indonesia</p>
        <Link
          to="/properties"
          className="mt-6 inline-block rounded-lg bg-white px-6 py-2.5 font-semibold text-teal-700 hover:bg-teal-50"
        >
          Jelajah Sekarang
        </Link>
      </section>

      {data && data.data.length > 0 && (
        <section className="mx-auto max-w-6xl px-4 py-10">
          <h2 className="mb-4 text-xl font-bold">Kost Terbaik</h2>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {data.data.map((p) => (
              <Link
                key={p.id}
                to={`/properties/${p.id}`}
                className="overflow-hidden rounded-xl border bg-white shadow-sm hover:shadow-md"
              >
                {p.photo_url ? (
                  <img src={p.photo_url} alt={p.name} className="h-40 w-full object-cover" />
                ) : (
                  <div className="flex h-40 items-center justify-center bg-slate-100 text-slate-400">—</div>
                )}
                <div className="p-4">
                  <p className="font-semibold">{p.name}</p>
                  <p className="text-sm text-slate-500">
                    {p.city} · ★ {p.rating_avg > 0 ? p.rating_avg.toFixed(1) : '-'}
                  </p>
                </div>
              </Link>
            ))}
          </div>
        </section>
      )}
    </div>
  )
}
