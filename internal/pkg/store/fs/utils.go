package fs

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/orbiqd/orbiqd-briefkit/pkg/briefkit"
	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"
)

func writeJSON(fs afero.Fs, filePath string, data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("write json: failed to marshal for %s: %w", filePath, err)
	}

	tmpPath := filePath + "~"
	exists, err := afero.Exists(fs, tmpPath)
	if err != nil {
		return fmt.Errorf("write json: failed to check temp file %s: %w", tmpPath, err)
	}
	if exists {
		return fmt.Errorf("write json: temp file %s already exists: %w", tmpPath, os.ErrExist)
	}

	if err := afero.WriteFile(fs, tmpPath, b, 0644); err != nil {
		return fmt.Errorf("write json: failed to write temp file %s: %w", tmpPath, err)
	}

	if err := fs.Rename(tmpPath, filePath); err != nil {
		return fmt.Errorf("write json: failed to rename %s to %s: %w", tmpPath, filePath, err)
	}

	return nil
}

func writeYAML(fs afero.Fs, filePath string, data interface{}) error {
	b, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("write yaml: failed to marshal for %s: %w", filePath, err)
	}

	tmpPath := filePath + "~"
	exists, err := afero.Exists(fs, tmpPath)
	if err != nil {
		return fmt.Errorf("write yaml: failed to check temp file %s: %w", tmpPath, err)
	}
	if exists {
		return fmt.Errorf("write yaml: temp file %s already exists: %w", tmpPath, os.ErrExist)
	}

	if err := afero.WriteFile(fs, tmpPath, b, 0644); err != nil {
		return fmt.Errorf("write yaml: failed to write temp file %s: %w", tmpPath, err)
	}

	if err := fs.Rename(tmpPath, filePath); err != nil {
		return fmt.Errorf("write yaml: failed to rename %s to %s: %w", tmpPath, filePath, err)
	}

	return nil
}

func readJSON[T any](fs afero.Fs, filePath string) (T, error) {
	var result T
	b, err := afero.ReadFile(fs, filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return result, fmt.Errorf("read json: %w", briefkit.ErrExecutionNotFound)
		}
		return result, fmt.Errorf("read json: failed to read file %s: %w", filePath, err)
	}

	if err := json.Unmarshal(b, &result); err != nil {
		return result, fmt.Errorf("read json: failed to unmarshal from %s: %w", filePath, err)
	}
	return result, nil
}

func readYAML[T any](fs afero.Fs, filePath string) (T, error) {
	var result T
	b, err := afero.ReadFile(fs, filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return result, fmt.Errorf("read yaml: %w", briefkit.ErrAgentConfigNotFound)
		}
		return result, fmt.Errorf("read yaml: failed to read file %s: %w", filePath, err)
	}

	if err := yaml.Unmarshal(b, &result); err != nil {
		return result, fmt.Errorf("read yaml: failed to unmarshal from %s: %w", filePath, err)
	}
	return result, nil
}

func hasJSON(fs afero.Fs, filePath string) (bool, error) {
	return afero.Exists(fs, filePath)
}
