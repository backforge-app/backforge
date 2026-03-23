import { useQuery } from '@tanstack/react-query';
import { tagApi } from '../api/tag-api';

export const useTags = () => {
  return useQuery({
    queryKey: ['tags'],
    queryFn: tagApi.list,
    staleTime: 1000 * 60 * 5,
  });
};