import { Button } from '@/components/ui/button'
import { Form, FormControl, FormField, FormItem } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { useSendMessage, useThread } from '@/hooks/use-thread'
import { useParams } from '@tanstack/react-router'
import { useForm } from 'react-hook-form'

interface FormData {
  message: string
}

export function ChatThread() {
  const { threadId } = useParams({ from: '/c/$threadId' })

  const { data: thread, isLoading, isError } = useThread(threadId)
  const sendMessageMutation = useSendMessage(threadId)

  const form = useForm<FormData>({
    defaultValues: {
      message: '',
    },
  })

  const onSubmit = (data: FormData) => {
    sendMessageMutation.mutate(data.message, {
      onSuccess: () => {
        form.reset()
      },
    })
  }

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <p className="text-muted-foreground">Loading conversation...</p>
      </div>
    )
  }

  if (isError) {
    return (
      <div className="flex h-screen items-center justify-center">
        <p className="text-destructive">Failed to load conversation.</p>
      </div>
    )
  }

  return (
    <div className="flex h-screen flex-col">
      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4">
        {thread?.messages.map((message) => (
          <div
            key={message.id}
            className={`mb-4 ${
              message.role === 'user' ? 'text-right' : 'text-left'
            }`}
          >
            <div
              className={`inline-block rounded-lg px-4 py-2 ${
                message.role === 'user'
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-muted'
              }`}
            >
              {message.content}
            </div>
          </div>
        ))}
      </div>

      {/* Input */}
      <div className="border-t p-4">
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="flex gap-2">
            <FormField
              control={form.control}
              name="message"
              render={({ field }) => (
                <FormItem className="flex-1">
                  <FormControl>
                    <Input
                      placeholder="Type your message..."
                      disabled={sendMessageMutation.isPending}
                      {...field}
                    />
                  </FormControl>
                </FormItem>
              )}
            />
            <Button type="submit" disabled={sendMessageMutation.isPending}>
              Send
            </Button>
          </form>
        </Form>
      </div>
    </div>
  )
}
