import { useRef, useState, type FormEvent } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { api, errMessage } from '../../lib/api'
import { ValidationError } from '../../lib/types'
import { StatusBadge } from '../../components/StatusBadge'
import type { PropertyDetail, Room } from '../../lib/types'

const FACILITIES = ['ac', 'wifi', 'private_bathroom', 'shared_bathroom', 'desk', 'wardrobe', 'balcony']

export function PropertyFormPage() {
  const { id } = useParams()
  const isNew = !id
  const nav = useNavigate()
  const qc = useQueryClient()
  const [tab, setTab] = useState<'info' | 'foto' | 'kamar'>('info')
  const [formError, setFormError] = useState('')

  const [name, setName] = useState('')
  const [city, setCity] = useState('')
  const [address, setAddress] = useState('')
  const [description, setDescription] = useState('')
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [loadedId, setLoadedId] = useState<string | null>(null)

  const { data: p } = useQuery({
    enabled: !!id,
    queryKey: ['owner-property', id],
    queryFn: async () => (await api.get(`/properties/${id}`)).data.data as PropertyDetail,
  })
  if (p && loadedId !== p.id) {
    setLoadedId(p.id)
    setName(p.name)
    setCity(p.city)
    setAddress(p.address)
    setDescription(p.description ?? '')
  }

  function mapValidation(e: unknown): boolean {
    if (e instanceof ValidationError) {
      const fe: Record<string, string> = {}
      for (const d of e.details) fe[d.field] = d.message
      setFieldErrors(fe)
      return true
    }
    return false
  }

  const saveInfo = useMutation({
    mutationFn: async () => {
      if (isNew) {
        const r = await api.post('/properties', {
          name,
          city,
          address,
          description: description || undefined,
        })
        return r.data.data.id as string
      }
      await api.put(`/properties/${id}`, {
        name,
        city,
        address,
        description: description || undefined,
      })
      return id!
    },
    onSuccess: async (savedId) => {
      await qc.invalidateQueries({ queryKey: ['owner-properties'] })
      if (isNew) nav(`/owner/properties/${savedId}/edit`, { replace: true })
    },
    onError: (e) => {
      if (!mapValidation(e)) setFormError(errMessage(e))
    },
  })

  async function onSubmitInfo(e: FormEvent) {
    e.preventDefault()
    setFieldErrors({})
    setFormError('')
    saveInfo.mutate()
  }

  const inputCls =
    'w-full rounded-md border px-3 py-2 focus:border-teal-500 focus:outline-none disabled:bg-slate-100'
  const errOf = (f: string) =>
    fieldErrors[f] ? <p className="mt-1 text-xs text-red-600">{fieldErrors[f]}</p> : null

  return (
    <div className="mx-auto max-w-3xl px-4 py-6">
      <Link to="/owner/properties" className="text-sm text-teal-700 hover:underline">
        ← Kost Saya
      </Link>
      <div className="mt-2 flex items-center gap-3">
        <h1 className="text-2xl font-bold">{isNew ? 'Tambah Kost' : 'Kelola Kost'}</h1>
        {p && <StatusBadge status={p.status} />}
      </div>
      {p?.rejection_reason && (
        <p className="mt-2 rounded-md bg-red-50 p-3 text-sm text-red-700">
          Ditolak: {p.rejection_reason}
        </p>
      )}

      {!isNew && (
        <div className="mt-4 flex gap-2 border-b">
          {(
            [
              ['info', 'Info'],
              ['foto', 'Foto'],
              ['kamar', 'Kamar'],
            ] as const
          ).map(([v, label]) => (
            <button
              key={v}
              onClick={() => setTab(v)}
              className={`px-4 py-2 text-sm font-medium ${
                tab === v ? 'border-b-2 border-teal-600 text-teal-700' : 'text-slate-500'
              }`}
            >
              {label}
            </button>
          ))}
        </div>
      )}

      {(tab === 'info' || isNew) && (
        <form onSubmit={onSubmitInfo} className="mt-4 space-y-4 rounded-xl border bg-white p-6 shadow-sm">
          {formError && <p className="rounded-md bg-red-50 p-2 text-sm text-red-700">{formError}</p>}
          <div>
            <label className="mb-1 block text-sm font-medium">Nama Kost *</label>
            <input value={name} onChange={(e) => setName(e.target.value)} className={inputCls} required maxLength={150} />
            {errOf('name')}
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Kota *</label>
            <input value={city} onChange={(e) => setCity(e.target.value)} className={inputCls} required maxLength={100} />
            {errOf('city')}
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Alamat Lengkap *</label>
            <textarea value={address} onChange={(e) => setAddress(e.target.value)} className={inputCls} required maxLength={500} rows={2} />
            {errOf('address')}
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium">Deskripsi</label>
            <textarea value={description} onChange={(e) => setDescription(e.target.value)} className={inputCls} rows={4} />
            {errOf('description')}
          </div>
          <button
            disabled={saveInfo.isPending}
            className="rounded-md bg-teal-600 px-5 py-2 text-white hover:bg-teal-700 disabled:opacity-50"
          >
            {saveInfo.isPending ? 'Menyimpan…' : isNew ? 'Simpan Draft & Lanjut' : 'Simpan Perubahan'}
          </button>
        </form>
      )}

      {!isNew && tab === 'foto' && <PhotoTab property={p} onRefresh={() => qc.invalidateQueries({ queryKey: ['owner-property', id] })} />}
      {!isNew && tab === 'kamar' && <RoomTab property={p} onRefresh={() => qc.invalidateQueries({ queryKey: ['owner-property', id] })} />}

      {!isNew && (
        <SubmitVerification
          property={p}
          onDone={() => qc.invalidateQueries({ queryKey: ['owner-property', id] })}
        />
      )}
    </div>
  )
}

