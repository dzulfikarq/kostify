import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { api } from '../../lib/api'
import { StatusBadge } from '../../components/StatusBadge'
import type { OwnerProperty } from '../../lib/types'

export function OwnerPropertiesPage() {
  const qc = useQueryClient()
  const { data, isLoading } = useQuery({
    queryKey: ['owner-properties'],
    queryFn: async () => (await api.get('/properties/owner')).data as { data: OwnerProperty[] },
  })

  return (
    <div className="mx-auto max-w-5xl px-4 py-6">
      <div className="mb-4 flex items-center justify-between">
        <h1 className="text-2xl font-bold">Kost Saya</h1>
        <Link
          to="/owner/properties/new"
          className="rounded-md bg-teal-600 px-4 py-2 text-sm text-white hover:bg-teal-700"
        >
          + Tambah Kost
        </Link>
      </div>

      <button hidden onClick={() => qc.invalidateQueries({ queryKey: ['owner-properties'] })} />

      {isLoading && <p className="text-slate-500">Memuat…</p>}
      {data && data.data.length === 0 && (
        <div className="rounded-xl border border-dashed p-10 text-center text-slate-500">
          Belum ada kost. Klik “Tambah Kost” untuk mulai.
        </div>
      )}
      <div className="space-y-3">
        {data?.data.map((p) => (
          <Link
            key={p.id}
            to={`/owner/properties/${p.id}/edit`}
            className="flex items-center justify-between rounded-xl border bg-white p-4 shadow-sm hover:shadow-md"
          >
            <div>
              <p className="font-semibold">{p.name}</p>
              <p className="text-sm text-slate-500">{p.city}</p>
              {p.status === 'rejected' && p.rejection_reason && (
                <p className="mt-1 text-xs text-red-600">Alasan: {p.rejection_reason}</p>
              )}
            </div>
            <StatusBadge status={p.status} />
          </Link>
        ))}
      </div>
    </div>
  )
}
