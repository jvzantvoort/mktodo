package messages

import (
	"strings"
	"testing"
)

func TestParseMessages(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected map[string]string
	}{
		{
			name: "simple key-value",
			content: `KEY1=Value 1
KEY2=Value 2`,
			expected: map[string]string{
				"KEY1": "Value 1",
				"KEY2": "Value 2",
			},
		},
		{
			name: "multiline values",
			content: `KEY1=Line 1
Line 2
Line 3
KEY2=Single line`,
			expected: map[string]string{
				"KEY1": "Line 1\nLine 2\nLine 3",
				"KEY2": "Single line",
			},
		},
		{
			name: "with comments",
			content: `# Comment
KEY1=Value 1
# Another comment
KEY2=Value 2`,
			expected: map[string]string{
				"KEY1": "Value 1",
				"KEY2": "Value 2",
			},
		},
		{
			name: "empty lines",
			content: `KEY1=Value 1

KEY2=Value 2`,
			expected: map[string]string{
				"KEY1": "Value 1",
				"KEY2": "Value 2",
			},
		},
		{
			name: "value with equals sign",
			content: `KEY1=Value with = sign
KEY2=Another=one`,
			expected: map[string]string{
				"KEY1": "Value with = sign",
				"KEY2": "Another=one",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseMessages(tt.content)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d messages, got %d", len(tt.expected), len(result))
			}

			for key, expectedValue := range tt.expected {
				if actualValue, ok := result[key]; !ok {
					t.Errorf("Missing key: %s", key)
				} else if actualValue != expectedValue {
					t.Errorf("Key %s: expected %q, got %q", key, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		args     []interface{}
		expected string
	}{
		{
			name:     "existing key without args",
			key:      "ERR_NO_DESCRIPTION",
			args:     nil,
			expected: "Error: Description is required.",
		},
		{
			name:     "existing key with args",
			key:      "ERR_FILE_NOT_FOUND",
			args:     []interface{}{"test.md"},
			expected: "Error: File 'test.md' not found.",
		},
		{
			name:     "non-existent key",
			key:      "NON_EXISTENT",
			args:     nil,
			expected: "Unknown error: NON_EXISTENT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Error(tt.key, tt.args...)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPrompt(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		args     []interface{}
		expected string
	}{
		{
			name:     "existing key without args",
			key:      "MSG_SUCCESS",
			args:     nil,
			expected: "Success!",
		},
		{
			name:     "existing key with args",
			key:      "MSG_ADDED",
			args:     []interface{}{"Test task"},
			expected: "Added: Test task",
		},
		{
			name:     "non-existent key",
			key:      "NON_EXISTENT",
			args:     nil,
			expected: "Unknown prompt: NON_EXISTENT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Prompt(tt.key, tt.args...)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		args     []interface{}
		contains string
	}{
		{
			name:     "error key",
			key:      "ERR_NO_DESCRIPTION",
			args:     nil,
			contains: "Description is required",
		},
		{
			name:     "prompt key",
			key:      "MSG_SUCCESS",
			args:     nil,
			contains: "Success",
		},
		{
			name:     "non-existent key",
			key:      "TOTALLY_FAKE",
			args:     nil,
			contains: "Unknown message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(tt.key, tt.args...)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected result to contain %q, got %q", tt.contains, result)
			}
		})
	}
}

func TestHelp(t *testing.T) {
	help := Help()

	if help == "" {
		t.Error("Help text should not be empty")
	}

	// Check for expected sections
	expectedSections := []string{
		"mktodo",
		"Commands",
		"add",
		"done",
		"remove",
		"list",
		"open",
		"report",
		"Configuration",
	}

	for _, section := range expectedSections {
		if !strings.Contains(help, section) {
			t.Errorf("Help text should contain section: %s", section)
		}
	}
}

func TestErrorf(t *testing.T) {
	err := Errorf("ERR_FILE_NOT_FOUND", "test.md")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expected := "Error: File 'test.md' not found."
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}

func TestIcon(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "checked icon",
			key:      "checked",
			expected: "✓",
		},
		{
			name:     "unchecked icon",
			key:      "unchecked",
			expected: "✗",
		},
		{
			name:     "fixme icon",
			key:      "fixme",
			expected: "🔴",
		},
		{
			name:     "celebrate icon",
			key:      "celebrate",
			expected: "🎉",
		},
		{
			name:     "warning icon",
			key:      "warning",
			expected: "⚠️",
		},
		{
			name:     "non-existent icon",
			key:      "fake",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Icon(tt.key)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestHasError(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{
			name:     "existing error",
			key:      "ERR_NO_DESCRIPTION",
			expected: true,
		},
		{
			name:     "non-existent error",
			key:      "ERR_FAKE",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasError(tt.key)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestHasPrompt(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{
			name:     "existing prompt",
			key:      "MSG_SUCCESS",
			expected: true,
		},
		{
			name:     "non-existent prompt",
			key:      "MSG_FAKE",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasPrompt(tt.key)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestAllErrors(t *testing.T) {
	keys := AllErrors()

	if len(keys) == 0 {
		t.Error("Expected error keys, got empty list")
	}

	// Check for some known errors
	expectedKeys := []string{
		"ERR_NOT_IN_GIT_REPO",
		"ERR_CONFIG_NOT_FOUND",
		"ERR_NO_DESCRIPTION",
	}

	keyMap := make(map[string]bool)
	for _, k := range keys {
		keyMap[k] = true
	}

	for _, expected := range expectedKeys {
		if !keyMap[expected] {
			t.Errorf("Expected to find key %s in AllErrors()", expected)
		}
	}
}

func TestAllPrompts(t *testing.T) {
	keys := AllPrompts()

	if len(keys) == 0 {
		t.Error("Expected prompt keys, got empty list")
	}

	// Check for some known prompts
	expectedKeys := []string{
		"MSG_SUCCESS",
		"MSG_ADDED",
		"ICON_CHECKED",
	}

	keyMap := make(map[string]bool)
	for _, k := range keys {
		keyMap[k] = true
	}

	for _, expected := range expectedKeys {
		if !keyMap[expected] {
			t.Errorf("Expected to find key %s in AllPrompts()", expected)
		}
	}
}

func TestEmbeddedFilesLoaded(t *testing.T) {
	if errorsText == "" {
		t.Error("errorsText should not be empty")
	}

	if promptsText == "" {
		t.Error("promptsText should not be empty")
	}

	if helpText == "" {
		t.Error("helpText should not be empty")
	}
}

func TestMessageConsistency(t *testing.T) {
	// Test that all error keys start with ERR_
	for key := range errors {
		if !strings.HasPrefix(key, "ERR_") {
			t.Errorf("Error key %s should start with ERR_", key)
		}
	}

	// Test that message keys start with MSG_, PROMPT_, or ICON_
	for key := range prompts {
		if !strings.HasPrefix(key, "MSG_") &&
			!strings.HasPrefix(key, "PROMPT_") &&
			!strings.HasPrefix(key, "ICON_") {
			t.Errorf("Prompt key %s should start with MSG_, PROMPT_, or ICON_", key)
		}
	}
}

func TestFormatting(t *testing.T) {
	// Test that formatted messages work correctly
	msg := Error("ERR_FILE_NOT_FOUND", "test.md")
	if !strings.Contains(msg, "test.md") {
		t.Error("Formatted message should contain the argument")
	}

	msg = Prompt("MSG_ADDED", "My Task")
	if !strings.Contains(msg, "My Task") {
		t.Error("Formatted prompt should contain the argument")
	}
}
