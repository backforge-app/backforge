import { SearchIcon, SquareTerminal, LogOut } from "lucide-react";
import { Link } from "@tanstack/react-router";
import { Avatar, AvatarFallback, AvatarImage } from "@/shared/ui/avatar";
import { ThemeToggle } from "@/features/theme/ui/theme-toggle";
import { InputGroup, InputGroupAddon, InputGroupInput } from "@/shared/ui/input-group";
import { Kbd } from "@/shared/ui/kbd";
import { useSession } from "@/entities/session/model/use-session";
import { Popover, PopoverContent, PopoverTrigger } from "@/shared/ui/popover";
import { DevLogin } from "@/features/auth/ui/dev-login";

const NAV_ITEMS = [
	{ label: "Вопросы", href: "/questions" },
	{ label: "Темы", href: "/topics" },
	{ label: "Прогресс", href: "/progress" },
] as const;

export const Header = () => {
	const { user, logout, isAuthenticated } = useSession();
	const isAuth = isAuthenticated();

	return (
		<header className="h-14 w-full border-b border-border bg-background">
			<div className="mx-auto flex h-full max-w-300 items-center justify-between px-2.5">

				{/* Left Section */}
				<div className="flex items-center gap-12.5">
					<Link to="/" className="flex items-center gap-1.25 outline-none">
						<SquareTerminal className="h-5 w-5 text-primary" />
						<span className="font-mono text-[16px] font-semibold tracking-tight">
							backforge
						</span>
					</Link>

					<nav className="flex items-center">
						{NAV_ITEMS.map((item) => (
							<Link
								key={item.href}
								to={item.href}
								className="flex h-8 items-center rounded-md bg-background px-2.5 text-body-small-medium text-foreground transition-colors hover:bg-secondary"
							>
								{item.label}
							</Link>
						))}
					</nav>
				</div>

				{/* Right Section */}
				<div className="flex items-center gap-3.5">
					<InputGroup className="w-60 bg-background dark:bg-input dark:border-0 rounded-md">
						<InputGroupInput placeholder="Поиск вопросов..." className="text-body-small" />
						<InputGroupAddon>
							<SearchIcon className="h-4 w-4 text-muted-foreground" />
						</InputGroupAddon>
						<InputGroupAddon align="inline-end">
							<Kbd>⌘K</Kbd>
						</InputGroupAddon>
					</InputGroup>

					<ThemeToggle />

					{isAuth && user ? (
						<Popover>
							<PopoverTrigger className="outline-none">
								<Avatar className="h-8 w-8 cursor-pointer transition-opacity hover:opacity-80">
									<AvatarImage src={user.photo_url} alt={user.first_name} />
									<AvatarFallback className="text-caption-bold">
										{user.first_name?.charAt(0).toUpperCase()}
									</AvatarFallback>
								</Avatar>
							</PopoverTrigger>
							<PopoverContent align="end" className="w-48 p-2">
								<div className="flex flex-col gap-2">
									<div className="px-2 py-1.5 text-sm font-medium border-b border-border">
										{user.first_name} {user.last_name}
									</div>
									<button
										onClick={logout}
										className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm text-destructive hover:bg-destructive/10 transition-colors"
									>
										<LogOut className="h-4 w-4" />
										Выйти
									</button>
								</div>
							</PopoverContent>
						</Popover>
					) : (
						<div className="flex items-center gap-2">
               <DevLogin />
            </div>
					)}
				</div>

			</div>
		</header>
	);
};