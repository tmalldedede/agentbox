import { createFileRoute } from '@tanstack/react-router'
import { TaskDetail } from '@/features/tasks/components/TaskDetail'

export const Route = createFileRoute('/_authenticated/tasks/$id')({
  component: () => {
    const { id } = Route.useParams()
    return <TaskDetail taskId={id} />
  },
})
