import { createLazyFileRoute } from '@tanstack/react-router';
import { TopicsPage } from '@/pages/topics/ui/topics-page';

export const Route = createLazyFileRoute('/topics/')({
	component: TopicsPage,
});