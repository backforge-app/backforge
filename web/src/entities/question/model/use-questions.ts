import { useInfiniteQuery } from '@tanstack/react-query';
import { questionApi } from '../api/question-api';
import type { ListQuestionsParams } from '@/shared/api/types';

export const useQuestions = (filters: Omit<ListQuestionsParams, 'limit' | 'offset'>) => {
	const LIMIT = 20;

	return useInfiniteQuery({
		queryKey: ['questions', filters],
		queryFn: ({ pageParam = 0 }) =>
			questionApi.listCards({
				...filters,
				limit: LIMIT,
				offset: pageParam
			}),
		getNextPageParam: (lastPage, allPages) => {
			return lastPage.length === LIMIT ? allPages.length * LIMIT : undefined;
		},
		initialPageParam: 0,
	});
};