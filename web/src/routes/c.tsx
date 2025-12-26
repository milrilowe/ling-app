import { createFileRoute, Outlet } from '@tanstack/react-router'
import { requireAuth } from '@/lib/auth-guard'

export const Route = createFileRoute('/c')({
  beforeLoad: requireAuth,
  component: Outlet,
})
