import { api } from '@/shared/api/base';
import type { QuestionCardDto, ListQuestionsParams, QuestionDetailDto } from '@/shared/api/types';

export const questionApi = {
	listCards: async (params: ListQuestionsParams) => {
		const queryParams = {
			...params,
			tags: params.tags?.join(','),
		};

		const { data } = await api.get<QuestionCardDto[]>('/questions', {
			params: queryParams,
		});
		return data;
	},

	getBySlug: async (slug: string) => {
		const { data } = await api.get<QuestionDetailDto>(`/questions/${slug}`);
		return data;
	}
};