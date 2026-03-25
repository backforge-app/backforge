import type { TelegramAuthPayload } from "@/shared/api/types";
import { create } from "zustand";
import { persist } from "zustand/middleware";

interface SessionState {
	accessToken: string | null;
	refreshToken: string | null;
	user: TelegramAuthPayload | null;
	setTokens: (access: string, refresh: string) => void;
	setUser: (user: TelegramAuthPayload) => void;
	logout: () => void;
	isAuthenticated: () => boolean;
}

export const useSession = create<SessionState>()(
	persist(
		(set, get) => ({
			accessToken: null,
			refreshToken: null,
			user: null,

			setTokens: (accessToken, refreshToken) => set({ accessToken, refreshToken }),
			setUser: (user) => set({ user }),
			logout: () => set({ accessToken: null, refreshToken: null, user: null }),
			isAuthenticated: () => !!get().accessToken,
		}),
		{
			name: "backforge-session",
		}
	)
);