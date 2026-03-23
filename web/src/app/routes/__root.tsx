import { createRootRoute, Outlet } from '@tanstack/react-router'
import { TanStackRouterDevtools } from '@tanstack/router-devtools'
import { Header } from '@/widgets/header'

export const Route = createRootRoute({
	component: () => (
		<>
			<div className="relative flex min-h-screen flex-col">
				<Header />
				<main className="flex-1">
					<Outlet />
				</main>
			</div>
			{/* Devtools only in development */}
			{import.meta.env.DEV && <TanStackRouterDevtools />}
		</>
	),
})