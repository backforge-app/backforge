import { Badge } from "@/shared/ui/badge";
import { cn } from "@/shared/lib/utils";
import { Link } from "@tanstack/react-router";

type Difficulty = "Beginner" | "Medium" | "Advanced";

interface QuestionCardProps {
	title: string;
	slug: string;
	tags: string[];
	difficulty: Difficulty;
	isNew?: boolean;
	className?: string;
}

const difficultyMap: Record<Difficulty, { label: string; className: string }> = {
	Beginner: {
		label: "Начальный",
		className: "text-lime bg-lime/10",
	},
	Medium: {
		label: "Средний",
		className: "text-blue bg-blue/10",
	},
	Advanced: {
		label: "Продвинутый",
		className: "text-violet bg-violet/10",
	},
};

export const QuestionCard = ({
	title,
	slug,
	tags,
	difficulty,
	isNew,
	className,
}: QuestionCardProps) => {
	const diff = difficultyMap[difficulty] || difficultyMap.Beginner;

	return (
		<Link
			to="/questions/$slug"
			params={{ slug }}
			className={cn("flex w-full flex-col gap-2.5 rounded-[10px] border border-border p-4 bg-card hover:border-primary/50 transition-colors cursor-pointer", className)}
		>
			{/* Upper Part */}
			<div className="flex items-start justify-between gap-4">
				<h4 className="text-h4 text-foreground line-clamp-2">
					{title}
				</h4>
				{isNew && (
					<Badge variant="default" className="bg-primary text-primary-foreground shrink-0 rounded-full">
						Новый
					</Badge>
				)}
			</div>

			{/* Lower Part (Badges & Tags) */}
			<div className="flex flex-wrap items-center gap-2.5">
				{/* Difficulty Tag */}
				<span
					className={cn(
						"rounded-md px-2.5 py-1.25 text-caption!",
						diff.className
					)}
				>
					{diff.label}
				</span>

				{/* Topic Tags */}
				{tags && tags.map((tag) => (
					<span
						key={tag}
						className="rounded-md border border-border bg-background dark:bg-secondary dark:border-0 px-2.5 py-1.25 text-foreground text-caption"
					>
						{tag}
					</span>
				))}
			</div>
		</Link>
	);
};