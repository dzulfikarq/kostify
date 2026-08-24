import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api, errMessage } from '../../lib/api'
import type { BookingWithRefs } from '../../lib/types'
import { BookingCard } from './TenantBookingsPage'

const TABS = [
  ['all', 'Semua'],
  ['pending', 'Pending'],
  ['survey', 'Survei'],
  ['booked', 'Dibooking'],
  ['active', 'Disewa'],
  ['completed', 'Selesai'],
] as const

export function OwnerBookingsPage() {
  const qc = useQueryClient()
  const [tab, setTab] = useState<string>('all')
  const [error, setError] = useState('')

  const qs = tab === 'all' ? '' : `?status=${tab}`
  const { data, isLoading } = useQuery({
    queryKey: ['owner-bookings', tab],
    queryFn: async () => (await api.get(`/bookings/owner${qs}`)).data as { data: BookingWithRefs[] },
  })

  const refresh = () => qc.invalidateQueries({ queryKey: ['owner-bookings'] })

  const act = useMutation({
    mutationFn: async ({ id, action, body }: { id: string; action: string; body?: unknown }) => {
      await api.put(`/bookings/${id}/${action}`, body ?? {})
    },
    onSuccess: () => {
      setError('')
      refresh()
    },
    onError: (e) => setError(errMessage(e)),
  })

  function reject(id: string, reason: string) {
    if (reason.trim().length < 10) {
      setError('Alasan penolakan minimal 10 karakter.')
      return
    }
    act.mutate({ id, action: 'reject', body: { reason } })
  }

  function confirm(id: string, startDate: string) {
    if (!/^\d{4}-\d{2}-\d{2}$/.test(startDate)) {
      setError('Format tanggal harus YYYY-MM-DD.')
      return
    }
    act.mutate({ id, action: 'confirm', body: { start_date: `${startDate}T00:00:00Z` } })
  }

  return (
    <div className="mx-auto max-w-5xl px-4 py-6">
      <h1 className="mb-4 text-2xl font-bold">Booking Masuk</h1>
      {error && <p className="mb-3 rounded-md bg-red-50 p-2 text-sm text-red-700">{error}</p>}

      <div className="mb-4 flex flex-wrap gap-2">
        {TABS.map(([v, label]) => (
          <button
            key={v}
            onClick={() => setTab(v)}
            className={`rounded-full px-3 py-1.5 text-sm ${
              tab === v ? 'bg-teal-600 text-white' : 'border bg-white hover:bg-slate-50'
            }`}
          >
            {label}
          </button>
        ))}
      </div>

      {isLoading && <p className="text-slate-500">Memuat…</p>}
      {data && data.data.length === 0 && (
        <div className="rounded-xl border border-dashed p-10 text-center text-slate-500">
          Tidak ada booking pada status ini.
        </div>
      )}
      <div className="space-y-3">
        {data?.data.map((b) => (
          <BookingCard
            key={b.id}
            b={b}
            role="owner"
            busy={act.isPending}
            onAct={(a) => act.mutate({ id: b.id, action: a })}
            onReject={reject}
            onConfirm={confirm}
          />
        ))}
      </div>
    </div>
  )
}
