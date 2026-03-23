import { Moon, Sun } from "lucide-react";
import { useTheme } from "@/app/providers/theme-provider.tsx";
import { Button } from "@/shared/ui/button";

export const ThemeToggle = () => {
	const { theme, setTheme } = useTheme();

	const toggleTheme = () => {
		setTheme(theme === "dark" ? "light" : "dark");
	};

	return (
		<Button
			variant="ghost"
			size="icon"
			onClick={toggleTheme}
			className="h-8 w-8 rounded-[8px]"
		>
			<Sun className="h-4 w-4 transition-all dark:hidden" />
			<Moon className="hidden h-4 w-4 transition-all dark:block" />
			<span className="sr-only">Toggle theme</span>
		</Button>
	);
};