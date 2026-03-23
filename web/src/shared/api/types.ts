export interface ApiError {
	error: string;
	details?: any;
}

export interface QuestionCardDto {
	id: string;
	title: string;
	slug: string;
	level: 'beginner' | 'intermediate' | 'advanced';
	tags: string[];
	is_new: boolean;
	is_free: boolean;
}

export interface ListQuestionsParams {
	limit?: number;
	offset?: number;
	search?: string;
	level?: string;
	tags?: string[];
}

export interface TagDto {
	id: string;
	name: string;
}

export interface QuestionDetailDto {
	id: string;
	title: string;
	slug: string;
	content: Record<string, any>;
	level: 'Beginner' | 'Medium' | 'Advanced';
	topic_id?: string;
	is_free: boolean;
	tag_ids: string[];
	created_by?: string;
	updated_by?: string;
}