package bot

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	"github.com/gosub/nudgent/agent"
	"github.com/gosub/nudgent/store"
)

func (b *Bot) nudgeScheduler(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(b.cfg.NudgeIntervalM) * time.Minute)
	defer ticker.Stop()

	l := b.log.With().Str("handler", "nudge").Logger()
	l.Debug().Int("interval_m", b.cfg.NudgeIntervalM).Msg("nudge scheduler started")
	b.runNudgeCycle(ctx)

	for {
		select {
		case <-ctx.Done():
			l.Debug().Msg("nudge scheduler stopped")
			return
		case <-ticker.C:
			b.runNudgeCycle(ctx)
		}
	}
}

func (b *Bot) runNudgeCycle(ctx context.Context) {
	l := b.log.With().Str("handler", "nudge").Logger()

	now := time.Now().In(b.loc)
	l.Debug().Str("now", now.Format("2006-01-02T15:04:05")).Msg("nudge cycle")
	nowISO := now.Format("2006-01-02T15:04:05")

	if err := b.store.SetLastWakeupAt(ctx, b.cfg.AllowedUserID, nowISO); err != nil {
		l.Error().Err(err).Msg("set last wakeup failed")
	}

	due, err := b.store.GetDueTasks(ctx, b.cfg.AllowedUserID, nowISO)
	if err != nil {
		l.Error().Err(err).Msg("get due tasks failed")
		return
	}

	if len(due) > 0 {
		b.runNudgePhase(ctx, l, due, now, nowISO)
	}

	b.runSchedulePhase(ctx, l, now)
}

func (b *Bot) runNudgePhase(ctx context.Context, l zerolog.Logger, due []*store.Task, now time.Time, nowISO string) {
	p, err := b.store.GetPrefs(ctx, b.cfg.AllowedUserID)
	if err != nil || p == nil {
		l.Error().Err(err).Msg("get prefs failed")
		return
	}

	prompt := agent.BuildNudgePrompt(p.Language, p.Schedule, due, now)
	l.Trace().Str("prompt", prompt).Msg("nudge system prompt")

	trigger := fmt.Sprintf("Nudge check at %s. %d task(s) due for a reminder.", nowISO, len(due))
	raw, err := b.agent.Chat(ctx, prompt, nil, trigger)
	if err != nil {
		if ctx.Err() != nil {
			return
		}
		l.Error().Err(err).Msg("nudge agent call failed")
		return
	}

	l.Debug().Str("raw", raw).Msg("nudge agent response")
	resp, err := agent.ParseResponse(raw)
	if err != nil {
		l.Warn().Err(err).Str("raw", raw).Msg("failed to parse nudge response")
		return
	}

	b.executeActions(ctx, resp.Actions)

	// Clear next_nudge_at for any tasks still due (LLM didn't reschedule them).
	// Prevents tasks from firing every cycle; runSchedulePhase will re-schedule them.
	stillDue, err := b.store.GetDueTasks(ctx, b.cfg.AllowedUserID, nowISO)
	if err == nil {
		for _, t := range stillDue {
			if err := b.store.SetNextNudgeAt(ctx, t.ID, ""); err != nil {
				l.Error().Err(err).Int64("id", t.ID).Msg("clear nudge failed")
			}
		}
		if len(stillDue) > 0 {
			l.Warn().Int("count", len(stillDue)).Msg("tasks not rescheduled by LLM — will schedule in next phase")
		}
	}

	if resp.Reply != "" {
		l.Info().Int("due_tasks", len(due)).Msg("sending nudge")
		b.send(b.cfg.AllowedUserID, resp.Reply)
	}
	l.Debug().Msg("nudge phase done")
}

func (b *Bot) runSchedulePhase(ctx context.Context, l zerolog.Logger, now time.Time) {
	tasks, err := b.store.GetTasks(ctx, b.cfg.AllowedUserID)
	if err != nil {
		l.Error().Err(err).Msg("get tasks for schedule phase failed")
		return
	}

	var unscheduled []*store.Task
	for _, t := range tasks {
		if t.NextNudgeAt == "" {
			unscheduled = append(unscheduled, t)
		}
	}
	if len(unscheduled) == 0 {
		return
	}

	l.Debug().Int("count", len(unscheduled)).Msg("scheduling unscheduled tasks")

	p, err := b.store.GetPrefs(ctx, b.cfg.AllowedUserID)
	if err != nil || p == nil {
		l.Error().Err(err).Msg("get prefs for schedule phase failed")
		return
	}

	prompt := agent.BuildSchedulePrompt(p.Language, p.Schedule, unscheduled, now)
	l.Trace().Str("prompt", prompt).Msg("schedule system prompt")

	trigger := fmt.Sprintf("%d task(s) have no reminder time set.", len(unscheduled))
	raw, err := b.agent.Chat(ctx, prompt, nil, trigger)
	if err != nil {
		if ctx.Err() != nil {
			return
		}
		l.Error().Err(err).Msg("schedule phase agent call failed")
		return
	}

	l.Debug().Str("raw", raw).Msg("schedule phase response")
	resp, err := agent.ParseResponse(raw)
	if err != nil {
		l.Warn().Err(err).Str("raw", raw).Msg("failed to parse schedule phase response")
		return
	}

	b.executeActions(ctx, resp.Actions)
	l.Debug().Msg("schedule phase done")
}
