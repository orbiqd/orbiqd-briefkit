package toml

import (
	"path/filepath"

	"github.com/spf13/afero"
)

const testFilePath = "config.toml"

func newTestFile(fs afero.Fs, buffer []byte) *File {
	return &File{
		fs:       fs,
		filePath: testFilePath,
		buffer:   buffer,

		blockStart:  make(map[string]int),
		blockLength: make(map[string]int),
	}
}

func newParsedTestFile(content string) (*File, error) {
	fs := afero.NewMemMapFs()
	if err := afero.WriteFile(fs, testFilePath, []byte(content), 0644); err != nil {
		return nil, err
	}

	return Read(testFilePath, WithFs(fs))
}

func tempPathFor(filePath string) string {
	return filepath.Join(filepath.Dir(filePath), "~"+filepath.Base(filePath))
}
