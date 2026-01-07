package agent

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/utils"
	"github.com/stretchr/testify/require"
)

func TestNewExecutionID(t *testing.T) {
	id := NewExecutionID()
	require.NotEmpty(t, id)
	_, err := uuid.Parse(string(id))
	require.NoError(t, err)

	other := NewExecutionID()
	require.NotEqual(t, id, other)
}

func TestExecutionIDValidate(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		err := ExecutionID("").Validate()
		require.ErrorIs(t, err, ErrExecutionIDInvalid)
	})

	t.Run("invalid", func(t *testing.T) {
		err := ExecutionID("not-a-uuid").Validate()
		require.ErrorIs(t, err, ErrExecutionIDInvalid)
	})

	t.Run("valid", func(t *testing.T) {
		err := ExecutionID(uuid.NewString()).Validate()
		require.NoError(t, err)
	})
}

func TestExecutionInputAttachmentValidate(t *testing.T) {
	t.Run("missing mime type", func(t *testing.T) {
		err := ExecutionInputAttachment{Path: "/tmp/file.png"}.Validate()
		require.ErrorIs(t, err, ErrExecutionAttachmentMimeTypeRequired)
	})

	t.Run("missing path", func(t *testing.T) {
		err := ExecutionInputAttachment{MimeType: "image/png"}.Validate()
		require.ErrorIs(t, err, ErrExecutionAttachmentPathRequired)
	})

	t.Run("valid", func(t *testing.T) {
		err := ExecutionInputAttachment{MimeType: "image/png", Path: "/tmp/file.png"}.Validate()
		require.NoError(t, err)
	})
}

func TestExecutionInputValidate(t *testing.T) {
	workingDir := t.TempDir()
	valid := ExecutionInput{
		Prompt:           "Hello",
		Timeout:          utils.Duration(5 * time.Second),
		WorkingDirectory: &workingDir,
	}

	t.Run("missing prompt", func(t *testing.T) {
		input := valid
		input.Prompt = "  "
		err := input.Validate()
		require.ErrorIs(t, err, ErrExecutionPromptRequired)
	})

	t.Run("missing timeout", func(t *testing.T) {
		input := valid
		input.Timeout = 0
		err := input.Validate()
		require.ErrorIs(t, err, ErrExecutionTimeoutRequired)
	})

	t.Run("missing working directory", func(t *testing.T) {
		input := valid
		input.WorkingDirectory = nil
		err := input.Validate()
		require.NoError(t, err)
	})

	t.Run("empty working directory", func(t *testing.T) {
		input := valid
		empty := " "
		input.WorkingDirectory = &empty
		err := input.Validate()
		require.ErrorIs(t, err, ErrExecutionWorkingDirectoryRequired)
	})

	t.Run("relative working directory", func(t *testing.T) {
		input := valid
		relative := "relative/path"
		input.WorkingDirectory = &relative
		err := input.Validate()
		require.ErrorIs(t, err, ErrExecutionWorkingDirectoryNotAbsolute)
	})

	t.Run("invalid attachment", func(t *testing.T) {
		input := valid
		input.Attachments = []ExecutionInputAttachment{{Path: "/tmp/file.png"}}
		err := input.Validate()
		require.ErrorIs(t, err, ErrExecutionAttachmentMimeTypeRequired)
	})

	t.Run("tilde working directory", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		input := valid
		tilde := "~"
		input.WorkingDirectory = &tilde
		err := input.Validate()
		require.NoError(t, err)
	})

	t.Run("valid", func(t *testing.T) {
		err := valid.Validate()
		require.NoError(t, err)
	})
}
