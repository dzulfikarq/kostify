export type Role = 'tenant' | 'owner' | 'super_admin'

export interface User {
  id: string
  name: string
  email: string
  role: Role
}

export type PropertyStatus =
  | 'draft'
  | 'pending_verification'
  | 'published'
  | 'rejected'
  | 'inactive'

export type RoomStatus =
  | 'available'
  | 'pending'
  | 'survey'
  | 'booked'
  | 'active'
  | 'maintenance'
  | 'completed'

export interface PropertySummary {
  id: string
  name: string
  city: string
  status: PropertyStatus
  rating_avg: number
  rating_count: number
  photo_url: string | null
  starting_price: number | null
  available_rooms: number
  created_at: string
}

export interface Photo {
  id: string
  url: string
  is_primary: boolean
  sort_order: number
}

export interface Room {
  id?: string
  room_number: string
  price_per_month: number
  area_m2?: number | null
  description?: string | null
  facilities: string[]
  status: RoomStatus
}

export interface PropertyDetail {
  id: string
  owner_id: string
  name: string
  description: string | null
  address: string
  city: string
  status: PropertyStatus
  rejection_reason: string | null
  reviews_summary: { avg: number; count: number }
  photos: Photo[]
  rooms: Room[]
  created_at: string
}

export interface OwnerProperty {
  id: string
  name: string
  city: string
  status: PropertyStatus
  rejection_reason: string | null
  rating_avg: number
  rating_count: number
  created_at: string
}

export interface ApiMeta {
  page: number
  limit: number
  total: number
  total_pages: number
}

export interface BookingWithRefs {
  id: string
  room_id: string
  property_id: string
  property_name: string
  room_number: string
  tenant_id: string
  owner_id: string
  status: import('./types').BookingStatusAlias
  price_per_month: number
  lease_duration_months: number
  start_date: string | null
  expires_at: string
  checked_in_at: string | null
  checked_out_at: string | null
  cancel_reason: string | null
  created_at: string
}

export type BookingStatusAlias = 'pending' | 'survey' | 'booked' | 'active' | 'completed' | 'cancelled' | 'rejected' | 'expired'

export class ValidationError extends Error {
  details: { field: string; message: string }[]
  constructor(details: { field: string; message: string }[]) {
    super('validation')
    this.details = details
  }
}
