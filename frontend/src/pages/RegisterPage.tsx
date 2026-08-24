import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { errMessage } from '../lib/api'
import type { Role } from '../lib/types'

export function RegisterPage() {
  const { register, login } = useAuth()
  const nav = useNavigate()
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [role, setRole] = useState<Role>('tenant')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    setBusy(true)
    try {
      await register(name, email, password, role)
      await login(email, password)
      nav('/')
    } catch (err) {
      setError(errMessage(err))
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="mx-auto max-w-md px-4 py-16">
      <h1 className="mb-6 text-center text-2xl font-bold">Daftar Akun Baru</h1>
      <form onSubmit={onSubmit} className="space-y-4 rounded-xl border bg-white p-6 shadow-sm">
        {error && <p className="rounded-md bg-red-50 p-2 text-sm text-red-700">{error}</p>}
        <div>
          <label className="mb-1 block text-sm font-medium">Nama Lengkap</label>
          <input
            required
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full rounded-md border px-3 py-2 focus:border-teal-500 focus:outline-none"
          />
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium">Email</label>
          <input
            type="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full rounded-md border px-3 py-2 focus:border-teal-500 focus:outline-none"
          />
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium">Kata Sandi</label>
          <input
            type="password"
            required
            minLength={8}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full rounded-md border px-3 py-2 focus:border-teal-500 focus:outline-none"
          />
          <p className="mt-1 text-xs text-slate-400">Minimal 8 karakter, mengandung huruf dan angka.</p>
        </div>
        <fieldset className="flex gap-4">
          <legend className="mb-1 block w-full text-sm font-medium">Daftar sebagai</legend>
          {(
            [
              ['tenant', 'Penyewa'],
              ['owner', 'Pemilik Kost'],
            ] as [Role, string][]
          ).map(([v, label]) => (
            <label key={v} className="flex items-center gap-2 text-sm">
              <input type="radio" name="role" checked={role === v} onChange={() => setRole(v)} />
              {label}
            </label>
          ))}
        </fieldset>
        <button
          disabled={busy}
          className="w-full rounded-md bg-teal-600 py-2 text-white hover:bg-teal-700 disabled:opacity-50"
        >
          {busy ? 'Memproses…' : 'Daftar'}
        </button>
        <p className="text-center text-sm text-slate-500">
          Sudah punya akun? <Link to="/login" className="text-teal-700 underline">Masuk</Link>
        </p>
      </form>
    </div>
  )
}
