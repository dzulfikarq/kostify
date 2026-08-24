import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api, errMessage } from '../../lib/api'
import { StatusBadge } from '../../components/StatusBadge'
import type { BookingWithRefs } from '../../lib/types'

export function TenantBookingsPage() {
  const qc = useQueryClient()
  const [error, setError] = useState('')
  const refresh = () => qc.invalidateQueries({ queryKey: ['my-bookings'] })

  const { data, isLoading } = useQuery({
    queryKey: ['my-bookings'],
    queryFn: async () => (await api.get('/bookings/me')).data as { data: BookingWithRefs[] },
  })

  const act = useMutation({
    mutationFn: async ({ id, action }: { id: string; action: string }) => {
      await api.put(`/bookings/${id}/${action}`)
    },
    onSuccess: refresh,
    onError: (e) => setError(errMessage(e)),
  })

  if (isLoading)
    return <div className="mx-auto max-w-5xl px-4 py-6 text-slate-500">Memuat…</div>

  return (
    <div className="mx-auto max-w-5xl px-4 py-6">
      <h1 className="mb-4 text-2xl font-bold">Booking Saya</h1>
      {error && <p className="mb-3 rounded-md bg-red-50 p-2 text-sm text-red-700">{error}</p>}
      {data && data.data.length === 0 && (
        <div className="rounded-xl border border-dashed p-10 text-center text-slate-500">
          Belum ada booking. Cari kost dan ajukan booking pertamamu.
        </div>
      )}
      <div className="space-y-3">
        {data?.data.map((b) => (
          <div key={b.id} className="space-y-2">
            <BookingCard b={b} role="tenant" onAct={(a) => act.mutate({ id: b.id, action: a })} busy={act.isPending} />
            {b.status === 'completed' && <ReviewForm bookingId={b.id} />}
          </div>
        ))}
      </div>
    </div>
  )
}

function fmtDate(s?: string | null): string {
  return s ? new Date(s).toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'short' }) : '—'
}

export function BookingCard({
  b,
  role,
  onAct,
  onReject,
  onConfirm,
  busy = false,
}: {
  b: BookingWithRefs
  role: 'tenant' | 'owner'
  onAct: (action: string) => void
  onReject?: (id: string, reason: string) => void
  onConfirm?: (id: string, startDate: string) => void
  busy?: boolean
}) {
  const expiredSoon =
    b.status === 'pending' && new Date(b.expires_at).getTime() - Date.now() < 24 * 3600e3

  return (
    <div className="rounded-xl border bg-white p-4 shadow-sm">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div>
          <p className="font-semibold">
            {b.property_name} — Kamar {b.room_number}
          </p>
          <p className="text-sm text-slate-500">
            Rp {b.price_per_month.toLocaleString('id-ID')}/bln · {b.lease_duration_months} bulan
            {b.start_date ? ` · mulai ${new Date(b.start_date).toLocaleDateString('id-ID')}` : ''}
          </p>
          {b.status === 'pending' && (
            <p className={`text-xs ${expiredSoon ? 'text-red-600' : 'text-slate-400'}`}>
              Kedaluwarsa: {fmtDate(b.expires_at)}
              {expiredSoon ? ' (segera!)' : ''}
            </p>
          )}
          {b.cancel_reason && <p className="mt-1 text-xs text-red-600">Alasan: {b.cancel_reason}</p>}
        </div>
        <StatusBadge status={b.status} />
      </div>

      <div className="mt-3 flex flex-wrap items-start gap-2">
        {role === 'owner' && b.status === 'pending' && (
          <>
            <button onClick={() => onAct('approve')} disabled={busy}
              className="rounded-md bg-green-600 px-3 py-1.5 text-sm text-white hover:bg-green-700 disabled:opacity-50">
              Setujui (Survei)
            </button>
            {onReject && (
              <button onClick={() => onReject(b.id, window.prompt('Alasan penolakan (min. 10 karakter):') ?? '')}
                disabled={busy}
                className="rounded-md bg-red-600 px-3 py-1.5 text-sm text-white hover:bg-red-700 disabled:opacity-50">
                Tolak
              </button>
            )}
          </>
        )}
        {role === 'owner' && b.status === 'survey' && onConfirm && (
          <button
            onClick={() => onConfirm(b.id, window.prompt('Tanggal mulai sewa (YYYY-MM-DD):') ?? '')}
            disabled={busy}
            className="rounded-md bg-indigo-600 px-3 py-1.5 text-sm text-white hover:bg-indigo-700 disabled:opacity-50">
            Konfirmasi Booking
          </button>
        )}
        {role === 'tenant' && b.status === 'pending' && (
          <button onClick={() => onAct('cancel')} disabled={busy}
            className="rounded-md border border-red-300 px-3 py-1.5 text-sm text-red-700 hover:bg-red-50 disabled:opacity-50">
            Batalkan
          </button>
        )}
        {role === 'tenant' && b.status === 'survey' && (
          <p className="text-xs text-slate-500">Menunggu pemilik menjadwalkan survei/konfirmasi.</p>
        )}
        {role === 'tenant' && b.status === 'booked' && (
          <button onClick={() => onAct('checkin')} disabled={busy}
            className="rounded-md bg-teal-600 px-3 py-1.5 text-sm text-white hover:bg-teal-700 disabled:opacity-50">
            Check-in
          </button>
        )}
        {role === 'tenant' && b.status === 'active' && (
          <button onClick={() => onAct('checkout')} disabled={busy}
            className="rounded-md bg-slate-700 px-3 py-1.5 text-sm text-white hover:bg-slate-800 disabled:opacity-50">
            Check-out
          </button>
        )}
      </div>
    </div>
  )
}

