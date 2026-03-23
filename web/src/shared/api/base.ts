import axios, { AxiosError } from 'axios';
import type { ApiError } from './types';

export const api = axios.create({
	baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1',
	headers: {
		'Content-Type': 'application/json',
	},
});

api.interceptors.response.use(
	(response) => response,
	(error: AxiosError<ApiError>) => {
		const message = error.response?.data?.error || 'Произошла непредвиденная ошибка';

		return Promise.reject(new Error(message));
	}
);