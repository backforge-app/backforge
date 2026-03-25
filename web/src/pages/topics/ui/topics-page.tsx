import { Link } from '@tanstack/react-router';
import { Loader2 } from 'lucide-react';
import { useTopics } from '@/entities/topic/model/use-topics';

const getPluralQuestion = (count: number) => {
  const n10 = count % 10;
  const n100 = count % 100;
  if (n10 === 1 && n100 !== 11) return 'вопрос';
  if ([2, 3, 4].includes(n10) && ![12, 13, 14].includes(n100)) return 'вопроса';
  return 'вопросов';
};

export const TopicsPage = () => {
  const { data: topics, status } = useTopics();

  return (
    <div className="mx-auto flex max-w-5xl flex-col gap-7 px-2.5 py-15">
      <div className="flex flex-col gap-5">
        <h1 className="text-h1 text-foreground">Темы</h1>
        <p className="text-body text-muted-foreground">
          Выберите тему и начните практиковать карточки
        </p>
      </div>

      {status === 'pending' && (
        <div className="flex py-10 justify-center">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      )}

      {status === 'success' && topics && (
        <div className="flex flex-col">
          {topics.map((topic) => (
            <Link
              key={topic.id}
              to="/topics/$slug"
              params={{ slug: topic.slug }}
              className="flex items-center justify-between border-b border-border p-2 transition-colors hover:bg-muted/50"
            >
              <span className="text-body text-foreground">{topic.title}</span>
              <span className="text-caption text-muted-foreground">
                {topic.question_count} {getPluralQuestion(topic.question_count)}
              </span>
            </Link>
          ))}

          {topics.length === 0 && (
            <p className="text-body text-muted-foreground p-2">
              Темы пока не добавлены.
            </p>
          )}
        </div>
      )}
    </div>
  );
};