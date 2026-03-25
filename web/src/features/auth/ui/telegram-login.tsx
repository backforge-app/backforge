import { useEffect, useRef } from "react";
import { sessionApi } from "@/entities/session/api/session-api";
import { useSession } from "@/entities/session/model/use-session";
import type { TelegramAuthPayload } from "@/shared/api/types";

interface TelegramLoginProps {
	botName: string;
}

export const TelegramLogin = ({ botName }: TelegramLoginProps) => {
	const containerRef = useRef<HTMLDivElement>(null);
	const { setTokens, setUser } = useSession();

	useEffect(() => {
		if (!containerRef.current) return;

		window.onTelegramAuth = async (user: TelegramAuthPayload) => {
			try {
				const tokens = await sessionApi.login(user);
				setTokens(tokens.access_token, tokens.refresh_token);
				setUser(user);
			} catch (error) {
				console.error("Failed to login with Telegram", error);
			}
		};

		const script = document.createElement("script");
		script.src = "https://telegram.org/js/telegram-widget.js?22";
		script.setAttribute("data-telegram-login", botName);
		script.setAttribute("data-size", "medium");
		script.setAttribute("data-radius", "10");
		script.setAttribute("data-onauth", "onTelegramAuth(user)");
		script.setAttribute("data-request-access", "write");
		script.async = true;

		containerRef.current.innerHTML = "";
		containerRef.current.appendChild(script);

		return () => {
			delete window.onTelegramAuth;
		};
	}, [botName, setTokens, setUser]);

	return <div ref={containerRef} className="h-9.5" />;
};

declare global {
	interface Window {
		onTelegramAuth?: (user: TelegramAuthPayload) => void;
	}
}