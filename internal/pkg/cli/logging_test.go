package cli

import (
	"log/slog"
	"testing"

	"github.com/MatusOllah/slogcolor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateLoggerFromConfig_Success(t *testing.T) {
	tests := []struct {
		name       string
		config     LogConfig
		wantFormat string
	}{
		{
			name: "text-no-color format",
			config: LogConfig{
				Level:  "info",
				Format: "text-no-color",
			},
			wantFormat: "text-no-color",
		},
		{
			name: "text-color format",
			config: LogConfig{
				Level:  "info",
				Format: "text-color",
			},
			wantFormat: "text-color",
		},
		{
			name: "json format",
			config: LogConfig{
				Level:  "debug",
				Format: "json",
			},
			wantFormat: "json",
		},
		{
			name: "trims input",
			config: LogConfig{
				Level:  " INFO ",
				Format: " Text-No-Color ",
			},
			wantFormat: "text-no-color",
		},
		{
			name: "quiet mode",
			config: LogConfig{
				Level:  "warn",
				Format: "text-no-color",
				Quiet:  true,
			},
			wantFormat: "text-no-color",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := CreateLoggerFromConfig(tt.config)
			require.NoError(t, err)
			require.NotNil(t, logger)

			switch tt.wantFormat {
			case "text-no-color":
				_, ok := logger.Handler().(*slog.TextHandler)
				assert.True(t, ok)
			case "text-color":
				_, ok := logger.Handler().(*slogcolor.Handler)
				assert.True(t, ok)
			case "json":
				_, ok := logger.Handler().(*slog.JSONHandler)
				assert.True(t, ok)
			default:
				t.Fatalf("unknown wantFormat: %s", tt.wantFormat)
			}
		})
	}
}

func TestCreateLoggerFromConfig_Errors(t *testing.T) {
	tests := []struct {
		name          string
		config        LogConfig
		errorContains string
	}{
		{
			name: "missing level",
			config: LogConfig{
				Format: "text-no-color",
			},
			errorContains: "log level is required",
		},
		{
			name: "unknown level",
			config: LogConfig{
				Level:  "verbose",
				Format: "text-no-color",
			},
			errorContains: "unknown log level",
		},
		{
			name: "missing format",
			config: LogConfig{
				Level: "info",
			},
			errorContains: "log format is required",
		},
		{
			name: "unknown format",
			config: LogConfig{
				Level:  "info",
				Format: "xml",
			},
			errorContains: "unknown log format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := CreateLoggerFromConfig(tt.config)
			require.Error(t, err)
			assert.Nil(t, logger)
			assert.ErrorContains(t, err, tt.errorContains)
		})
	}
}
