package toml

import (
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRead_WhenFileMissing_ThenReturnsError(t *testing.T) {
	fs := afero.NewMemMapFs()

	file, err := Read("config.toml", WithFs(fs))

	require.Error(t, err)
	assert.Nil(t, file)
	assert.True(t, os.IsNotExist(err))
}

func TestRead_WhenFileEmpty_ThenInitializesMainBlock(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "config.toml"

	require.NoError(t, afero.WriteFile(fs, filePath, []byte(""), 0644))

	file, err := Read(filePath, WithFs(fs))

	require.NoError(t, err)
	require.NotNil(t, file)
	assert.Empty(t, file.buffer)
	assert.Len(t, file.blockStart, 1)
	assert.Len(t, file.blockLength, 1)
	assert.Equal(t, 0, file.blockStart[MainBlock])
	assert.Equal(t, 0, file.blockLength[MainBlock])
}

func TestRead_WhenFileHasMainAndSectionBlocks_ThenParsesBlockOffsets(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "config.toml"
	content := "name = \"demo\"\n[section]\nvalue = 1\n"

	require.NoError(t, afero.WriteFile(fs, filePath, []byte(content), 0644))

	file, err := Read(filePath, WithFs(fs))

	require.NoError(t, err)
	require.NotNil(t, file)

	expectedMainLength := len("name = \"demo\"\n")
	expectedSectionStart := len("name = \"demo\"\n[section]\n")
	expectedSectionLength := len("value = 1\n")

	assert.Equal(t, []byte(content), file.buffer)
	assert.Len(t, file.blockStart, 2)
	assert.Len(t, file.blockLength, 2)
	assert.Equal(t, 0, file.blockStart[MainBlock])
	assert.Equal(t, expectedMainLength, file.blockLength[MainBlock])
	assert.Equal(t, expectedSectionStart, file.blockStart["section"])
	assert.Equal(t, expectedSectionLength, file.blockLength["section"])
}

func TestRead_WhenUsingCustomFs_ThenReadsFromProvidedFs(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "config.toml"
	content := "key = 1\n"

	require.NoError(t, afero.WriteFile(fs, filePath, []byte(content), 0644))

	file, err := Read(filePath, WithFs(fs))

	require.NoError(t, err)
	require.NotNil(t, file)
	assert.Equal(t, fs, file.fs)
}

func TestRead_WhenLineIsCommentOnly_ThenDoesNotCreateBlock(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "config.toml"
	content := "# comment\n[section]\nvalue = 1\n"

	require.NoError(t, afero.WriteFile(fs, filePath, []byte(content), 0644))

	file, err := Read(filePath, WithFs(fs))

	require.NoError(t, err)
	require.NotNil(t, file)

	expectedMainLength := len("# comment\n")
	expectedSectionStart := len("# comment\n[section]\n")
	expectedSectionLength := len("value = 1\n")

	assert.Len(t, file.blockStart, 2)
	assert.Len(t, file.blockLength, 2)
	assert.Equal(t, 0, file.blockStart[MainBlock])
	assert.Equal(t, expectedMainLength, file.blockLength[MainBlock])
	assert.Equal(t, expectedSectionStart, file.blockStart["section"])
	assert.Equal(t, expectedSectionLength, file.blockLength["section"])
}

func TestRead_WhenHeaderHasInlineComment_ThenParsesBlockName(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "config.toml"
	content := "key = 1\n[section] # note\nvalue = 1\n"

	require.NoError(t, afero.WriteFile(fs, filePath, []byte(content), 0644))

	file, err := Read(filePath, WithFs(fs))

	require.NoError(t, err)
	require.NotNil(t, file)

	expectedMainLength := len("key = 1\n")
	expectedSectionStart := len("key = 1\n[section] # note\n")
	expectedSectionLength := len("value = 1\n")

	assert.Len(t, file.blockStart, 2)
	assert.Len(t, file.blockLength, 2)
	assert.Equal(t, 0, file.blockStart[MainBlock])
	assert.Equal(t, expectedMainLength, file.blockLength[MainBlock])
	assert.Equal(t, expectedSectionStart, file.blockStart["section"])
	assert.Equal(t, expectedSectionLength, file.blockLength["section"])
}

func TestRead_WhenHeaderIsCommentedOut_ThenDoesNotCreateBlock(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "config.toml"
	content := "key = 1\n# [section]\nvalue = 1\n"

	require.NoError(t, afero.WriteFile(fs, filePath, []byte(content), 0644))

	file, err := Read(filePath, WithFs(fs))

	require.NoError(t, err)
	require.NotNil(t, file)

	assert.Len(t, file.blockStart, 1)
	assert.Len(t, file.blockLength, 1)
	assert.Equal(t, 0, file.blockStart[MainBlock])
	assert.Equal(t, len(content), file.blockLength[MainBlock])
}

func TestRead_WhenHeaderHasTrailingTextBeforeComment_ThenDoesNotCreateBlock(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "config.toml"
	content := "key = 1\n[section] trailing # note\nvalue = 1\n"

	require.NoError(t, afero.WriteFile(fs, filePath, []byte(content), 0644))

	file, err := Read(filePath, WithFs(fs))

	require.NoError(t, err)
	require.NotNil(t, file)

	assert.Len(t, file.blockStart, 1)
	assert.Len(t, file.blockLength, 1)
	assert.Equal(t, 0, file.blockStart[MainBlock])
	assert.Equal(t, len(content), file.blockLength[MainBlock])
}

func TestRead_WhenHeaderMissingClosingBracket_ThenDoesNotCreateBlock(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "config.toml"
	content := "[section\nkey = 1\n"

	require.NoError(t, afero.WriteFile(fs, filePath, []byte(content), 0644))

	file, err := Read(filePath, WithFs(fs))

	require.NoError(t, err)
	require.NotNil(t, file)

	assert.Len(t, file.blockStart, 1)
	assert.Len(t, file.blockLength, 1)
	assert.Equal(t, 0, file.blockStart[MainBlock])
	assert.Equal(t, len(content), file.blockLength[MainBlock])
}

func TestRead_WhenHeaderHasEmptyName_ThenDoesNotCreateBlock(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "config.toml"
	content := "[]\nkey = 1\n"

	require.NoError(t, afero.WriteFile(fs, filePath, []byte(content), 0644))

	file, err := Read(filePath, WithFs(fs))

	require.NoError(t, err)
	require.NotNil(t, file)

	assert.Len(t, file.blockStart, 1)
	assert.Len(t, file.blockLength, 1)
	assert.Equal(t, 0, file.blockStart[MainBlock])
	assert.Equal(t, len(content), file.blockLength[MainBlock])
}

func TestRead_WhenFileUsesCRLF_ThenParsesBlocks(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "config.toml"
	content := "key = 1\r\n[section]\r\nvalue = 1\r\n"

	require.NoError(t, afero.WriteFile(fs, filePath, []byte(content), 0644))

	file, err := Read(filePath, WithFs(fs))

	require.NoError(t, err)
	require.NotNil(t, file)

	expectedMainLength := len("key = 1\r\n")
	expectedSectionStart := len("key = 1\r\n[section]\r\n")
	expectedSectionLength := len("value = 1\r\n")

	assert.Len(t, file.blockStart, 2)
	assert.Len(t, file.blockLength, 2)
	assert.Equal(t, 0, file.blockStart[MainBlock])
	assert.Equal(t, expectedMainLength, file.blockLength[MainBlock])
	assert.Equal(t, expectedSectionStart, file.blockStart["section"])
	assert.Equal(t, expectedSectionLength, file.blockLength["section"])
}

func TestWrite_WhenFileExists_ThenOverwritesWithSameMode(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "config.toml"
	existingMode := os.FileMode(0600)

	require.NoError(t, afero.WriteFile(fs, filePath, []byte("old = 1\n"), existingMode))

	file := newTestFile(fs, []byte("new = 2\n"))

	err := file.Write()

	require.NoError(t, err)

	updated, err := afero.ReadFile(fs, filePath)
	require.NoError(t, err)
	assert.Equal(t, "new = 2\n", string(updated))

	info, err := fs.Stat(filePath)
	require.NoError(t, err)
	assert.Equal(t, existingMode, info.Mode().Perm())
}

func TestWrite_WhenFileMissing_ThenCreatesFileWithDefaultMode(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "config.toml"

	file := newTestFile(fs, []byte("key = 1\n"))

	err := file.Write()

	require.NoError(t, err)

	updated, err := afero.ReadFile(fs, filePath)
	require.NoError(t, err)
	assert.Equal(t, "key = 1\n", string(updated))

	info, err := fs.Stat(filePath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0644), info.Mode().Perm())
}

