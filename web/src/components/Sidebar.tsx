import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import { Sheet, SheetContent, SheetTrigger } from '@/components/ui/sheet'
import { Skeleton } from '@/components/ui/skeleton'
import { ThreadListItem } from '@/features/sidebar/components/ThreadListItem'
import { useThreads } from '@/hooks/use-thread'
import { useNavigate } from '@tanstack/react-router'
import { Menu, Plus, Settings } from 'lucide-react'

function SidebarContent() {
  const navigate = useNavigate()
  const { data: threads, isLoading } = useThreads()

  const handleNewChat = () => {
    navigate({ to: '/' })
  }

  return (
    <div className="flex h-full flex-col">
      {/* Header */}
      <div className="p-4">
        <Button onClick={handleNewChat} className="w-full gap-2">
          <Plus className="h-4 w-4" />
          New Conversation
        </Button>
      </div>

      <Separator />
      {/* Thread List */}
      <ScrollArea className="flex-1 px-2 py-2">
        {isLoading ? (
          <div className="space-y-2 px-2">
            {[...Array(5)].map((_, i) => (
              <Skeleton key={i} className="h-14 w-full" />
            ))}
          </div>
        ) : threads && threads.length > 0 ? (
          <div className="space-y-1">
            {threads.map((thread) => (
              <ThreadListItem key={thread.id} thread={thread} />
            ))}
          </div>
        ) : (
          <div className="px-3 py-8 text-center text-sm text-muted-foreground">
            No conversations yet. Start a new one!
          </div>
        )}
      </ScrollArea>

      {/* Footer */}
      <div className="border-t p-4">
        <Button variant="ghost" className="w-full justify-start gap-2">
          <Settings className="h-4 w-4" />
          Settings
        </Button>
      </div>
    </div>
  )
}

export function Sidebar() {
  return (
    <>
      {/* Desktop Sidebar */}
      <aside className="hidden h-screen w-64 shrink-0 border-r bg-sidebar md:flex">
        <SidebarContent />
      </aside>

      {/* Mobile Sidebar */}
      <div className="md:hidden">
        <Sheet>
          <SheetTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              className="fixed left-4 top-4 z-50"
            >
              <Menu className="h-5 w-5" />
            </Button>
          </SheetTrigger>
          <SheetContent side="left" className="w-64 p-0">
            <SidebarContent />
          </SheetContent>
        </Sheet>
      </div>
    </>
  )
}
