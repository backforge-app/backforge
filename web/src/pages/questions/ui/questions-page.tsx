import { useState } from 'react';
import { useDebounce } from 'use-debounce';
import { SearchIcon, Loader2 } from "lucide-react";
import { useQuestions } from '@/entities/question/model/use-questions';
import { QuestionCard } from '@/entities/question/ui/question-card';
import { QuestionSkeleton } from '@/entities/question/ui/question-skeleton';
import { InputGroup, InputGroupAddon, InputGroupInput } from "@/shared/ui/input-group";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { Button } from "@/shared/ui/button";
import { useTags } from '@/entities/tag/model/use-tags';
import { MultiSelect } from "@/shared/ui/multi-select"

const levels = [
	{ value: "all", label: "Все уровни" },
	{ value: "beginner", label: "Начальный" },
	{ value: "medium", label: "Средний" },
	{ value: "advanced", label: "Продвинутый" },
]

export const QuestionsPage = () => {
	// States
	const [search, setSearch] = useState('');
	const [level, setLevel] = useState<string | null>('all');
	const [selectedTopics, setSelectedTopics] = useState<string[]>([]);

	const [debouncedSearch] = useDebounce(search, 300);

	// Data fetching
	const { data: tagsData } = useTags();
	const availableTags =
		tagsData?.map((t) => ({
			label: t.name,
			value: t.name,
		})) || []

	const { data, fetchNextPage, hasNextPage, isFetchingNextPage, status, error } = useQuestions({
		search: debouncedSearch || undefined,
		level: (level === 'all' || !level) ? undefined : level,
		tags: selectedTopics.length > 0 ? selectedTopics : undefined,
	});

	const allQuestions = data?.pages.flat() || [];

	if (status === 'error') {
		return (
			<div className="flex h-100 flex-col items-center justify-center gap-4 px-4 text-center">
				<p className="text-destructive font-medium text-body">
					Ошибка: {error.message}
				</p>
				<Button onClick={() => window.location.reload()} size="sm">
					Попробовать снова
				</Button>
			</div>
		);
	}

	return (
		<div className="mx-auto flex max-w-5xl flex-col gap-7 px-2.5 py-15">
			<h1 className="text-h1 text-foreground">Вопросы</h1>

			<div className="flex items-start gap-3.5">
				{/* Search */}
				<InputGroup className="flex-1 bg-background dark:bg-input dark:border-0 rounded-[8px]">
					<InputGroupInput
						placeholder="Поиск вопросов..."
						className="text-body-small"
						value={search}
						onChange={(e) => setSearch(e.target.value)}
					/>
					<InputGroupAddon><SearchIcon className="h-4 w-4 text-muted-foreground" /></InputGroupAddon>
				</InputGroup>

				{/* Level Select */}
				<Select value={level} onValueChange={setLevel}>
					<SelectTrigger className="w-44 bg-background dark:bg-input dark:border-0 rounded-[8px] px-3">
						<SelectValue>
							{
								levels.find((l) => l.value === level)?.label ??
								"Все уровни"
							}
						</SelectValue>
					</SelectTrigger>
					<SelectContent alignItemWithTrigger={false}>
						{levels.map((l) => (
							<SelectItem key={l.value} value={l.value}>
								{l.label}
							</SelectItem>
						))}
					</SelectContent>
				</Select>

				{/* Topics Multi Select */}
				<MultiSelect
					className='w-44 bg-background dark:bg-input'
					items={availableTags}
					value={selectedTopics}
					onChange={setSelectedTopics}
					placeholder="Выберите темы..."
				/>
			</div>

			<p className="text-muted-foreground text-body-small">
				Найдено вопросов: {allQuestions.length}
			</p>

			<div className="flex flex-col gap-3.5">
				{status === 'pending' ? (
					<>
						<QuestionSkeleton />
						<QuestionSkeleton />
						<QuestionSkeleton />
					</>
				) : (
					allQuestions.map((q) => (
						<QuestionCard
							key={q.id}
							title={q.title}
							slug={q.slug}
							tags={q.tags}
							difficulty={q.level as any}
							isNew={q.is_new}
						/>
					))
				)}
			</div>

			{hasNextPage && (
				<div className="flex justify-center pt-4">
					<Button
						variant="outline"
						onClick={() => fetchNextPage()}
						disabled={isFetchingNextPage}
					>
						{isFetchingNextPage ? (
							<Loader2 className="mr-2 h-4 w-4 animate-spin" />
						) : null}
						Показать еще
					</Button>
				</div>
			)}
		</div>
	);
};