function PhotoTab({ property, onRefresh }: { property?: PropertyDetail; onRefresh: () => void }) {
  const fileRef = useRef<HTMLInputElement>(null)
  const [error, setError] = useState('')
  if (!property) return null

  async function upload() {
    setError('')
    const f = fileRef.current?.files?.[0]
    if (!f) return
    const fd = new FormData()
    fd.append('file', f)
    try {
      await api.post(`/properties/${property!.id}/photos`, fd)
      if (fileRef.current) fileRef.current.value = ''
      onRefresh()
    } catch (e) {
      setError(errMessage(e, 'Gagal mengunggah foto.'))
    }
  }

  async function remove(photoId: string) {
    setError('')
    try {
      await api.delete(`/properties/${property!.id}/photos/${photoId}`)
      onRefresh()
    } catch (e) {
      setError(errMessage(e, 'Gagal menghapus foto.'))
    }
  }

  return (
    <div className="mt-4 space-y-4 rounded-xl border bg-white p-6 shadow-sm">
      {error && <p className="rounded-md bg-red-50 p-2 text-sm text-red-700">{error}</p>}
      <div>
        <input ref={fileRef} type="file" accept="image/jpeg,image/png,image/webp" />
        <button onClick={upload} className="ml-2 rounded-md bg-teal-600 px-3 py-1.5 text-sm text-white hover:bg-teal-700">
          Unggah
        </button>
        <p className="mt-1 text-xs text-slate-400">JPG/PNG/WebP maks 5MB. Minimal 1 foto untuk verifikasi.</p>
      </div>
      {property.photos.length > 0 && (
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
          {property.photos.map((ph) => (
            <div key={ph.id} className="relative overflow-hidden rounded-lg border">
              <img src={ph.url} alt="" className="h-32 w-full object-cover" />
              {ph.is_primary && (
                <span className="absolute top-1 left-1 rounded bg-teal-600 px-1.5 py-0.5 text-[10px] text-white">
                  Cover
                </span>
              )}
              <button
                onClick={() => remove(ph.id)}
                className="absolute top-1 right-1 rounded bg-red-600 px-1.5 py-0.5 text-[10px] text-white"
              >
                Hapus
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

function RoomTab({ property, onRefresh }: { property?: PropertyDetail; onRefresh: () => void }) {
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<Room | null>(null)
  const [number, setNumber] = useState('')
  const [price, setPrice] = useState('')
  const [area, setArea] = useState('')
  const [desc, setDesc] = useState('')
  const [facilities, setFacilities] = useState<string[]>([])
  const [error, setError] = useState('')
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  if (!property) return null

  function openNew() {
    setEditing(null)
    setNumber('')
    setPrice('')
    setArea('')
    setDesc('')
    setFacilities([])
    setError('')
    setFieldErrors({})
    setOpen(true)
  }

  function openEdit(r: Room) {
    setEditing(r)
    setNumber(r.room_number)
    setPrice(String(r.price_per_month))
    setArea(r.area_m2 ? String(r.area_m2) : '')
    setDesc(r.description ?? '')
    setFacilities([...r.facilities])
    setError('')
    setFieldErrors({})
    setOpen(true)
  }

  function toggleFacility(f: string) {
    setFacilities((prev) => (prev.includes(f) ? prev.filter((x) => x !== f) : [...prev, f]))
  }

  async function save() {
    setError('')
    setFieldErrors({})
    const body: Record<string, unknown> = {
      room_number: number,
      price_per_month: Number(price),
      facilities,
    }
    if (area) body.area_m2 = Number(area)
    if (desc) body.description = desc
    try {
      if (editing) await api.put(`/rooms/${editing.id}`, body)
      else await api.post(`/properties/${property!.id}/rooms`, body)
      setOpen(false)
      onRefresh()
    } catch (e) {
      if (e instanceof ValidationError) {
        const fe: Record<string, string> = {}
        for (const d of e.details) fe[d.field] = d.message
        setFieldErrors(fe)
      } else setError(errMessage(e))
    }
  }

  async function remove(room: Room) {
    setError('')
    try {
      await api.delete(`/rooms/${room.id}`)
      onRefresh()
    } catch (e) {
      setError(errMessage(e))
    }
  }

  async function toggleMaintenance(room: Room) {
    setError('')
    const to = room.status === 'available' ? 'maintenance' : 'available'
    try {
      await api.patch(`/rooms/${room.id}/status`, { status: to })
      onRefresh()
    } catch (e) {
      setError(errMessage(e))
    }
  }

  const inputCls =
    'w-full rounded-md border px-3 py-2 focus:border-teal-500 focus:outline-none'
  const errOf = (f: string) =>
    fieldErrors[f] ? <p className="mt-1 text-xs text-red-600">{fieldErrors[f]}</p> : null

  return (
    <div className="mt-4 space-y-4">
      {error && <p className="rounded-md bg-red-50 p-2 text-sm text-red-700">{error}</p>}
      <button onClick={openNew} className="rounded-md bg-teal-600 px-4 py-2 text-sm text-white hover:bg-teal-700">
        + Tambah Kamar
      </button>

      {property.rooms.length === 0 && (
        <p className="rounded-xl border border-dashed p-8 text-center text-slate-500">
          Belum ada kamar. Minimal 1 kamar untuk verifikasi.
        </p>
      )}

      <div className="overflow-hidden rounded-xl border bg-white shadow-sm">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 text-left text-xs uppercase text-slate-500">
            <tr>
              <th className="px-4 py-2">No</th>
              <th className="px-4 py-2">Harga</th>
              <th className="px-4 py-2">Status</th>
              <th className="px-4 py-2 text-right">Aksi</th>
            </tr>
          </thead>
          <tbody>
            {property.rooms.map((r) => (
              <tr key={r.id} className="border-t">
                <td className="px-4 py-2 font-medium">{r.room_number}</td>
                <td className="px-4 py-2">Rp {r.price_per_month.toLocaleString('id-ID')}</td>
                <td className="px-4 py-2"><StatusBadge status={r.status} /></td>
                <td className="space-x-2 px-4 py-2 text-right whitespace-nowrap">
                  <button onClick={() => openEdit(r)} className="text-teal-700 hover:underline">Edit</button>
                  {(r.status === 'available' || r.status === 'maintenance') && (
                    <button onClick={() => toggleMaintenance(r)} className="text-orange-600 hover:underline">
                      {r.status === 'available' ? 'Perbaikan' : 'Aktifkan'}
                    </button>
                  )}
                  <button onClick={() => remove(r)} className="text-red-600 hover:underline">Hapus</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {open && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={() => setOpen(false)}>
          <div className="max-h-[90vh] w-full max-w-lg space-y-3 overflow-y-auto rounded-xl bg-white p-6 shadow-xl" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-lg font-bold">{editing ? `Edit Kamar ${editing.room_number}` : 'Kamar Baru'}</h3>
            <div>
              <label className="mb-1 block text-sm font-medium">Nomor Kamar *</label>
              <input value={number} onChange={(e) => setNumber(e.target.value)} className={inputCls} maxLength={20} />
              {errOf('room_number')}
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium">Harga per Bulan (Rp) *</label>
              <input type="number" min={1} value={price} onChange={(e) => setPrice(e.target.value)} className={inputCls} />
              {errOf('price_per_month')}
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium">Luas (m²)</label>
              <input type="number" min={1} value={area} onChange={(e) => setArea(e.target.value)} className={inputCls} />
              {errOf('area_m2')}
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium">Deskripsi</label>
              <textarea value={desc} onChange={(e) => setDesc(e.target.value)} className={inputCls} rows={2} />
              {errOf('description')}
            </div>
            <fieldset>
              <legend className="mb-1 text-sm font-medium">Fasilitas</legend>
              <div className="flex flex-wrap gap-x-4 gap-y-1">
                {FACILITIES.map((f) => (
                  <label key={f} className="flex items-center gap-1.5 text-sm capitalize">
                    <input type="checkbox" checked={facilities.includes(f)} onChange={() => toggleFacility(f)} />
                    {f.replaceAll('_', ' ')}
                  </label>
                ))}
              </div>
              {errOf('facilities')}
            </fieldset>
            <div className="flex justify-end gap-2 pt-2">
              <button onClick={() => setOpen(false)} className="rounded-md border px-4 py-2">Batal</button>
              <button onClick={save} className="rounded-md bg-teal-600 px-4 py-2 text-white hover:bg-teal-700">Simpan</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function SubmitVerification({ property, onDone }: { property?: PropertyDetail; onDone: () => void }) {
  const [error, setError] = useState('')
  const [msg, setMsg] = useState('')
  if (!property) return null

  const eligible =
    property.status === 'draft' ||
    property.status === 'rejected' ||
    (property.status === 'published' && property.rooms.some((r) => r.room_number)) // published bisa re-edit lalu resubmit? tidak — hanya draft/rejected
  const canSubmit = property.status === 'draft' || property.status === 'rejected'
  void eligible
  const ready = property.photos.length > 0 && property.rooms.length > 0

  async function submit() {
    setError('')
    setMsg('')
    try {
      await api.post(`/properties/${property!.id}/submit`)
      setMsg('Kost diajukan untuk verifikasi.')
      onDone()
    } catch (e) {
      setError(errMessage(e))
    }
  }

  if (!canSubmit) return null
  return (
    <div className="mt-6 rounded-xl border bg-white p-6 shadow-sm">
      <h3 className="font-semibold">Verifikasi</h3>
      <p className="mt-1 text-sm text-slate-500">
        Ajukan kost untuk diverifikasi Super Admin sebelum tayang publik. Syarat: minimal 1 foto dan 1 kamar.
      </p>
      {!ready && <p className="mt-2 text-sm text-amber-700">Lengkapi minimal 1 foto dan 1 kamar terlebih dahulu.</p>}
      {error && <p className="mt-2 rounded-md bg-red-50 p-2 text-sm text-red-700">{error}</p>}
      {msg && <p className="mt-2 rounded-md bg-green-50 p-2 text-sm text-green-700">{msg}</p>}
      <button
        onClick={submit}
        disabled={!ready}
        className="mt-3 rounded-md bg-teal-600 px-5 py-2 text-sm text-white hover:bg-teal-700 disabled:opacity-50"
      >
        {property.status === 'rejected' ? 'Ajukan Ulang Verifikasi' : 'Ajukan Verifikasi'}
      </button>
    </div>
  )
}
