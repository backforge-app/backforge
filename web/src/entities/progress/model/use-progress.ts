import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { progressApi } from '../api/progress-api';
import type { MarkProgressDto } from '@/shared/api/types';

export const useTopicProgress = (topicId?: string) => {
  return useQuery({
    queryKey: ['progress', topicId],
    queryFn: () => progressApi.getTopicProgress(topicId!),
    enabled: !!topicId,
  });
};

export const useMarkProgress = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ type, payload }: { type: 'known' | 'learned' | 'skipped', payload: MarkProgressDto }) => {
      if (type === 'known') return progressApi.markKnown(payload);
      if (type === 'learned') return progressApi.markLearned(payload);
      return progressApi.markSkipped(payload);
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['progress', variables.payload.topic_id] });
    }
  });
};

export const useResetProgress = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (topicId: string) => progressApi.resetTopic(topicId),
    onSuccess: (_, topicId) => {
      queryClient.invalidateQueries({ queryKey: ['progress', topicId] });
    }
  });
};