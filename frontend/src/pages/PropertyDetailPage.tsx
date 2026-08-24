import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { api, errMessage } from '../lib/api'
import { StatusBadge } from '../components/StatusBadge'
import type { PropertyDetail } from '../lib/types'
import { formatIDR } from './PropertiesPage'
import { useAuth } from '../context/AuthContext'

export function PropertyDetailPage() {
  const { id } = useParams()
  const { user } = useAuth()
  const nav = useNavigate()
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
          <RoomCard key={r.id} room={r} canBook={user?.role === 'tenant'} onBook={() => nav(`/properties/${p.id}?book=${r.id}`)} />
        ))}
      </div>

      {user?.role === 'tenant' && (
        <BookingModal rooms={p.rooms} />
      )}
    </div>
  )
}

function RoomCard({ room: r, canBook, onBook }: { room: PropertyDetail['rooms'][number]; canBook: boolean; onBook: () => void }) {
  return (
    <div className="rounded-xl border bg-white p-4 shadow-sm">
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
      {canBook && (
        <button
          onClick={onBook}
          disabled={r.status !== 'available'}
          className="mt-3 w-full rounded-md bg-teal-600 py-2 text-sm font-medium text-white hover:bg-teal-700 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {r.status === 'available' ? 'Ajukan Booking' : 'Tidak Tersedia'}
        </button>
      )}
    </div>
  )
}

function BookingModal({ rooms }: { rooms: PropertyDetail['rooms'] }) {
  const [searchParams, setSearchParams] = useSearchParams()
  const bookRoom = searchParams.get('book')
  const room = rooms.find((r) => r.id === bookRoom)
  const [months, setMonths] = useState('12')
  const [note, setNote] = useState('')
  const [error, setError] = useState('')
  const qc = useQueryClient()

  const create = useMutation({
    mutationFn: async () => {
      await api.post('/bookings', {
        room_id: room!.id,
        lease_duration_months: Number(months),
        note: note || undefined,
      })
    },
    onSuccess: () => {
      setSearchParams({})
      qc.invalidateQueries()
    },
    onError: (e) => setError(errMessage(e)),
  })

  if (!room) return null
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={() => setSearchParams({})}>
      <div className="w-full max-w-md space-y-3 rounded-xl bg-white p-6 shadow-xl" onClick={(e) => e.stopPropagation()}>
        <h3 className="text-lg font-bold">Booking Kamar {room.room_number}</h3>
        {error && <p className="rounded-md bg-red-50 p-2 text-sm text-red-700">{error}</p>}
        <div>
          <label className="mb-1 block text-sm font-medium">Durasi Sewa (bulan) *</label>
          <input type="number" min={1} max={36} value={months} onChange={(e) => setMonths(e.target.value)} className="w-full rounded-md border px-3 py-2 focus:border-teal-500 focus:outline-none" />
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium">Catatan untuk Pemilik</label>
          <textarea value={note} onChange={(e) => setNote(e.target.value)} rows={3} className="w-full rounded-md border px-3 py-2 focus:border-teal-500 focus:outline-none" />
        </div>
        <p className="text-xs text-slate-400">Booking otomatis kedaluwarsa jika pemilik tidak merespons dalam 3 hari.</p>
        <div className="flex justify-end gap-2 pt-1">
          <button onClick={() => setSearchParams({})} className="rounded-md border px-4 py-2">Batal</button>
          <button
            onClick={() => create.mutate()}
            disabled={create.isPending || !Number(months)}
            className="rounded-md bg-teal-600 px-4 py-2 text-white hover:bg-teal-700 disabled:opacity-50"
          >
            {create.isPending ? 'Mengirim…' : 'Kirim Booking'}
          </button>
        </div>
      </div>
    </div>
  )
}
