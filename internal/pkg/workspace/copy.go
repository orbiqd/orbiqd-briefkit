package workspace

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

// CopyDir recursively copies the directory at src into dst using the provided filesystem.
// Symlinks are preserved as symlinks (not followed) to prevent workspace escape.
// On failure, any partially created files at dst are left for the caller to clean up.
func CopyDir(fs afero.Fs, src, dst string) error {
	return afero.Walk(fs, src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk %s: %w", path, err)
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("relative path: %w", err)
		}
		target := filepath.Join(dst, rel)

		if info.Mode()&os.ModeSymlink != 0 {
			return copySymlink(path, target)
		}

		if info.IsDir() {
			return fs.MkdirAll(target, info.Mode())
		}

		return copyFile(fs, path, target, info.Mode())
	})
}

func copyFile(fs afero.Fs, src, dst string, mode os.FileMode) error {
	srcFile, err := fs.Open(src)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer func() { _ = srcFile.Close() }()

	dstFile, err := fs.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("create destination file: %w", err)
	}
	defer func() { _ = dstFile.Close() }()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copy file contents: %w", err)
	}

	return nil
}

func copySymlink(src, dst string) error {
	// afero does not expose symlink APIs, so fall back to os for symlink handling.
	target, err := os.Readlink(src)
	if err != nil {
		return fmt.Errorf("read symlink: %w", err)
	}

	// Rewrite relative symlinks so they remain valid inside the copied tree.
	if !filepath.IsAbs(target) {
		srcDir := filepath.Dir(src)
		absTarget := filepath.Join(srcDir, target)
		rel, err := filepath.Rel(filepath.Dir(dst), absTarget)
		if err != nil {
			return fmt.Errorf("rewrite symlink target: %w", err)
		}
		target = rel
	}

	return os.Symlink(target, dst)
}
