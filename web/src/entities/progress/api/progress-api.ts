import { api } from '@/shared/api/base';
import type { ProgressAggregateDto, MarkProgressDto } from '@/shared/api/types';

export const progressApi = {
  getTopicProgress: async (topicId: string) => {
    const { data } = await api.get<ProgressAggregateDto>(`/progress/topics/${topicId}`);
    return data;
  },
  markKnown: async (payload: MarkProgressDto) => {
    await api.post('/progress/known', payload);
  },
  markLearned: async (payload: MarkProgressDto) => {
    await api.post('/progress/learned', payload);
  },
  markSkipped: async (payload: MarkProgressDto) => {
    await api.post('/progress/skipped', payload);
  },
  resetTopic: async (topicId: string) => {
    await api.delete(`/progress/topics/${topicId}`);
  }
};