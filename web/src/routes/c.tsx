import { AppLayout } from '@/components/layouts/AppLayout'
import { createFileRoute, Outlet } from '@tanstack/react-router'

function ChatLayout() {
  return (
    <AppLayout>
      <Outlet />
    </AppLayout>
  )
}

export const Route = createFileRoute('/c')({
  component: ChatLayout,
})
