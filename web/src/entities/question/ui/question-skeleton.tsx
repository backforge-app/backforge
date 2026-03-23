import { Skeleton } from "@/shared/ui/skeleton";
import { cn } from "@/shared/lib/utils";

export const QuestionSkeleton = ({ className }: { className?: string }) => {
	return (
		<div
			className={cn(
				"flex w-full flex-col gap-2.5 rounded-[10px] border border-border bg-card p-4",
				className
			)}
		>
			<div className="flex items-start justify-between gap-4">
				<div className="flex flex-col gap-2 w-full">
					<Skeleton className="h-5 w-[70%] rounded-sm" />
					<Skeleton className="h-5 w-[40%] rounded-sm" />
				</div>

				<Skeleton className="h-6 w-15 rounded-full shrink-0" />
			</div>

			<div className="flex flex-wrap items-center gap-2.5">
				<Skeleton className="h-6.5 w-24 rounded-md" />

				<Skeleton className="h-6.5 w-18 rounded-md" />
				<Skeleton className="h-6.5 w-20 rounded-md" />
			</div>
		</div>
	);
};