function ReviewForm({ bookingId }: { bookingId: string }) {
  const qc = useQueryClient()
  const [open, setOpen] = useState(false)
  const [rating, setRating] = useState(5)
  const [comment, setComment] = useState('')
  const [error, setError] = useState('')
  const [done, setDone] = useState(false)

  const create = useMutation({
    mutationFn: async () => await api.post(`/bookings/${bookingId}/reviews`, { rating, comment }),
    onSuccess: () => {
      setDone(true)
      setOpen(false)
      qc.invalidateQueries({ queryKey: ['property-reviews'] })
    },
    onError: (e) => setError(errMessage(e)),
  })

  if (done) return <p className="text-xs text-green-700">Ulasan terkirim. Terima kasih!</p>
  if (!open) {
    return (
      <button onClick={() => setOpen(true)} className="rounded-md bg-purple-600 px-3 py-1.5 text-sm text-white hover:bg-purple-700">
        Beri Ulasan
      </button>
    )
  }
  return (
    <div className="w-full space-y-2 rounded-lg border bg-slate-50 p-3">
      {error && <p className="rounded-md bg-red-50 p-2 text-xs text-red-700">{error}</p>}
      <div className="flex gap-1">
        {[1, 2, 3, 4, 5].map((n) => (
          <button
            key={n}
            onClick={() => setRating(n)}
            className={`text-xl leading-none ${n <= rating ? 'text-amber-400' : 'text-slate-300'}`}
            aria-label={`${n} bintang`}
          >
            ★
          </button>
        ))}
      </div>
      <textarea
        value={comment}
        onChange={(e) => setComment(e.target.value)}
        rows={2}
        placeholder="Bagaimana pengalamanmu? (opsional)"
        className="w-full rounded-md border px-3 py-2 text-sm focus:border-teal-500 focus:outline-none"
      />
      <div className="flex justify-end gap-2">
        <button onClick={() => setOpen(false)} className="rounded-md border px-3 py-1.5 text-sm">Batal</button>
        <button onClick={() => create.mutate()} disabled={create.isPending}
          className="rounded-md bg-purple-600 px-3 py-1.5 text-sm text-white hover:bg-purple-700 disabled:opacity-50">
          Kirim
        </button>
      </div>
    </div>
  )
}
