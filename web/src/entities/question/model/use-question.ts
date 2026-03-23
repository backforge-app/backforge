import { useQuery } from '@tanstack/react-query';
import { questionApi } from '../api/question-api';

export const useQuestion = (slug: string) => {
  return useQuery({
    queryKey: ['question', slug],
    queryFn: () => questionApi.getBySlug(slug),
    enabled: !!slug,
  });
};