import { Skeleton } from '@/components/ui/skeleton'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
  SidebarRail,
  SidebarTrigger,
} from '@/components/ui/sidebar'
import { ThreadListItem } from '@/features/sidebar/components/ThreadListItem'
import { useThreads } from '@/hooks/use-thread'
import { useNavigate } from '@tanstack/react-router'
import { Sparkles, PenSquare, Search, Settings } from 'lucide-react'

export function AppSidebar() {
  const navigate = useNavigate()
  const { data: threads, isLoading } = useThreads()

  const handleNewChat = () => {
    navigate({ to: '/' })
  }

  return (
    <Sidebar collapsible="icon">
      <SidebarHeader>
        <div className="group/header flex items-center gap-2 p-2">
          <div className="flex items-center gap-2 min-w-0 flex-1">
            <Sparkles className="h-5 w-5 shrink-0" />
            <span className="font-semibold text-sm truncate group-data-[collapsible=icon]:hidden">
              Ling App
            </span>
          </div>
          <div className="group-data-[collapsible=icon]:absolute group-data-[collapsible=icon]:right-2 group-data-[collapsible=icon]:opacity-0 group-data-[collapsible=icon]:group-hover/header:opacity-100 transition-opacity">
            <SidebarTrigger className="shrink-0" />
          </div>
        </div>
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            {/* Collapsed view - just compose and search icons */}
            <div className="hidden group-data-[collapsible=icon]:flex flex-col gap-1">
              <SidebarMenu>
                <SidebarMenuItem>
                  <SidebarMenuButton
                    onClick={handleNewChat}
                    tooltip="New Conversation"
                  >
                    <PenSquare className="h-4 w-4" />
                  </SidebarMenuButton>
                </SidebarMenuItem>
                <SidebarMenuItem>
                  <SidebarMenuButton
                    tooltip="Search"
                  >
                    <Search className="h-4 w-4" />
                  </SidebarMenuButton>
                </SidebarMenuItem>
              </SidebarMenu>
            </div>

            {/* Expanded view - show all threads */}
            <div className="group-data-[collapsible=icon]:hidden">
              <SidebarMenu>
                <SidebarMenuItem>
                  <SidebarMenuButton
                    onClick={handleNewChat}
                    className="mb-2"
                  >
                    <PenSquare className="h-4 w-4" />
                    <span>New Conversation</span>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              </SidebarMenu>

              {isLoading ? (
                <div className="space-y-2 px-2">
                  {[...Array(5)].map((_, i) => (
                    <Skeleton key={i} className="h-10 w-full" />
                  ))}
                </div>
              ) : threads && threads.length > 0 ? (
                <SidebarMenu>
                  {threads.map((thread) => (
                    <ThreadListItem key={thread.id} thread={thread} />
                  ))}
                </SidebarMenu>
              ) : (
                <div className="px-3 py-8 text-center text-sm text-muted-foreground">
                  No conversations yet. Start a new one!
                </div>
              )}
            </div>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton tooltip="Settings">
              <Settings className="h-4 w-4" />
              <span>Settings</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>

      <SidebarRail />
    </Sidebar>
  )
}
