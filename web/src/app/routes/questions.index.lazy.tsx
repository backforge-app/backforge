import { createLazyFileRoute } from '@tanstack/react-router'
import { QuestionsPage } from '@/pages/questions/ui/questions-page'

export const Route = createLazyFileRoute('/questions/')({
  component: QuestionsPage,
})