import { createFileRoute } from '@tanstack/react-router'
import { TopicPage } from '@/pages/topic/ui/topic-page'

export const Route = createFileRoute('/topics/$slug')({
  component: () => {
    const { slug } = Route.useParams()
    return <TopicPage slug={slug} />
  }
})