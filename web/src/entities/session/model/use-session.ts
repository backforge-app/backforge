import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { TelegramAuthPayload } from "@/shared/api/types";
import { sessionApi } from "../api/session-api";

interface SessionState {
  accessToken: string | null;
  refreshToken: string | null;
  user: TelegramAuthPayload | null;
  
  // Actions
  setTokens: (access: string, refresh: string) => void;
  setUser: (user: TelegramAuthPayload) => void;
  logout: () => void;
  
  // Dev Action
  loginDev: (userId: string) => Promise<void>;
  
  // Selectors
  isAuthenticated: () => boolean;
}

export const useSession = create<SessionState>()(
  persist(
    (set, get) => ({
      accessToken: null,
      refreshToken: null,
      user: null,

      setTokens: (accessToken, refreshToken) => 
        set({ accessToken, refreshToken }),
      
      setUser: (user) => 
        set({ user }),

      logout: () => 
        set({ accessToken: null, refreshToken: null, user: null }),

      // Реализация входа для разработки
      loginDev: async (userId: string) => {
        try {
          const data = await sessionApi.devLogin(userId);
          
          set({ 
            accessToken: data.access_token, 
            refreshToken: data.refresh_token 
          });

          // Опционально: здесь можно сразу вызвать метод получения 
          // данных профиля, если бэкенд не отдает их в login-ответе
          // const userProfile = await userApi.getMe();
          // set({ user: userProfile });

        } catch (error) {
          console.error("Dev login failed:", error);
          throw error;
        }
      },

      isAuthenticated: () => !!get().accessToken,
    }),
    {
      name: "backforge-session",
      // Фильтруем стейт, чтобы не хранить функции в localStorage
      partialize: (state) => ({
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        user: state.user,
      }),
    }
  )
);