import { createLazyFileRoute } from '@tanstack/react-router'

export const Route = createLazyFileRoute('/')({
  component: Index,
})

function Index() {
  return (
    <div className="p-2">
      <h3 className="text-2xl font-bold">Welcome to Backforge!</h3>
      <p className="mt-2 text-muted-foreground">
        Select a section from the navigation to start.
      </p>
    </div>
  )
}