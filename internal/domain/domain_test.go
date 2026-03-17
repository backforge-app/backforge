package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestProgressStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status ProgressStatus
		want   bool
	}{
		{"new is valid", ProgressStatusNew, true},
		{"skipped is valid", ProgressStatusSkipped, true},
		{"known is valid", ProgressStatusKnown, true},
		{"learned is valid", ProgressStatusLearned, true},
		{"empty is invalid", ProgressStatus(""), false},
		{"random string is invalid", ProgressStatus("deleted"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status.IsValid())
		})
	}
}

func TestQuestionLevel_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		level QuestionLevel
		want  bool
	}{
		{"beginner is valid", QuestionLevelBeginner, true},
		{"medium is valid", QuestionLevelMedium, true},
		{"advanced is valid", QuestionLevelAdvanced, true},
		{"negative value is invalid", QuestionLevel(-1), false},
		{"out of range is invalid", QuestionLevel(3), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.level.IsValid())
		})
	}
}

func TestQuestionLevel_String(t *testing.T) {
	tests := []struct {
		level QuestionLevel
		want  string
	}{
		{QuestionLevelBeginner, "Beginner"},
		{QuestionLevelMedium, "Medium"},
		{QuestionLevelAdvanced, "Advanced"},
		{QuestionLevel(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.level.String())
		})
	}
}

func TestUser_IsAdmin(t *testing.T) {
	t.Run("admin role", func(t *testing.T) {
		u := &User{Role: UserRoleAdmin}
		assert.True(t, u.IsAdmin())
	})

	t.Run("user role", func(t *testing.T) {
		u := &User{Role: UserRoleUser}
		assert.False(t, u.IsAdmin())
	})

	t.Run("empty role", func(t *testing.T) {
		u := &User{Role: ""}
		assert.False(t, u.IsAdmin())
	})
}

func TestQuestion_Update(t *testing.T) {
	q := &Question{Title: "Old Title"}
	newTopicID := uuid.New()
	adminID := uuid.New()

	before := time.Now().UTC().Add(-time.Second)

	q.Update(
		"New Title",
		"new-slug",
		map[string]interface{}{"text": "content"},
		QuestionLevelAdvanced,
		&newTopicID,
		true,
		&adminID,
	)

	assert.Equal(t, "New Title", q.Title)
	assert.Equal(t, "new-slug", q.Slug)
	assert.Equal(t, QuestionLevelAdvanced, q.Level)
	assert.Equal(t, &newTopicID, q.TopicID)
	assert.True(t, q.IsFree)
	assert.Equal(t, &adminID, q.UpdatedBy)
	assert.True(t, q.UpdatedAt.After(before))
}

func TestTopic_Update(t *testing.T) {
	topic := &Topic{Title: "Old"}
	adminID := uuid.New()

	before := time.Now().UTC().Add(-time.Second)

	topic.Update("New", "slug", "desc", &adminID)

	assert.Equal(t, "New", topic.Title)
	assert.Equal(t, "slug", topic.Slug)
	assert.Equal(t, "desc", topic.Description)
	assert.Equal(t, &adminID, topic.UpdatedBy)
	assert.True(t, topic.UpdatedAt.After(before))
}
