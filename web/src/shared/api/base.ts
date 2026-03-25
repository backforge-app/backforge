import axios, { AxiosError, type InternalAxiosRequestConfig } from 'axios';
import type { ApiError } from './types';
import { sessionApi } from "@/entities/session/api/session-api";
import { useSession } from "@/entities/session/model/use-session";

interface CustomAxiosRequestConfig extends InternalAxiosRequestConfig {
	_retry?: boolean;
}

export const api = axios.create({
	baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1',
	headers: {
		'Content-Type': 'application/json',
	},
});

api.interceptors.request.use((config) => {
	const token = useSession.getState().accessToken;
	if (token && config.headers) {
		config.headers.Authorization = `Bearer ${token}`;
	}
	return config;
});

api.interceptors.response.use(
	(response) => response,
	async (error: AxiosError<ApiError>) => {
		const originalRequest = error.config as CustomAxiosRequestConfig;

		if (error.response?.status === 401 && originalRequest && !originalRequest._retry) {
			originalRequest._retry = true;

			const refreshToken = useSession.getState().refreshToken;

			if (!refreshToken) {
				useSession.getState().logout();
			} else {
				try {
					const { access_token, refresh_token } = await sessionApi.refresh(refreshToken);

					useSession.getState().setTokens(access_token, refresh_token);

					if (originalRequest.headers) {
						originalRequest.headers.Authorization = `Bearer ${access_token}`;
					}

					return api(originalRequest);
				} catch (refreshError) {
					useSession.getState().logout();

					const refreshMessage = (refreshError as AxiosError<ApiError>).response?.data?.error || 'Сессия истекла. Пожалуйста, авторизуйтесь заново';
					return Promise.reject(new Error(refreshMessage));
				}
			}
		}

		const message = error.response?.data?.error || 'Произошла непредвиденная ошибка';
		return Promise.reject(new Error(message));
	}
);