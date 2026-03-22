package agent

import (
	"fmt"
	"strings"
	"time"

	"github.com/gosub/nudgent/store"
)

func BuildChatPrompt(language, schedule string, tasks []*store.Task, now time.Time) string {
	var sb strings.Builder

	sb.WriteString("You are Nudgent, an intelligent task and nudge assistant.\n")
	sb.WriteString("You help the user track tasks, remember commitments, and get things done.\n")
	sb.WriteString("Be concise and direct.\n")
	sb.WriteString(fmt.Sprintf("Always respond in %s.\n\n", langName(language)))

	sb.WriteString(fmt.Sprintf("Current time: %s\n\n", now.Format("2006-01-02T15:04:05 (Monday)")))

	if schedule != "" {
		sb.WriteString(fmt.Sprintf("User's schedule: %s\n\n", schedule))
	} else {
		sb.WriteString("User's schedule: not set\n\n")
	}

	if len(tasks) == 0 {
		sb.WriteString("Active tasks: none\n\n")
	} else {
		sb.WriteString(fmt.Sprintf("Active tasks (%d):\n", len(tasks)))
		for _, t := range tasks {
			nudge := "not set"
			if t.NextNudgeAt != "" {
				nudge = t.NextNudgeAt
			}
			prefix := ""
			if t.Recurring {
				prefix = "↻ "
			}
			sb.WriteString(fmt.Sprintf("  %d. %s%s — next nudge: %s\n", t.ID, prefix, t.Description, nudge))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Respond ONLY with a JSON object: {\"reply\": \"...\", \"actions\": [...]}\n")
	sb.WriteString("No text outside the JSON. If no actions are needed, use \"actions\": [].\n\n")
	sb.WriteString("Available actions:\n")
	sb.WriteString("  {\"type\": \"add_task\",        \"description\": \"...\", \"next_nudge_at\": \"ISO8601\", \"recurring\": true}  — next_nudge_at and recurring optional\n")
	sb.WriteString("  {\"type\": \"update_task\",     \"id\": N, \"description\": \"...\", \"next_nudge_at\": \"ISO8601\", \"recurring\": true}  — all fields optional\n")
	sb.WriteString("  {\"type\": \"complete_task\",   \"id\": N}\n")
	sb.WriteString("  {\"type\": \"delete_task\",     \"id\": N}\n")
	sb.WriteString("  {\"type\": \"update_schedule\", \"schedule\": \"...\"}\n")
	sb.WriteString("Always use numeric id from the task list. When adding a task with a known time, always include next_nudge_at.\n")
	sb.WriteString("next_nudge_at must be ISO 8601 (e.g. 2026-03-21T09:00:00). Respect the user's schedule.\n")
	sb.WriteString("Set recurring: true when the task is habitual or repeats on a schedule; leave it false for one-off tasks.\n")
	sb.WriteString("Recurring tasks (↻) must never be completed — use update_task with the next next_nudge_at instead.\n")

	return sb.String()
}

func BuildNudgePrompt(language, schedule string, tasks []*store.Task, now time.Time) string {
	var sb strings.Builder

	sb.WriteString("You are Nudgent, a nudge agent.\n")
	sb.WriteString(fmt.Sprintf("Current time: %s\n", now.Format("2006-01-02T15:04:05 (Monday)")))
	if schedule != "" {
		sb.WriteString(fmt.Sprintf("User's schedule: %s\n\n", schedule))
	}
	sb.WriteString(fmt.Sprintf("Always respond in %s.\n\n", langName(language)))

	sb.WriteString("The following tasks are due for a nudge:\n")
	for _, t := range tasks {
		prefix := ""
		if t.Recurring {
			prefix = "↻ "
		}
		sb.WriteString(fmt.Sprintf("  %d. %s%s\n", t.ID, prefix, t.Description))
	}

	sb.WriteString("\nSend the user a short nudge. One task, one sentence, no fluff.\n")
	sb.WriteString("If multiple tasks are due, pick the most urgent one.\n")
	sb.WriteString("After nudging a recurring task (↻), always use update_task to set the next next_nudge_at.\n")
	sb.WriteString("If no nudge is appropriate right now, return empty reply.\n\n")
	sb.WriteString("Respond: {\"reply\": \"...\", \"actions\": [...]}\n")
	sb.WriteString("Actions available: update_task (id, description optional, next_nudge_at optional), complete_task (id), delete_task (id).\n")

	return sb.String()
}

func BuildSchedulePrompt(language, schedule string, tasks []*store.Task, now time.Time) string {
	var sb strings.Builder

	sb.WriteString("You are Nudgent, a scheduling assistant.\n")
	sb.WriteString(fmt.Sprintf("Current time: %s\n", now.Format("2006-01-02T15:04:05 (Monday)")))
	if schedule != "" {
		sb.WriteString(fmt.Sprintf("User's schedule: %s\n\n", schedule))
	}
	sb.WriteString(fmt.Sprintf("Always respond in %s.\n\n", langName(language)))

	sb.WriteString("The following tasks have no scheduled reminder time:\n")
	for _, t := range tasks {
		sb.WriteString(fmt.Sprintf("  %d. %s\n", t.ID, t.Description))
	}

	sb.WriteString("\nAssign each task a next_nudge_at using update_task actions.\n")
	sb.WriteString("Base the time on the task description and the user's schedule. If unclear, schedule within 24 hours.\n")
	sb.WriteString("Respond: {\"reply\": \"\", \"actions\": [...]}\n")
	sb.WriteString("Actions: update_task (id, next_nudge_at required). No reply text.\n")

	return sb.String()
}

func langName(code string) string {
	switch code {
	case "it":
		return "Italian"
	case "en":
		return "English"
	default:
		return "English"
	}
}
