import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { api, errMessage } from '../lib/api'
import type { PropertySummary } from '../lib/types'
import { formatIDR } from './PropertiesPage'

export function TenantWishlistPage() {
  const qc = useQueryClient()
  const [error, setError] = useState('')
  const { data, isLoading } = useQuery({
    queryKey: ['wishlist'],
    queryFn: async () => (await api.get('/wishlist')).data as { data: PropertySummary[] },
  })

  const remove = useMutation({
    mutationFn: async (id: string) => await api.delete(`/wishlist/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['wishlist'] }),
    onError: (e) => setError(errMessage(e)),
  })

  if (isLoading) return <div className="mx-auto max-w-5xl px-4 py-6 text-slate-500">Memuat…</div>

  return (
    <div className="mx-auto max-w-5xl px-4 py-6">
      <h1 className="mb-4 text-2xl font-bold">Wishlist</h1>
      {error && <p className="mb-3 rounded-md bg-red-50 p-2 text-sm text-red-700">{error}</p>}
      {data && data.data.length === 0 && (
        <div className="rounded-xl border border-dashed p-10 text-center text-slate-500">
          Belum ada kost di wishlist.
        </div>
      )}
      <div className="space-y-3">
        {data?.data.map((p) => (
          <div key={p.id} className="flex items-center justify-between rounded-xl border bg-white p-4 shadow-sm">
            <Link to={`/properties/${p.id}`} className="flex items-center gap-3">
              {p.photo_url ? (
                <img src={p.photo_url} alt="" className="h-14 w-20 rounded-lg object-cover" />
              ) : (
                <div className="h-14 w-20 rounded-lg bg-slate-100" />
              )}
              <div>
                <p className="font-semibold">{p.name}</p>
                <p className="text-sm text-slate-500">{p.city} · {formatIDR(p.starting_price ?? 0)}/bln</p>
              </div>
            </Link>
            <button
              onClick={() => remove.mutate(p.id)}
              disabled={remove.isPending}
              className="text-sm text-red-600 hover:underline disabled:opacity-50"
            >
              Hapus
            </button>
          </div>
        ))}
      </div>
    </div>
  )
}
