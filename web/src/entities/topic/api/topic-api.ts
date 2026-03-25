import { api } from '@/shared/api/base';
import type { TopicDetailDto, TopicRowDto } from '@/shared/api/types';

export const topicApi = {
	listRows: async () => {
		const { data } = await api.get<TopicRowDto[]>('/topics');
		return data;
	},

	getBySlug: async (slug: string) => {
    const { data } = await api.get<TopicDetailDto>(`/topics/slug/${slug}`);
    return data;
  }
};