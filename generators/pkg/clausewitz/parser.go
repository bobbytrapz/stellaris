package clausewitz

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExtractTopLevelKeys reads a Stellaris script file and extracts all keys
// that define top-level blocks.
func ExtractTopLevelKeys(filePath string) ([]string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Remove UTF-8 BOM if present
	if bytes.HasPrefix(content, []byte("\xef\xbb\xbf")) {
		content = content[3:]
	}

	return parseTopLevelKeys(string(content)), nil
}

// ExtractTopLevelKeysFromDir reads all .txt files in a directory and extracts their top level keys.
func ExtractTopLevelKeysFromDir(dirPath string) ([]string, error) {
	var allKeys []string
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".txt") {
			continue
		}
		filePath := filepath.Join(dirPath, entry.Name())
		keys, err := ExtractTopLevelKeys(filePath)
		if err != nil {
			return nil, err
		}
		allKeys = append(allKeys, keys...)
	}
	return allKeys, nil
}

func parseTopLevelKeys(content string) []string {
	var keys []string
	
	// Fast path: use regex for standard definitions at depth 0
	// e.g. `trait_adaptive = {`
	// However, we have to be careful about things inside blocks.
	// Since we only care about depth 0, we should write a simple state machine.

	depth := 0
	reader := bufio.NewReader(strings.NewReader(content))
	
	// A simple regex to catch potential identifiers at the start of lines or preceded by whitespace.
	// Matches `identifier = {` or `identifier = { ... }` or `identifier = ...`
	inComment := false
	inString := false

	var tokens []string
	var currentToken strings.Builder

	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return keys
		}

		// Handle comments
		if inComment {
			if r == '\n' {
				inComment = false
			}
			continue
		}

		if r == '#' {
			inComment = true
			continue
		}

		// Handle strings (rudimentary)
		if r == '"' {
			inString = !inString
			currentToken.WriteRune(r)
			continue
		}

		if inString {
			currentToken.WriteRune(r)
			continue
		}

		switch r {
		case '{':
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			tokens = append(tokens, "{")
			depth++
		case '}':
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			tokens = append(tokens, "}")
			depth--
		case '=':
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			tokens = append(tokens, "=")
		case ' ', '\t', '\n', '\r':
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
		default:
			currentToken.WriteRune(r)
		}
	}
	if currentToken.Len() > 0 {
		tokens = append(tokens, currentToken.String())
	}

	// Now analyze tokens for depth 0 definitions: KEY = {
	currentDepth := 0
	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == "{" {
			currentDepth++
		} else if tok == "}" {
			currentDepth--
		} else if tok == "=" && currentDepth == 0 {
			// The token before `=` is the key, IF the token after `=` is `{` 
			// (some top level things might not be blocks, but in common/traits they usually are,
			// actually we want all top level keys).
			if i > 0 {
				key := tokens[i-1]
				// Avoid adding random boolean/int assignments if we only want blocks
				// We assume any top-level assignment is a key we might care about.
				// For traits/civics, they are almost always blocks.
				// Let's ensure it's not a block close brace etc
				if key != "}" && key != "{" {
					// Some files start with a @var assignment e.g. `@tier1cost = 1`
					if !strings.HasPrefix(key, "@") {
						// add key
						keys = append(keys, key)
					}
				}
			}
		}
	}

	return keys
}
