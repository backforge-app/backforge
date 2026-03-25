import { useState, useEffect } from "react";
import { Link } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import {
	ArrowLeft, ChevronLeft, ChevronRight,
	CircleCheckBig, Sparkles, SkipForward, RotateCcw, Loader2,
	Info
} from "lucide-react";

import { topicApi } from "@/entities/topic/api/topic-api";
import { questionApi } from "@/entities/question/api/question-api";
import { useTopicProgress, useMarkProgress, useResetProgress } from "@/entities/progress/model/use-progress";
import { Progress } from "@/shared/ui/progress";
import { Badge } from "@/shared/ui/badge";

const levelLabelMap: Record<string, string> = {
	Beginner: "начинающих",
	Medium: "средних",
	Advanced: "продвинутых",
};

const isAuthenticated = false

export const TopicPage = ({ slug }: { slug: string }) => {
	const [currentIndex, setCurrentIndex] = useState(0);
	const [showAnswer, setShowAnswer] = useState(false);
	const [isInitialized, setIsInitialized] = useState(false);

	const [localProgress, setLocalProgress] = useState({
		known: 0,
		learned: 0,
		skipped: 0
	});

	const { data: topic, status: topicStatus } = useQuery({
		queryKey: ['topic', slug],
		queryFn: () => topicApi.getBySlug(slug),
	});

	const { data: questions = [], status: questionsStatus } = useQuery({
		queryKey: ['topic-questions', topic?.id],
		queryFn: () => questionApi.listByTopic(topic!.id),
		enabled: !!topic?.id,
	});

	const { data: serverProgress } = useTopicProgress(topic?.id);
	const markMutation = useMarkProgress();
	const resetMutation = useResetProgress();

	const displayProgress = isAuthenticated
		? {
			known: serverProgress?.known || 0,
			learned: serverProgress?.learned || 0,
			skipped: serverProgress?.skipped || 0
		}
		: localProgress;

	useEffect(() => {
		if (isAuthenticated && serverProgress && !isInitialized && questions.length > 0) {
			setCurrentIndex(Math.min(serverProgress.current_position, questions.length - 1));
			setIsInitialized(true);
		}
	}, [serverProgress, isInitialized, questions.length, isAuthenticated]);

	if (topicStatus === 'pending' || questionsStatus === 'pending') {
		return <div className="flex h-[50vh] items-center justify-center"><Loader2 className="h-8 w-8 animate-spin text-primary" /></div>;
	}

	if (!topic) return null;

	const currentQuestion = questions[currentIndex];
	const progressPercent = questions.length > 0 ? ((currentIndex + 1) / questions.length) * 100 : 0;

	const goNext = () => {
		if (currentIndex < questions.length - 1) {
			setCurrentIndex(p => p + 1);
			setShowAnswer(false);
		}
	};

	const goPrev = () => {
		if (currentIndex > 0) {
			setCurrentIndex(p => p - 1);
			setShowAnswer(false);
		}
	};

	const handleMark = (type: 'known' | 'learned' | 'skipped') => {
		if (!topic || !questions[currentIndex]) return;

		if (isAuthenticated) {
			markMutation.mutate({
				type,
				payload: { topic_id: topic.id, question_id: questions[currentIndex].id }
			});
		} else {
			setLocalProgress(prev => ({
				...prev,
				[type]: prev[type] + 1
			}));
		}
		goNext();
	};

	const handleReset = () => {
		if (isAuthenticated && topic) {
			resetMutation.mutate(topic.id);
		} else {
			setLocalProgress({ known: 0, learned: 0, skipped: 0 });
		}
		setCurrentIndex(0);
		setShowAnswer(false);
	};

	const questionsByLevel = questions.reduce((acc, q) => {
		acc[q.level] = acc[q.level] || [];
		acc[q.level].push(q);
		return acc;
	}, {} as Record<string, typeof questions>);

	return (
		<div className="mx-auto flex w-full max-w-175 flex-col gap-7 px-2.5 py-15">
			<Link
				to="/topics"
				className="text-muted-foreground hover:text-foreground inline-flex w-fit items-center gap-2.5 transition-colors"
			>
				<ArrowLeft className="h-4 w-4" />
				<span className="text-body-small-medium">Назад к темам</span>
			</Link>

			<div className="flex flex-col gap-5">
				<h1 className="text-h1 text-foreground leading-tight">{topic.title}</h1>
				<p className="text-body text-muted-foreground">
					Тренируйтесь на карточках с вопросами и проверяйте свои знания.
					Внизу вы можете увидеть полный список всех вопросов по этой теме.
				</p>
			</div>

			{questions.length > 0 ? (
				<>
					{!isAuthenticated && (
						<div className="flex items-start gap-3 p-4 border border-border rounded-[10px] bg-muted/30">
							<Info className="w-5 h-5 text-muted-foreground shrink-0 mt-0.5" />
							<p className="text-body-small text-muted-foreground">
								Вы не авторизованы. Ваш прогресс будет виден сейчас, но <strong>не сохранится</strong> после закрытия или обновления страницы.
								<Link to="/" className="text-foreground underline ml-1">Войти</Link>
							</p>
						</div>
					)}

					<div className="flex flex-col gap-3.5">
						<div className="border border-border rounded-[10px] p-6 flex flex-col gap-2.5 bg-card">

							<div className="flex items-center gap-2.5">
								<Progress
									value={progressPercent}
									className="flex-1"
									trackClassName="h-2 rounded-full"
								/>
								<div className="flex items-center gap-1.25">
									<button
										onClick={goPrev}
										disabled={currentIndex === 0}
										className="text-muted-foreground transition-colors hover:text-foreground disabled:pointer-events-none disabled:opacity-30"
									>
										<ChevronLeft className="w-4 h-4" />
									</button>
									<span className="text-caption-bold text-foreground w-12.5 text-center">
										{currentIndex + 1} / {questions.length}
									</span>
									<button
										onClick={goNext}
										disabled={currentIndex === questions.length - 1}
										className="text-muted-foreground transition-colors hover:text-foreground disabled:pointer-events-none disabled:opacity-30"
									>
										<ChevronRight className="w-4 h-4" />
									</button>
								</div>
							</div>

							<div className="flex flex-wrap items-center gap-5 pt-2">
								<StatItem icon={<CircleCheckBig className="w-4 h-4 text-foreground" />} label="Знал" value={displayProgress.known} />
								<StatItem icon={<Sparkles className="w-4 h-4 text-foreground" />} label="Выучил" value={displayProgress.learned} />
								<StatItem icon={<SkipForward className="w-4 h-4 text-foreground" />} label="Пропустил" value={displayProgress.skipped} />

								<button onClick={handleReset} className="flex items-center gap-1.25 text-destructive hover:opacity-80 cursor-pointer transition-opacity ml-auto sm:ml-0">
									<RotateCcw className="w-4 h-4" />
									<span className="text-body-small">Сбросить</span>
								</button>
							</div>
						</div>

						<div className="border border-border rounded-[10px] p-8 flex flex-col items-center gap-2.5 bg-card min-h-96.5 transition-all duration-300">
							<span className="text-body text-muted-foreground">
								Вопросы для собеседования по {topic.title} для {levelLabelMap[currentQuestion.level] || 'всех'}
							</span>

							<h2 className="flex items-center justify-center h-65 text-center text-question text-foreground">
								{currentQuestion.title}
							</h2>

							{showAnswer ? (
								<div className="flex flex-col gap-5 items-center w-full">
									<div className="flex items-center gap-5 w-full">
										<div className="h-px bg-border flex-1" />
										<span className="text-body text-muted-foreground">Ответ</span>
										<div className="h-px bg-border flex-1" />
									</div>

									{/* Временный JSON блок до подключения TipTap */}
									<pre className="text-body text-foreground whitespace-pre-wrap font-mono bg-muted/30 p-4 rounded-lg overflow-x-auto w-full">
										{JSON.stringify(currentQuestion.content, null, 2)}
									</pre>

									<button onClick={() => setShowAnswer(false)} className="text-body text-muted-foreground underline text-center hover:text-foreground w-fit cursor-pointer transition-colors">
										Скрыть ответ
									</button>
								</div>
							) : (
								<button onClick={() => setShowAnswer(true)} className="text-body text-muted-foreground underline text-center hover:text-foreground w-fit cursor-pointer transition-colors">
									Показать ответ
								</button>
							)}
						</div>

						<div className="flex flex-wrap items-center gap-3.5">
							<button
								onClick={() => handleMark('known')}
								className="flex flex-1 items-center justify-center gap-2.5 px-3 py-3 border border-border rounded-[10px] bg-card hover:bg-muted cursor-pointer transition-colors"
							>
								<CircleCheckBig className="w-4.5 h-4.5 text-foreground" />
								<span className="text-body text-foreground">Знаю этот вопрос</span>
							</button>

							<button
								onClick={() => handleMark('learned')}
								className="flex flex-1 items-center justify-center gap-2.5 px-3 py-3 border border-border rounded-[10px] bg-card hover:bg-muted cursor-pointer transition-colors"
							>
								<Sparkles className="w-4.5 h-4.5 text-foreground" />
								<span className="text-body text-foreground">Не знал этот вопрос</span>
							</button>

							<button
								onClick={() => handleMark('skipped')}
								className="flex flex-1 items-center justify-center gap-2.5 px-3 py-3 border border-border rounded-[10px] bg-card hover:bg-destructive/10 cursor-pointer group transition-colors"
							>
								<SkipForward className="w-4.5 h-4.5 text-destructive" />
								<span className="text-body text-destructive">Пропустить</span>
							</button>
						</div>
					</div>
				</>
			) : (
				<div className="text-center py-10 border border-border rounded-lg bg-card text-muted-foreground">
					В этой теме пока нет вопросов.
				</div>
			)}

			{/* --- Справочник вопросов --- */}
			{questions.length > 0 && (
				<div className="flex flex-col gap-7">
					<div className="flex flex-col gap-5">
						<h2 className="text-h2 text-foreground">Список вопросов</h2>
						<p className="text-body text-muted-foreground">
							Для удобства вы можете просмотреть все вопросы этой темы в списке ниже.
						</p>
					</div>

					<div className="flex flex-col gap-7">
						{['Beginner', 'Medium', 'Advanced'].map((level) => {
							const levelQuestions = questionsByLevel[level];
							if (!levelQuestions || levelQuestions.length === 0) return null;

							return (
								<div key={level} className="flex flex-col gap-7">
									<h3 className="text-h3 text-foreground">
										Вопросы для собеседования по {topic.title} для {levelLabelMap[level]}
									</h3>

									{levelQuestions.map((q) => (
										<div key={q.id} className="flex flex-col gap-7 border-b border-border/50 last:border-0">
											<h4 className="text-h4 text-foreground">{q.title}</h4>
											<pre className="text-body text-muted-foreground bg-muted/20 p-4 rounded-md overflow-x-auto text-sm">
												{JSON.stringify(q.content, null, 2)}
											</pre>
										</div>
									))}
								</div>
							);
						})}
					</div>
				</div>
			)}

		</div>
	);
};

const StatItem = ({ icon, label, value }: { icon: React.ReactNode, label: string, value: number }) => (
	<div className="flex items-center gap-2.5">
		<div className="flex items-center gap-1.25">
			{icon}
			<span className="text-body-small text-foreground">{label}</span>
		</div>
		<Badge variant="secondary" className="bg-secondary text-foreground text-caption-bold rounded-full">
			{value}
		</Badge>
	</div>
);
