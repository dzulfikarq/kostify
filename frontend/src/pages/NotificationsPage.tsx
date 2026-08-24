import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api, errMessage } from '../lib/api'

interface Notification {
  id: string
  title: string
  body: string
  type: string
  is_read: boolean
  created_at: string
}

export function NotificationsPage() {
  const qc = useQueryClient()
  const [error, setError] = useState('')
  const { data, isLoading } = useQuery({
    queryKey: ['notifications'],
    queryFn: async () =>
      (await api.get('/notifications?limit=50')).data as {
        data: Notification[]
        meta: { total: number }
      },
    refetchInterval: 15_000,
  })

  const markAll = useMutation({
    mutationFn: async () => await api.put('/notifications/read-all'),
    onSuccess: () => {
      setError('')
      qc.invalidateQueries({ queryKey: ['notifications'] })
      qc.invalidateQueries({ queryKey: ['notif-unread'] })
    },
    onError: (e) => setError(errMessage(e)),
  })

  if (isLoading) return <div className="mx-auto max-w-3xl px-4 py-6 text-slate-500">Memuat…</div>

  return (
    <div className="mx-auto max-w-3xl px-4 py-6">
      <div className="mb-4 flex items-center justify-between">
        <h1 className="text-2xl font-bold">Notifikasi</h1>
        <button
          onClick={() => markAll.mutate()}
          disabled={markAll.isPending}
          className="rounded-md border px-3 py-1.5 text-sm hover:bg-slate-50 disabled:opacity-50"
        >
          Tandai semua dibaca
        </button>
      </div>
      {error && <p className="mb-3 rounded-md bg-red-50 p-2 text-sm text-red-700">{error}</p>}
      {data && data.data.length === 0 && (
        <div className="rounded-xl border border-dashed p-10 text-center text-slate-500">
          Belum ada notifikasi.
        </div>
      )}
      <div className="space-y-2">
        {data?.data.map((n) => (
          <div
            key={n.id}
            className={`rounded-xl border p-4 ${n.is_read ? 'bg-white' : 'border-teal-200 bg-teal-50'}`}
          >
            <div className="flex items-center justify-between">
              <p className="font-semibold">{n.title}</p>
              {!n.is_read && <span className="h-2 w-2 rounded-full bg-teal-600" />}
            </div>
            <p className="mt-0.5 text-sm text-slate-600">{n.body}</p>
            <p className="mt-1 text-xs text-slate-400">
              {new Date(n.created_at).toLocaleString('id-ID')}
            </p>
          </div>
        ))}
      </div>
    </div>
  )
}

export function NotificationBell() {
  const { data } = useQuery({
    queryKey: ['notif-unread'],
    queryFn: async () =>
      (await api.get('/notifications?is_read=false&limit=1')).data as { meta: { total: number } },
    refetchInterval: 15_000,
  })
  const count = data?.meta.total ?? 0
  return (
    <a href="/notifications" className="relative hover:text-teal-700">
      Notifikasi
      {count > 0 && (
        <span className="absolute -top-2 -right-4 rounded-full bg-red-600 px-1.5 text-xs text-white">
          {count > 99 ? '99+' : count}
        </span>
      )}
    </a>
  )
}