func TestWrite_WhenTempFileExists_ThenReturnsError(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "config.toml"
	tmpPath := tempPathFor(filePath)

	require.NoError(t, afero.WriteFile(fs, tmpPath, []byte("tmp"), 0644))

	file := newTestFile(fs, []byte("key = 1\n"))

	err := file.Write()

	require.Error(t, err)
	require.ErrorContains(t, err, "temp file already exists")
	require.ErrorIs(t, err, os.ErrExist)

	exists, existsErr := afero.Exists(fs, filePath)
	require.NoError(t, existsErr)
	assert.False(t, exists)
}

func TestWrite_WhenTempDirMissing_ThenReturnsError(t *testing.T) {
	baseFs := afero.NewMemMapFs()
	filePath := "config.toml"

	require.NoError(t, afero.WriteFile(baseFs, filePath, []byte("old = 1\n"), 0644))

	fs := afero.NewReadOnlyFs(baseFs)
	file := newTestFile(fs, []byte("key = 1\n"))

	err := file.Write()

	require.Error(t, err)
	assert.ErrorContains(t, err, "temp file write")
}

func TestOpenBlock_WhenCalled_ThenReturnsBlockViewWithFileAndName(t *testing.T) {
	file := &File{filePath: "config.toml"}

	block := file.OpenBlock("section")

	require.NotNil(t, block)
	assert.Equal(t, file, block.file)
	assert.Equal(t, "section", block.blockName)
}
