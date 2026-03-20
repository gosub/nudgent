package agent

import (
	"strings"
	"testing"
	"time"

	"github.com/gosub/nudgent/store"
)

var testNow = time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)

func TestParseResponse_Valid(t *testing.T) {
	raw := `{"reply": "Got it.", "actions": [{"type": "add_task", "description": "Call dentist"}]}`
	r, err := ParseResponse(raw)
	if err != nil {
		t.Fatalf("ParseResponse: %v", err)
	}
	if r.Reply != "Got it." {
		t.Errorf("Reply = %q", r.Reply)
	}
	if len(r.Actions) != 1 {
		t.Fatalf("len(Actions) = %d, want 1", len(r.Actions))
	}
	if r.Actions[0].Type != ActionAddTask {
		t.Errorf("Actions[0].Type = %q, want %q", r.Actions[0].Type, ActionAddTask)
	}
	if r.Actions[0].Description != "Call dentist" {
		t.Errorf("Actions[0].Description = %q", r.Actions[0].Description)
	}
}

func TestParseResponse_EmptyActions(t *testing.T) {
	raw := `{"reply": "Nothing to do.", "actions": []}`
	r, err := ParseResponse(raw)
	if err != nil {
		t.Fatalf("ParseResponse: %v", err)
	}
	if len(r.Actions) != 0 {
		t.Errorf("len(Actions) = %d, want 0", len(r.Actions))
	}
}

func TestParseResponse_Invalid(t *testing.T) {
	_, err := ParseResponse("not json")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseResponse_AllActionTypes(t *testing.T) {
	raw := `{"reply": "Done.", "actions": [
		{"type": "add_task",        "description": "Task"},
		{"type": "update_task",     "id": 1, "description": "Updated"},
		{"type": "set_next_nudge",  "id": 2, "next_nudge_at": "2026-03-21T09:00:00"},
		{"type": "complete_task",   "id": 3},
		{"type": "delete_task",     "id": 4},
		{"type": "update_schedule", "schedule": "weekdays 9-13"}
	]}`
	r, err := ParseResponse(raw)
	if err != nil {
		t.Fatalf("ParseResponse: %v", err)
	}
	if len(r.Actions) != 6 {
		t.Fatalf("len(Actions) = %d, want 6", len(r.Actions))
	}
}

func TestBuildChatPrompt_ContainsEssentials(t *testing.T) {
	tasks := []*store.Task{
		{ID: 1, Description: "Call dentist", NextNudgeAt: "2026-03-21T15:00:00"},
	}
	prompt := BuildChatPrompt("en", "weekdays 9-13", tasks, testNow)

	checks := []string{
		"Nudgent",
		"English",
		"2026-03-20",
		"weekdays 9-13",
		"Call dentist",
		"2026-03-21T15:00:00",
		`"reply"`,
		`"actions"`,
	}
	for _, c := range checks {
		if !strings.Contains(prompt, c) {
			t.Errorf("prompt missing %q", c)
		}
	}
}

func TestBuildChatPrompt_NoSchedule(t *testing.T) {
	prompt := BuildChatPrompt("en", "", nil, testNow)
	if !strings.Contains(prompt, "not set") {
		t.Error("prompt should say schedule not set")
	}
	if !strings.Contains(prompt, "none") {
		t.Error("prompt should say no active tasks")
	}
}

func TestBuildChatPrompt_Italian(t *testing.T) {
	prompt := BuildChatPrompt("it", "", nil, testNow)
	if !strings.Contains(prompt, "Italian") {
		t.Error("prompt missing Italian language instruction")
	}
}

func TestBuildNudgePrompt_ContainsEssentials(t *testing.T) {
	tasks := []*store.Task{
		{ID: 5, Description: "Submit report"},
	}
	prompt := BuildNudgePrompt("en", "weekdays 9-18", tasks, testNow)

	checks := []string{"Submit report", "nudge", "2026-03-20"}
	for _, c := range checks {
		if !strings.Contains(prompt, c) {
			t.Errorf("nudge prompt missing %q", c)
		}
	}
}
