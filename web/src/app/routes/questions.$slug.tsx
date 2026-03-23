import { createFileRoute } from '@tanstack/react-router'
import { QuestionPage } from '@/pages/question/ui/question-page'

export const Route = createFileRoute('/questions/$slug')({
	component: () => {
		const { slug } = Route.useParams()

		return <QuestionPage slug={slug} />
	}
})