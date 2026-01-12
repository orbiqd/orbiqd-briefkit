package toml

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

func replaceRange(buffer []byte, start int, end int, replacement []byte) []byte {
	updated := make([]byte, 0, len(buffer)-(end-start)+len(replacement))
	updated = append(updated, buffer[:start]...)
	updated = append(updated, replacement...)
	updated = append(updated, buffer[end:]...)
	return updated
}

func detectNewline(buffer []byte) string {
	if bytes.Contains(buffer, []byte("\r\n")) {
		return "\r\n"
	}

	return "\n"
}

func formatKeyValue(key string, value any) string {
	return fmt.Sprintf("%s = %s", key, formatValue(value))
}

func formatValue(value any) string {
	switch typed := value.(type) {
	case string:
		return strconv.Quote(typed)
	case bool:
		if typed {
			return "true"
		}
		return "false"
	case int:
		return strconv.FormatInt(int64(typed), 10)
	case int8:
		return strconv.FormatInt(int64(typed), 10)
	case int16:
		return strconv.FormatInt(int64(typed), 10)
	case int32:
		return strconv.FormatInt(int64(typed), 10)
	case int64:
		return strconv.FormatInt(typed, 10)
	case uint:
		return strconv.FormatUint(uint64(typed), 10)
	case uint8:
		return strconv.FormatUint(uint64(typed), 10)
	case uint16:
		return strconv.FormatUint(uint64(typed), 10)
	case uint32:
		return strconv.FormatUint(uint64(typed), 10)
	case uint64:
		return strconv.FormatUint(typed, 10)
	case float32:
		return strconv.FormatFloat(float64(typed), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return fmt.Sprint(value)
	}
}

func findKeyLine(blockBytes []byte, key string) (int, int, int, string, bool) {
	lineStart := 0
	for i := 0; i <= len(blockBytes); i++ {
		if i == len(blockBytes) || blockBytes[i] == '\n' {
			lineEnd := i
			lineEndWithNewline := i
			if i < len(blockBytes) {
				lineEndWithNewline = i + 1
			}

			line := blockBytes[lineStart:lineEnd]
			contentEnd := lineEnd
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
				contentEnd = lineEnd - 1
			}

			parsedKey, parsedValue, ok := parseKeyLine(line)
			if ok && parsedKey == key {
				return lineStart, contentEnd, lineEndWithNewline, parsedValue, true
			}

			lineStart = i + 1
		}
	}

	return 0, 0, 0, "", false
}

func parseKeyLine(line []byte) (string, string, bool) {
	trimmed := strings.TrimSpace(string(line))
	if trimmed == "" {
		return "", "", false
	}

	if commentIndex := strings.Index(trimmed, "#"); commentIndex >= 0 {
		trimmed = strings.TrimSpace(trimmed[:commentIndex])
		if trimmed == "" {
			return "", "", false
		}
	}

	eqIndex := strings.Index(trimmed, "=")
	if eqIndex < 0 {
		return "", "", false
	}

	key := strings.TrimSpace(trimmed[:eqIndex])
	if key == "" {
		return "", "", false
	}

	value := strings.TrimSpace(trimmed[eqIndex+1:])
	return key, value, true
}

func parseValue(raw string) any {
	if raw == "" {
		return ""
	}

	if len(raw) >= 2 && raw[0] == '"' && raw[len(raw)-1] == '"' {
		unquoted, err := strconv.Unquote(raw)
		if err == nil {
			return unquoted
		}
		return raw[1 : len(raw)-1]
	}

	if len(raw) >= 2 && raw[0] == '\'' && raw[len(raw)-1] == '\'' {
		return raw[1 : len(raw)-1]
	}

	if strings.EqualFold(raw, "true") {
		return true
	}
	if strings.EqualFold(raw, "false") {
		return false
	}

	if intValue, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return intValue
	}

	if floatValue, err := strconv.ParseFloat(raw, 64); err == nil {
		return floatValue
	}

	return raw
}

func parseBlockName(line []byte) (string, bool) {
	trimmed := strings.TrimSpace(string(line))
	if trimmed == "" {
		return "", false
	}

	if commentIndex := strings.Index(trimmed, "#"); commentIndex >= 0 {
		trimmed = strings.TrimSpace(trimmed[:commentIndex])
		if trimmed == "" {
			return "", false
		}
	}

	if strings.HasPrefix(trimmed, "[") {
		end := strings.Index(trimmed, "]")
		if end < 0 {
			return "", false
		}

		name := strings.TrimSpace(trimmed[1:end])
		if name == "" {
			return "", false
		}

		if strings.TrimSpace(trimmed[end+1:]) != "" {
			return "", false
		}

		return name, true
	}

	return "", false
}
