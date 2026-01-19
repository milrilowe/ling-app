import { Outlet, createRootRouteWithContext, useLocation } from '@tanstack/react-router'
import { TanStackRouterDevtoolsPanel } from '@tanstack/react-router-devtools'
import { TanStackDevtools } from '@tanstack/react-devtools'
import { Toaster } from 'sonner'

import TanStackQueryDevtools from '@/integrations/tanstack-query/devtools'
import { AuthProvider } from '@/contexts/AuthContext'
import { AudioPlayerProvider } from '@/contexts/AudioPlayerContext'
import { AppSidebar } from '@/components/Sidebar'
import { SidebarProvider, SidebarInset } from '@/components/ui/sidebar'

import type { QueryClient } from '@tanstack/react-query'

interface MyRouterContext {
  queryClient: QueryClient
}

export const Route = createRootRouteWithContext<MyRouterContext>()({
  component: RootComponent,
})

function RootComponent() {
  const location = useLocation()
  const isAuthPage = location.pathname === '/login' || location.pathname === '/register'

  return (
    <AuthProvider>
      <AudioPlayerProvider>
        {isAuthPage ? (
          <main className="flex-1 overflow-auto">
            <Outlet />
          </main>
        ) : (
          <SidebarProvider defaultOpen={false}>
            <AppSidebar />
            <SidebarInset>
              <main className="flex-1 overflow-auto">
                <Outlet />
              </main>
            </SidebarInset>
          </SidebarProvider>
        )}
        <Toaster position="top-center" />
        <TanStackDevtools
          config={{
            position: 'bottom-right',
          }}
          plugins={[
            {
              name: 'Tanstack Router',
              render: <TanStackRouterDevtoolsPanel />,
            },
            TanStackQueryDevtools,
          ]}
        />
      </AudioPlayerProvider>
    </AuthProvider>
  )
}
