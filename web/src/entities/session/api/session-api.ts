import { api } from "@/shared/api/base";
import type { TelegramAuthPayload, TokenResponse } from "@/shared/api/types";

export const sessionApi = {
	login: async (payload: TelegramAuthPayload): Promise<TokenResponse> => {
		const { data } = await api.post<TokenResponse>("/auth/login", payload);
		return data;
	},

	refresh: async (refresh_token: string): Promise<TokenResponse> => {
		const { data } = await api.post<TokenResponse>("/auth/refresh", { refresh_token });
		return data;
	},

	devLogin: async (userId: string) => {
    const response = await api.post<{ access_token: string; refresh_token: string }>(
      '/auth/dev-login', 
      { user_id: userId }
    );
    return response.data;
  },
};