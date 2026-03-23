import { api } from '@/shared/api/base';
import type { TagDto } from '@/shared/api/types';

export const tagApi = {
	list: async () => {
		const { data } = await api.get<TagDto[]>('/tags');
		return data;
	},
};