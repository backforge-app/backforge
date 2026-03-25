import { useQuery } from '@tanstack/react-query';
import { topicApi } from '../api/topic-api';

export const useTopics = () => {
  return useQuery({
    queryKey: ['topics'],
    queryFn: topicApi.listRows,
  });
};