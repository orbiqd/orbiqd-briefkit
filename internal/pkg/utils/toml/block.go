package toml

import (
	"bytes"
	"errors"
	"strings"
)

// Block provides access to a named TOML block for in-place edits.
type Block struct {
	file      *File
	blockName string
}

// SetKey sets or updates a key-value pair in the block.
func (block *Block) SetKey(key string, value any) {
	if block == nil || block.file == nil {
		return
	}

	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return
	}

	start, length, ok := block.blockRange()
	if !ok {
		if block.blockName == MainBlock {
			start = 0
			length = len(block.file.buffer)
			block.file.blockStart[MainBlock] = 0
			block.file.blockLength[MainBlock] = length
		} else {
			block.appendBlockWithKey(trimmedKey, value)
			return
		}
	}

	blockBytes := block.file.buffer[start : start+length]
	lineStart, lineContentEnd, _, _, found := findKeyLine(blockBytes, trimmedKey)
	newLine := formatKeyValue(trimmedKey, value)

	if found {
		block.updateBuffer(start+lineStart, start+lineContentEnd, []byte(newLine), start)
		return
	}

	newline := detectNewline(block.file.buffer)
	prefix := ""
	if length > 0 && !bytes.HasSuffix(blockBytes, []byte(newline)) {
		prefix = newline
	}
	insertion := []byte(prefix + newLine + newline)
	block.updateBuffer(start+length, start+length, insertion, start)
}

// GetKey returns the parsed value for a key in the block.
func (block *Block) GetKey(key string) (any, error) {
	if block == nil || block.file == nil {
		return nil, ErrBlockNotFound
	}

	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return nil, ErrKeyNotFound
	}

	start, length, ok := block.blockRange()
	if !ok {
		return nil, ErrBlockNotFound
	}

	blockBytes := block.file.buffer[start : start+length]
	_, _, _, rawValue, found := findKeyLine(blockBytes, trimmedKey)
	if !found {
		return nil, ErrKeyNotFound
	}

	return parseValue(rawValue), nil
}

// UnsetKey removes a key from the block.
func (block *Block) UnsetKey(key string) error {
	if block == nil || block.file == nil {
		return ErrBlockNotFound
	}

	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return ErrKeyNotFound
	}

	start, length, ok := block.blockRange()
	if !ok {
		return ErrBlockNotFound
	}

	blockBytes := block.file.buffer[start : start+length]
	lineStart, _, lineEndWithNewline, _, found := findKeyLine(blockBytes, trimmedKey)
	if !found {
		return ErrKeyNotFound
	}

	block.updateBuffer(start+lineStart, start+lineEndWithNewline, nil, start)
	return nil
}

// HasKey reports whether the block contains the key.
func (block *Block) HasKey(key string) bool {
	if block == nil || block.file == nil {
		return false
	}

	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return false
	}

	start, length, ok := block.blockRange()
	if !ok {
		return false
	}

	blockBytes := block.file.buffer[start : start+length]
	_, _, _, _, found := findKeyLine(blockBytes, trimmedKey)
	return found
}

func (block *Block) blockRange() (int, int, bool) {
	start, ok := block.file.blockStart[block.blockName]
	if !ok {
		return 0, 0, false
	}

	length, ok := block.file.blockLength[block.blockName]
	if !ok {
		return 0, 0, false
	}

	return start, length, true
}

func (block *Block) appendBlockWithKey(key string, value any) {
	newline := detectNewline(block.file.buffer)
	header := "[" + block.blockName + "]"
	newLine := formatKeyValue(key, value)

	oldLen := len(block.file.buffer)
	prefix := ""
	if oldLen > 0 && !bytes.HasSuffix(block.file.buffer, []byte(newline)) {
		prefix = newline
	}

	insertion := []byte(prefix + header + newline + newLine + newline)
	block.file.buffer = append(block.file.buffer, insertion...)

	if oldLen > 0 && prefix != "" {
		if lastName, ok := block.lastBlockName(); ok {
			block.file.blockLength[lastName] += len(prefix)
		}
	}

	contentStart := oldLen + len(prefix) + len(header) + len(newline)
	block.file.blockStart[block.blockName] = contentStart
	block.file.blockLength[block.blockName] = len(newLine) + len(newline)
}

func (block *Block) lastBlockName() (string, bool) {
	if len(block.file.blockStart) == 0 {
		return "", false
	}

	lastName := ""
	lastStart := -1
	for name, start := range block.file.blockStart {
		if start > lastStart {
			lastStart = start
			lastName = name
		}
	}

	return lastName, true
}

func (block *Block) updateBuffer(start int, end int, replacement []byte, blockStart int) {
	delta := len(replacement) - (end - start)
	block.file.buffer = replaceRange(block.file.buffer, start, end, replacement)
	if delta == 0 {
		return
	}

	block.file.blockLength[block.blockName] += delta
	block.shiftBlocksAfter(blockStart, delta)
}

func (block *Block) shiftBlocksAfter(blockStart int, delta int) {
	if delta == 0 {
		return
	}

	for name, start := range block.file.blockStart {
		if name == block.blockName {
			continue
		}
		if start > blockStart {
			block.file.blockStart[name] = start + delta
		}
	}
}

// ErrKeyNotFound indicates a requested key was not found in the block.
var ErrKeyNotFound = errors.New("key not found")
