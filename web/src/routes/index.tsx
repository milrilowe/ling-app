import { Home } from '@/features/home'
import { createFileRoute } from '@tanstack/react-router'
import { requireAuth } from '@/lib/auth-guard'

export const Route = createFileRoute('/')({
  beforeLoad: requireAuth,
  component: Home,
})
