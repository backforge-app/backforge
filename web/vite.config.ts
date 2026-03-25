import path from "path"
import { tanstackRouter } from '@tanstack/router-vite-plugin'
import tailwindcss from '@tailwindcss/vite'
import tsconfigPaths from 'vite-tsconfig-paths'
import react from "@vitejs/plugin-react"
import { defineConfig } from "vite"

// https://vite.dev/config/
export default defineConfig({
	plugins: [
    react(),
    tanstackRouter({
      routesDirectory: './src/app/routes',
      generatedRouteTree: './src/app/routeTree.gen.ts',
      quoteStyle: 'single',
      semicolons: false,
    }),
    tailwindcss(),
    tsconfigPaths(),
  ],
	resolve: {
		alias: {
			"@": path.resolve(__dirname, "./src"),
		},
	},
	server: {
    host: '127.0.0.1',
    port: 80,
    strictPort: true,
  },
})
