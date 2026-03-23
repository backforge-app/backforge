import { ArrowLeft, Loader2 } from "lucide-react";
import { Link } from "@tanstack/react-router";
import { cn } from "@/shared/lib/utils";
import { useQuestion } from "@/entities/question/model/use-question";
import { useTags } from "@/entities/tag/model/use-tags";
import { Button } from "@/shared/ui/button";

interface QuestionPageProps {
	slug: string;
}

const difficultyMap: Record<string, { label: string; className: string }> = {
	Beginner: { label: "Начальный", className: "text-lime bg-lime/10" },
	Medium: { label: "Средний", className: "text-blue bg-blue/10" },
	Advanced: { label: "Продвинутый", className: "text-violet bg-violet/10" },
};

export const QuestionPage = ({ slug }: QuestionPageProps) => {
	const { data: question, status, error } = useQuestion(slug);
	const { data: tagsData } = useTags();

	if (status === 'pending') {
		return (
			<div className="flex h-[50vh] items-center justify-center">
				<Loader2 className="h-8 w-8 animate-spin text-primary" />
			</div>
		);
	}

	if (status === 'error') {
		return (
			<div className="flex h-[50vh] flex-col items-center justify-center gap-4">
				<p className="text-destructive text-body-medium">Ошибка: {error.message}</p>
				<Link to="/questions">
					<Button variant="outline">Вернуться к списку</Button>
				</Link>
			</div>
		);
	}

	if (!question) return null;

	const diff = difficultyMap[question.level] || difficultyMap.Beginner;

	const resolvedTags = question.tag_ids
		.map((id) => tagsData?.find((t) => t.id === id)?.name)
		.filter(Boolean) as string[];

	return (
		<div className="mx-auto flex w-full max-w-175 flex-col gap-7 px-2.5 py-15">
			<Link
				to="/questions"
				className="text-muted-foreground hover:text-foreground inline-flex w-fit items-center gap-2.5 transition-colors"
			>
				<ArrowLeft className="h-4 w-4" />
				<span className="text-body-small-medium">Назад к вопросам</span>
			</Link>

			<h1 className="text-h1 text-foreground">{question.title}</h1>

			<div className="flex flex-wrap items-center gap-2.5">
				<span
					className={cn(
						"rounded-md px-2.5 py-1.25 text-caption!",
						diff.className
					)}
				>
					{diff.label}
				</span>

				{resolvedTags.map((tagName) => (
					<span
						key={tagName}
						className="border-border bg-background dark:bg-secondary dark:border-0 text-foreground text-caption rounded-md border px-2.5 py-1.25"
					>
						{tagName}
					</span>
				))}
			</div>

			<div className="text-body text-foreground mt-4 overflow-x-auto">
				<pre className="bg-muted/30 rounded-lg p-4 font-mono text-sm whitespace-pre-wrap">
					{JSON.stringify(question.content, null, 2)}
				</pre>
			</div>
		</div>
	);
};