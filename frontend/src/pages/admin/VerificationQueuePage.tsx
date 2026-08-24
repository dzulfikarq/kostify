import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api, errMessage } from '../../lib/api'
import { StatusBadge } from '../../components/StatusBadge'
import type { PropertyDetail, PropertySummary } from '../../lib/types'

export function VerificationQueuePage() {
  const qc = useQueryClient()
  const [selected, setSelected] = useState<string | null>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['pending-properties'],
    queryFn: async () => (await api.get('/properties?sort=created_at&order=asc&limit=100')).data as {
      data: PropertySummary[]
    },
  })

  async function act(id: string, action: 'approve' | 'reject', reason?: string) {
    if (action === 'reject') await api.post(`/properties/${id}/reject`, { reason })
    else await api.post(`/properties/${id}/approve`)
    await qc.invalidateQueries({ queryKey: ['pending-properties'] })
    setSelected(null)
  }

  return (
    <div className="mx-auto max-w-5xl px-4 py-6">
      <h1 className="mb-4 text-2xl font-bold">Antrean Verifikasi</h1>
      {isLoading && <p className="text-slate-500">Memuat…</p>}
      {data && data.data.filter((p) => p.status === 'pending_verification').length === 0 && (
        <div className="rounded-xl border border-dashed p-10 text-center text-slate-500">
          Tidak ada kost yang menunggu verifikasi.
        </div>
      )}
      <div className="space-y-3">
        {data?.data
          .filter((p) => p.status === 'pending_verification')
          .map((p) => (
            <div key={p.id} className="rounded-xl border bg-white p-4 shadow-sm">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  {p.photo_url ? (
                    <img src={p.photo_url} alt="" className="h-14 w-20 rounded-lg object-cover" />
                  ) : (
                    <div className="flex h-14 w-20 items-center justify-center rounded-lg bg-slate-100 text-xs text-slate-400">
                      —
                    </div>
                  )}
                  <div>
                    <p className="font-semibold">{p.name}</p>
                    <p className="text-sm text-slate-500">
                      {p.city} · diajukan {new Date(p.created_at).toLocaleDateString('id-ID')}
                    </p>
                  </div>
                </div>
                <button
                  onClick={() => setSelected(selected === p.id ? null : p.id)}
                  className="rounded-md border px-3 py-1.5 text-sm hover:bg-slate-50"
                >
                  {selected === p.id ? 'Tutup' : 'Review'}
                </button>
              </div>
              {selected === p.id && <VerificationDetail id={p.id} onAct={act} />}
            </div>
          ))}
      </div>
    </div>
  )
}

function VerificationDetail({
  id,
  onAct,
}: {
  id: string
  onAct: (id: string, action: 'approve' | 'reject', reason?: string) => Promise<void>
}) {
  const [reason, setReason] = useState('')
  const [error, setError] = useState('')
  const approve = useMutation({ mutationFn: () => onAct(id, 'approve') })
  const reject = useMutation({
    mutationFn: () => onAct(id, 'reject', reason),
    onError: (e) => setError(errMessage(e)),
  })

  const { data: p } = useQuery({
    queryKey: ['verify-property', id],
    queryFn: async () => (await api.get(`/properties/${id}`)).data.data as PropertyDetail,
  })

  return (
    <div className="mt-4 space-y-3 border-t pt-4">
      {error && <p className="rounded-md bg-red-50 p-2 text-sm text-red-700">{error}</p>}
      {!p ? (
        <p className="text-slate-500">Memuat detail…</p>
      ) : (
        <>
          <div className="grid gap-2 sm:grid-cols-4">
            {p.photos.map((ph) => (
              <img key={ph.id} src={ph.url} alt="" className="h-32 w-full rounded-lg object-cover" />
            ))}
          </div>
          <p className="text-sm"><span className="font-medium">Alamat:</span> {p.address}, {p.city}</p>
          {p.description && <p className="text-sm text-slate-600">{p.description}</p>}
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-xs uppercase text-slate-500">
                <th className="py-1">Kamar</th>
                <th className="py-1">Harga</th>
                <th className="py-1">Fasilitas</th>
                <th className="py-1">Status</th>
              </tr>
            </thead>
            <tbody>
              {p.rooms.map((r) => (
                <tr key={r.id} className="border-t">
                  <td className="py-1.5">{r.room_number}</td>
                  <td>Rp {r.price_per_month.toLocaleString('id-ID')}</td>
                  <td>{r.facilities.join(', ') || '—'}</td>
                  <td><StatusBadge status={r.status} /></td>
                </tr>
              ))}
            </tbody>
          </table>
        </>
      )}
      <div className="flex flex-wrap items-start gap-2 pt-2">
        <button
          onClick={() => approve.mutate()}
          disabled={approve.isPending}
          className="rounded-md bg-green-600 px-4 py-2 text-sm text-white hover:bg-green-700 disabled:opacity-50"
        >
          Setujui & Tayangkan
        </button>
        <input
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          placeholder="Alasan penolakan (min. 10 karakter)"
          className="min-w-64 flex-1 rounded-md border px-3 py-2 text-sm focus:border-teal-500 focus:outline-none"
        />
        <button
          onClick={() => reject.mutate()}
          disabled={reject.isPending || reason.trim().length < 10}
          title={reason.trim().length < 10 ? 'Minimal 10 karakter' : undefined}
          className="rounded-md bg-red-600 px-4 py-2 text-sm text-white hover:bg-red-700 disabled:opacity-50"
        >
          Tolak
        </button>
      </div>
    </div>
  )
}
