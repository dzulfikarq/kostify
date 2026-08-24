import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { AuthProvider } from './context/AuthContext'
import { Layout } from './components/Layout'
import { RequireAuth } from './components/RequireAuth'
import { HomePage } from './pages/HomePage'
import { PropertiesPage } from './pages/PropertiesPage'
import { PropertyDetailPage } from './pages/PropertyDetailPage'
import { LoginPage } from './pages/LoginPage'
import { RegisterPage } from './pages/RegisterPage'
import { OwnerPropertiesPage } from './pages/owner/OwnerPropertiesPage'
import { PropertyFormPage } from './pages/owner/PropertyFormPage'
import { VerificationQueuePage } from './pages/admin/VerificationQueuePage'

const qc = new QueryClient({
  defaultOptions: { queries: { retry: 1, refetchOnWindowFocus: false } },
})

export default function App() {
  return (
    <QueryClientProvider client={qc}>
      <AuthProvider>
        <BrowserRouter>
          <Routes>
            <Route element={<Layout />}>
              <Route path="/" element={<HomePage />} />
              <Route path="/properties" element={<PropertiesPage />} />
              <Route path="/properties/:id" element={<PropertyDetailPage />} />
              <Route path="/login" element={<LoginPage />} />
              <Route path="/register" element={<RegisterPage />} />

              <Route
                path="/owner/properties"
                element={
                  <RequireAuth roles={['owner']}>
                    <OwnerPropertiesPage />
                  </RequireAuth>
                }
              />
              <Route
                path="/owner/properties/new"
                element={
                  <RequireAuth roles={['owner']}>
                    <PropertyFormPage />
                  </RequireAuth>
                }
              />
              <Route
                path="/owner/properties/:id/edit"
                element={
                  <RequireAuth roles={['owner']}>
                    <PropertyFormPage />
                  </RequireAuth>
                }
              />

              <Route
                path="/admin/verifications"
                element={
                  <RequireAuth roles={['super_admin']}>
                    <VerificationQueuePage />
                  </RequireAuth>
                }
              />

              <Route path="/dashboard" element={<Navigate to="/" replace />} />
              <Route
                path="*"
                element={
                  <div className="py-20 text-center">
                    <p className="text-4xl font-bold text-slate-700">404</p>
                    <p className="mt-2 text-slate-500">Halaman tidak ditemukan.</p>
                  </div>
                }
              />
            </Route>
          </Routes>
        </BrowserRouter>
      </AuthProvider>
    </QueryClientProvider>
  )
}
