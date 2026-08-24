import type { PropertyStatus, RoomStatus } from '../lib/types'

const propertyLabels: Record<PropertyStatus, { label: string; cls: string }> = {
  draft: { label: 'Draft', cls: 'bg-slate-100 text-slate-700' },
  pending_verification: { label: 'Menunggu Verifikasi', cls: 'bg-amber-100 text-amber-800' },
  published: { label: 'Published', cls: 'bg-green-100 text-green-800' },
  rejected: { label: 'Ditolak', cls: 'bg-red-100 text-red-800' },
  inactive: { label: 'Nonaktif', cls: 'bg-slate-100 text-slate-500' },
}

const roomLabels: Record<RoomStatus, { label: string; cls: string }> = {
  available: { label: 'Tersedia', cls: 'bg-green-100 text-green-800' },
  pending: { label: 'Menunggu Konfirmasi', cls: 'bg-amber-100 text-amber-800' },
  survey: { label: 'Survei', cls: 'bg-blue-100 text-blue-800' },
  booked: { label: 'Dibooking', cls: 'bg-indigo-100 text-indigo-800' },
  active: { label: 'Disewa', cls: 'bg-purple-100 text-purple-800' },
  maintenance: { label: 'Perbaikan', cls: 'bg-orange-100 text-orange-800' },
  completed: { label: 'Selesai', cls: 'bg-slate-100 text-slate-600' },
}

export function StatusBadge({ status }: { status: PropertyStatus | RoomStatus }) {
  const m = (propertyLabels as Record<string, { label: string; cls: string }>)[status] ??
    roomLabels[status as RoomStatus] ?? { label: status, cls: 'bg-slate-100 text-slate-700' }
  return (
    <span className={`inline-block rounded-full px-2.5 py-0.5 text-xs font-medium ${m.cls}`}>
      {m.label}
    </span>
  )
}
