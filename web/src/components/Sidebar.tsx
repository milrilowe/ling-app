import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
  SidebarTrigger,
} from '@/components/ui/sidebar'
import { Skeleton } from '@/components/ui/skeleton'
import { ThreadListItem } from '@/features/sidebar/components/ThreadListItem'
import { useArchivedThreads, useThreads } from '@/hooks/use-thread'
import { cn } from '@/lib/utils'
import { Link, useNavigate } from '@tanstack/react-router'
import {
  Archive,
  BarChart3,
  ChevronDown,
  Milk,
  PenSquare,
  Search,
  Settings,
} from 'lucide-react'
import { useState } from 'react'

export function AppSidebar() {
  const navigate = useNavigate()
  const { data: threads, isLoading } = useThreads()
  const { data: archivedThreads } = useArchivedThreads()
  const [showArchived, setShowArchived] = useState(false)

  const handleNewChat = () => {
    navigate({ to: '/' })
  }

  return (
    <Sidebar collapsible="icon">
      <SidebarHeader>
        <div className="group/header flex items-center gap-2 p-2">
          <div className="flex items-center gap-2 min-w-0 flex-1">
            <Milk className="h-5 w-5 shrink-0" />
            <span className="font-semibold text-sm truncate group-data-[collapsible=icon]:hidden">
              UtterLabs
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
                  <SidebarMenuButton tooltip="Search">
                    <Search className="h-4 w-4" />
                  </SidebarMenuButton>
                </SidebarMenuItem>
              </SidebarMenu>
            </div>

            {/* Expanded view - show all threads */}
            <div className="group-data-[collapsible=icon]:hidden">
              <SidebarMenu>
                <SidebarMenuItem>
                  <SidebarMenuButton onClick={handleNewChat} className="mb-2">
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

              {/* Archived threads section */}
              {archivedThreads && archivedThreads.length > 0 && (
                <Collapsible
                  open={showArchived}
                  onOpenChange={setShowArchived}
                  className="mt-4"
                >
                  <CollapsibleTrigger className="flex w-full items-center gap-2 px-3 py-2 text-sm text-muted-foreground hover:text-foreground">
                    <Archive className="h-4 w-4" />
                    <span>Archived</span>
                    <ChevronDown
                      className={cn(
                        'ml-auto h-4 w-4 transition-transform',
                        showArchived && 'rotate-180',
                      )}
                    />
                  </CollapsibleTrigger>
                  <CollapsibleContent>
                    <SidebarMenu>
                      {archivedThreads.map((thread) => (
                        <ThreadListItem key={thread.id} thread={thread} />
                      ))}
                    </SidebarMenu>
                  </CollapsibleContent>
                </Collapsible>
              )}
            </div>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <Link to="/pronunciation" className="w-full">
              <SidebarMenuButton tooltip="Pronunciation Stats">
                <BarChart3 className="h-4 w-4" />
                <span>Pronunciation</span>
              </SidebarMenuButton>
            </Link>
          </SidebarMenuItem>
          <SidebarMenuItem>
            <Link to="/settings" className="w-full">
              <SidebarMenuButton tooltip="Settings">
                <Settings className="h-4 w-4" />
                <span>Settings</span>
              </SidebarMenuButton>
            </Link>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>

      <SidebarRail />
    </Sidebar>
  )
}
