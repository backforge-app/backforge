package render

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type validationTestStruct struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Age      int    `json:"age" validate:"min=18,max=99"`
	UserType string `json:"user_type" validate:"oneof=admin user guest"`
	ID       string `json:"id" validate:"uuid"`
	Website  string `json:"website" validate:"url"`
}

func TestValidate_Success(t *testing.T) {
	req := validationTestStruct{
		Name:     "Alice",
		Email:    "alice@example.com",
		Age:      25,
		UserType: "admin",
		ID:       "550e8400-e29b-41d4-a716-446655440000",
		Website:  "https://example.com",
	}

	err := Validate(req)
	require.NoError(t, err)
}

func TestValidate_Failures(t *testing.T) {
	req := validationTestStruct{
		Name:     "",
		Email:    "invalid-email",
		Age:      17,
		UserType: "superuser",
		ID:       "not-a-uuid",
		Website:  "invalid-url",
	}

	err := Validate(req)
	require.Error(t, err)

	validationErrs := ValidationErrors(err)
	expected := map[string]string{
		"name":      "field is required",
		"email":     "must be a valid email",
		"age":       "must be at least 18",
		"user_type": "must be one of: admin user guest",
		"id":        "must be a valid UUID",
		"website":   "must be a valid URL",
	}

	require.Equal(t, expected, validationErrs)
}

func TestValidationErrors_NonValidationError(t *testing.T) {
	err := ValidationErrors(nil)
	require.Empty(t, err)

	err = ValidationErrors(errors.New("dummy error"))
	require.Empty(t, err)
}

func TestValidationErrors_MinMaxString(t *testing.T) {
	type stringTest struct {
		Password string `json:"password" validate:"min=8,max=16"`
	}

	req := stringTest{Password: "short"}
	err := Validate(req)
	require.Error(t, err)

	validationErrs := ValidationErrors(err)
	require.Equal(t, map[string]string{"password": "must be at least 8 characters"}, validationErrs)

	req.Password = "thispasswordistoolong123"
	err = Validate(req)
	require.Error(t, err)
	validationErrs = ValidationErrors(err)
	require.Equal(t, map[string]string{"password": "must be at most 16 characters"}, validationErrs)
}
