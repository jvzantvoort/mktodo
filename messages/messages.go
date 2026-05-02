package messages

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed errors.txt
var errorsText string

//go:embed prompts.txt
var promptsText string

//go:embed help.txt
var helpText string

var (
	errors  map[string]string
	prompts map[string]string
)

func init() {
	errors = parseMessages(errorsText)
	prompts = parseMessages(promptsText)
}

// parseMessages parses a message file into a key-value map
func parseMessages(content string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(content, "\n")

	var currentKey string
	var currentValue strings.Builder

	for _, line := range lines {
		// Skip comments and empty lines at the start of a message
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}

		// Check if this is a key=value line
		if strings.Contains(line, "=") && !strings.HasPrefix(line, " ") {
			// Save previous message if exists
			if currentKey != "" {
				result[currentKey] = strings.TrimSpace(currentValue.String())
			}

			// Parse new message
			parts := strings.SplitN(line, "=", 2)
			currentKey = strings.TrimSpace(parts[0])
			currentValue.Reset()
			if len(parts) > 1 {
				currentValue.WriteString(parts[1])
			}
		} else if currentKey != "" && line != "" {
			// Continuation of previous message
			currentValue.WriteString("\n")
			currentValue.WriteString(line)
		}
	}

	// Save last message
	if currentKey != "" {
		result[currentKey] = strings.TrimSpace(currentValue.String())
	}

	return result
}

// Error returns an error message by key, with optional formatting
func Error(key string, args ...interface{}) string {
	msg, ok := errors[key]
	if !ok {
		return fmt.Sprintf("Unknown error: %s", key)
	}

	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}
	return msg
}

// Prompt returns a prompt message by key, with optional formatting
func Prompt(key string, args ...interface{}) string {
	msg, ok := prompts[key]
	if !ok {
		return fmt.Sprintf("Unknown prompt: %s", key)
	}

	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}
	return msg
}

// Get returns a message by key from any category, with optional formatting
func Get(key string, args ...interface{}) string {
	// Try errors first
	if msg, ok := errors[key]; ok {
		if len(args) > 0 {
			return fmt.Sprintf(msg, args...)
		}
		return msg
	}

	// Try prompts
	if msg, ok := prompts[key]; ok {
		if len(args) > 0 {
			return fmt.Sprintf(msg, args...)
		}
		return msg
	}

	// Not found
	return fmt.Sprintf("Unknown message: %s", key)
}

// Help returns the full help text
func Help() string {
	return helpText
}

// Errorf creates an error with a message key and formatting
func Errorf(key string, args ...interface{}) error {
	return fmt.Errorf("%s", Error(key, args...))
}

// Icon returns an icon/emoji by key
func Icon(key string) string {
	icons := map[string]string{
		"checked":   prompts["ICON_CHECKED"],
		"unchecked": prompts["ICON_UNCHECKED"],
		"fixme":     prompts["ICON_FIXME"],
		"celebrate": prompts["ICON_CELEBRATE"],
		"warning":   prompts["ICON_WARNING"],
	}

	if icon, ok := icons[key]; ok {
		return icon
	}
	return ""
}

// HasError checks if an error key exists
func HasError(key string) bool {
	_, ok := errors[key]
	return ok
}

// HasPrompt checks if a prompt key exists
func HasPrompt(key string) bool {
	_, ok := prompts[key]
	return ok
}

// AllErrors returns all error keys (useful for testing)
func AllErrors() []string {
	keys := make([]string, 0, len(errors))
	for k := range errors {
		keys = append(keys, k)
	}
	return keys
}

// AllPrompts returns all prompt keys (useful for testing)
func AllPrompts() []string {
	keys := make([]string, 0, len(prompts))
	for k := range prompts {
		keys = append(keys, k)
	}
	return keys
}
