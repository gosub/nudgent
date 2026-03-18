package coach

import (
	"fmt"
	"strings"
	"testing"
)

func TestBuildSystemPrompt_Identity(t *testing.T) {
	prompt := BuildSystemPrompt("compendium text", "en", "warm", 0, nil, nil)

	if !strings.Contains(prompt, "You are Maxxx Agency") {
		t.Error("prompt missing identity line")
	}
	if !strings.Contains(prompt, "personal agency coach") {
		t.Error("prompt missing role description")
	}
}

func TestBuildSystemPrompt_Language(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{"en", "Always respond in English"},
		{"it", "Always respond in Italian"},
		{"fr", "Always respond in English"}, // unknown defaults to English
	}

	for _, tt := range tests {
		prompt := BuildSystemPrompt("", tt.code, "warm", 0, nil, nil)
		if !strings.Contains(prompt, tt.expected) {
			t.Errorf("language %q: prompt missing %q", tt.code, tt.expected)
		}
	}
}

func TestBuildSystemPrompt_Tone(t *testing.T) {
	tests := []struct {
		tone     string
		expected string
	}{
		{"warm", "warm and encouraging"},
		{"direct", "direct and efficient"},
		{"drill-sergeant", "intense and demanding"},
		{"unknown", "warm and encouraging"}, // unknown defaults to warm
	}

	for _, tt := range tests {
		prompt := BuildSystemPrompt("", "en", tt.tone, 0, nil, nil)
		if !strings.Contains(prompt, tt.expected) {
			t.Errorf("tone %q: prompt missing %q", tt.tone, tt.expected)
		}
	}
}

func TestBuildSystemPrompt_Compendium(t *testing.T) {
	compendium := "Agency is the ability to get what you want."
	prompt := BuildSystemPrompt(compendium, "en", "warm", 0, nil, nil)

	if !strings.Contains(prompt, compendium) {
		t.Error("prompt missing compendium content")
	}
	if !strings.Contains(prompt, "--- AGENCY FRAMEWORK REFERENCE ---") {
		t.Error("prompt missing reference markers")
	}
	if !strings.Contains(prompt, "--- END REFERENCE ---") {
		t.Error("prompt missing end reference marker")
	}
}

func TestBuildSystemPrompt_Phase(t *testing.T) {
	for phase := 0; phase <= 3; phase++ {
		prompt := BuildSystemPrompt("", "en", "warm", phase, nil, nil)
		expected := fmt.Sprintf("Phase: %d", phase)
		if !strings.Contains(prompt, expected) {
			t.Errorf("phase %d: prompt missing %q", phase, expected)
		}
	}
}

func TestBuildSystemPrompt_Goals(t *testing.T) {
	t.Run("with goals", func(t *testing.T) {
		prompt := BuildSystemPrompt("", "en", "warm", 0,
			[]string{"Read Dune", "Start log"}, nil)
		if !strings.Contains(prompt, "Read Dune") {
			t.Error("prompt missing first goal")
		}
		if !strings.Contains(prompt, "Start log") {
			t.Error("prompt missing second goal")
		}
	})

	t.Run("no goals", func(t *testing.T) {
		prompt := BuildSystemPrompt("", "en", "warm", 0, nil, nil)
		if !strings.Contains(prompt, "Active goals: none") {
			t.Error("prompt should say 'Active goals: none'")
		}
	})
}

func TestBuildSystemPrompt_RejectionCount(t *testing.T) {
	rejections := []string{"2026-01-01", "2026-01-02", "2026-01-03"}
	prompt := BuildSystemPrompt("", "en", "warm", 0, nil, rejections)

	if !strings.Contains(prompt, "Rejections logged: 3") {
		t.Error("prompt missing correct rejection count")
	}
}

func TestBuildSystemPrompt_BehavioralRules(t *testing.T) {
	prompt := BuildSystemPrompt("", "en", "warm", 0, nil, nil)

	rules := []string{
		"Ask one good question at a time",
		"Celebrate rejections and small wins",
		"Suggest next phase when current tasks are done",
	}

	for _, rule := range rules {
		if !strings.Contains(prompt, rule) {
			t.Errorf("prompt missing behavioral rule: %q", rule)
		}
	}
}
