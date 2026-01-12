package toml

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

// MainBlock identifies the content before the first table header.
const MainBlock = ""

type FileOpt func(*File)

// File holds raw TOML content and block indexes to allow in-place edits.
type File struct {
	fs       afero.Fs
	filePath string

	buffer      []byte
	blockStart  map[string]int
	blockLength map[string]int
}

// WithFs overrides the filesystem used for TOML file operations.
func WithFs(fs afero.Fs) FileOpt {
	return func(f *File) {
		f.fs = fs
	}
}

// Read loads the TOML file into memory.
func Read(filePath string, opts ...FileOpt) (*File, error) {
	file := &File{
		fs:       afero.NewOsFs(),
		filePath: filePath,

		blockStart:  make(map[string]int),
		blockLength: make(map[string]int),
		buffer:      make([]byte, 0),
	}

	for _, opt := range opts {
		opt(file)
	}

	err := file.read()
	if err != nil {
		return nil, err
	}

	return file, nil
}

// Write persists the in-memory buffer back to the original file.
func (file *File) Write() error {
	tmpPath := filepath.Join(filepath.Dir(file.filePath), "~"+filepath.Base(file.filePath))

	exists, err := afero.Exists(file.fs, tmpPath)
	if err != nil {
		return fmt.Errorf("temp file existence check: %w", err)
	}
	if exists {
		return fmt.Errorf("temp file already exists: %s: %w", tmpPath, os.ErrExist)
	}

	mode := os.FileMode(0644)
	info, err := file.fs.Stat(file.filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("file mode lookup: %w", err)
	}
	if err == nil {
		mode = info.Mode()
	}

	if err := afero.WriteFile(file.fs, tmpPath, file.buffer, mode); err != nil {
		return fmt.Errorf("temp file write: %w", err)
	}

	if err := file.fs.Rename(tmpPath, file.filePath); err != nil {
		_ = file.fs.Remove(tmpPath)
		return fmt.Errorf("temp file rename: %w", err)
	}

	return nil
}

// OpenBlock returns a block view by name for raw byte edits.
func (file *File) OpenBlock(blockName string) *Block {
	return &Block{
		file:      file,
		blockName: blockName,
	}
}

func (file *File) read() error {
	buffer, err := afero.ReadFile(file.fs, file.filePath)
	if err != nil {
		return err
	}

	file.buffer = buffer

	return file.parse()
}

func (file *File) parse() error {
	file.blockStart[MainBlock] = 0
	if len(file.buffer) == 0 {
		file.blockLength[MainBlock] = 0
		return nil
	}

	currentBlock := MainBlock
	currentBlockStart := 0
	lineStart := 0

	for i := 0; i <= len(file.buffer); i++ {
		if i == len(file.buffer) || file.buffer[i] == '\n' {
			line := file.buffer[lineStart:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}

			blockName, ok := parseBlockName(line)
			if ok {
				file.blockLength[currentBlock] = lineStart - currentBlockStart

				nextStart := i
				if i < len(file.buffer) {
					nextStart = i + 1
				}

				currentBlock = blockName
				file.blockStart[currentBlock] = nextStart
				currentBlockStart = nextStart
			}

			lineStart = i + 1
		}
	}

	file.blockLength[currentBlock] = len(file.buffer) - currentBlockStart

	return nil
}

// ErrBlockNotFound indicates a requested block was not found in the file.
var ErrBlockNotFound = errors.New("block not found")
