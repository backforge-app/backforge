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

func TestUser_PasswordManagement(t *testing.T) {
	t.Run("Set and Check Password Success", func(t *testing.T) {
		u := &User{}
		err := u.SetPassword("my_secure_password")

		assert.NoError(t, err)
		assert.NotNil(t, u.PasswordHash)
		assert.True(t, u.HasPassword())

		assert.True(t, u.CheckPassword("my_secure_password"))
		assert.False(t, u.CheckPassword("wrong_password"))
	})

	t.Run("Password too long", func(t *testing.T) {
		u := &User{}
		longPassword := string(make([]byte, 73))

		err := u.SetPassword(longPassword)

		assert.ErrorIs(t, err, ErrPasswordTooLong)
		assert.Nil(t, u.PasswordHash)
		assert.False(t, u.HasPassword())
	})

	t.Run("CheckPassword without hash", func(t *testing.T) {
		u := &User{PasswordHash: nil} // Имитация OAuth пользователя

		assert.False(t, u.HasPassword())
		assert.False(t, u.CheckPassword("any_password"))
	})
}

func TestUser_Constructors(t *testing.T) {
	t.Run("NewUserWithPassword", func(t *testing.T) {
		firstName := "John"

		u, err := NewUserWithPassword("test@example.com", "valid_pass", firstName, nil, nil, nil)

		assert.NoError(t, err)
		assert.Equal(t, "test@example.com", u.Email)
		assert.Equal(t, "John", u.FirstName)
		assert.Equal(t, UserRoleUser, u.Role)
		assert.False(t, u.IsEmailVerified)
		assert.True(t, u.HasPassword())
	})

	t.Run("NewUserFromOAuth", func(t *testing.T) {
		firstName := "OAuth"

		u := NewUserFromOAuth("oauth@example.com", firstName, nil, nil, nil, true)

		assert.Equal(t, "oauth@example.com", u.Email)
		assert.Equal(t, "OAuth", u.FirstName)
		assert.Equal(t, UserRoleUser, u.Role)
		assert.True(t, u.IsEmailVerified)
		assert.False(t, u.HasPassword())
		assert.Nil(t, u.PasswordHash)
	})
}

func TestVerificationToken(t *testing.T) {
	t.Run("NewVerificationToken creates valid token and hash", func(t *testing.T) {
		userID := uuid.New()
		ttl := 1 * time.Hour

		rawToken, entity, err := NewVerificationToken(userID, TokenPurposeEmailVerification, ttl)

		assert.NoError(t, err)
		assert.NotEmpty(t, rawToken)
		assert.NotNil(t, entity)

		expectedHash := HashVerificationToken(rawToken)
		assert.Equal(t, expectedHash, entity.TokenHash)

		assert.Equal(t, userID, entity.UserID)
		assert.Equal(t, TokenPurposeEmailVerification, entity.Purpose)
		assert.False(t, entity.IsExpired())
	})

	t.Run("IsExpired returns correct status", func(t *testing.T) {
		expiredToken := &VerificationToken{ExpiresAt: time.Now().Add(-1 * time.Minute)}
		assert.True(t, expiredToken.IsExpired())

		validToken := &VerificationToken{ExpiresAt: time.Now().Add(1 * time.Minute)}
		assert.False(t, validToken.IsExpired())
	})

	t.Run("HashVerificationToken is deterministic", func(t *testing.T) {
		token := "my_secret_token_string"

		hash1 := HashVerificationToken(token)
		hash2 := HashVerificationToken(token)

		assert.Equal(t, hash1, hash2, "Hashing the same token should produce the same output")
		assert.NotEqual(t, token, hash1, "Hash should not equal the plain text token")
	})
}
