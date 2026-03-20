package store

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type Task struct {
	ID           int64
	UserID       int64
	Description  string
	NextNudgeAt  string // ISO 8601, empty if not set
	Done         bool
}

type Prefs struct {
	UserID         int64
	Language       string
	NudgeIntervalM int
	Schedule       string // freeform, e.g. "weekdays 9-13 and 15-19"
	LastWakeupAt   string // ISO 8601, empty if never run
}

type Store struct {
	db *sql.DB
}

type Storager interface {
	// Tasks
	AddTask(ctx context.Context, userID int64, description string) (*Task, error)
	GetTasks(ctx context.Context, userID int64) ([]*Task, error)
	UpdateTask(ctx context.Context, id int64, description string) error
	SetNextNudgeAt(ctx context.Context, id int64, nextNudgeAt string) error
	CompleteTask(ctx context.Context, id int64) error
	DeleteTask(ctx context.Context, id int64) error
	GetDueTasks(ctx context.Context, userID int64, now string) ([]*Task, error)

	// Prefs
	EnsurePrefs(ctx context.Context, userID int64, defaultLang string, defaultInterval int) (*Prefs, error)
	GetPrefs(ctx context.Context, userID int64) (*Prefs, error)
	SetLanguage(ctx context.Context, userID int64, lang string) error
	SetSchedule(ctx context.Context, userID int64, schedule string) error
	SetLastWakeupAt(ctx context.Context, userID int64, t string) error
}

var _ Storager = (*Store)(nil)

func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if _, err := db.Exec(`PRAGMA journal_mode=WAL`); err != nil {
		return nil, fmt.Errorf("set pragma: %w", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS tasks (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id       INTEGER NOT NULL,
		description   TEXT    NOT NULL,
		next_nudge_at TEXT,
		done          INTEGER NOT NULL DEFAULT 0
	);
	CREATE TABLE IF NOT EXISTS prefs (
		user_id          INTEGER PRIMARY KEY,
		language         TEXT    NOT NULL DEFAULT 'en',
		nudge_interval_m INTEGER NOT NULL DEFAULT 30,
		schedule         TEXT    NOT NULL DEFAULT '',
		last_wakeup_at   TEXT
	);`

	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("create schema: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

// --- Tasks ---

func (s *Store) AddTask(ctx context.Context, userID int64, description string) (*Task, error) {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO tasks (user_id, description) VALUES (?, ?)`,
		userID, description)
	if err != nil {
		return nil, fmt.Errorf("insert task: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id: %w", err)
	}
	return &Task{ID: id, UserID: userID, Description: description}, nil
}

func (s *Store) GetTasks(ctx context.Context, userID int64) ([]*Task, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, description, next_nudge_at, done
		 FROM tasks WHERE user_id = ? AND done = 0
		 ORDER BY id ASC`, userID)
	if err != nil {
		return nil, fmt.Errorf("query tasks: %w", err)
	}
	defer rows.Close()
	return scanTasks(rows)
}

func (s *Store) GetDueTasks(ctx context.Context, userID int64, now string) ([]*Task, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, description, next_nudge_at, done
		 FROM tasks WHERE user_id = ? AND done = 0 AND next_nudge_at <= ?
		 ORDER BY next_nudge_at ASC`, userID, now)
	if err != nil {
		return nil, fmt.Errorf("query due tasks: %w", err)
	}
	defer rows.Close()
	return scanTasks(rows)
}

func (s *Store) UpdateTask(ctx context.Context, id int64, description string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE tasks SET description = ? WHERE id = ?`, description, id)
	return err
}

func (s *Store) SetNextNudgeAt(ctx context.Context, id int64, nextNudgeAt string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE tasks SET next_nudge_at = ? WHERE id = ?`, nextNudgeAt, id)
	return err
}

func (s *Store) CompleteTask(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE tasks SET done = 1 WHERE id = ?`, id)
	return err
}

func (s *Store) DeleteTask(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM tasks WHERE id = ?`, id)
	return err
}

func scanTasks(rows *sql.Rows) ([]*Task, error) {
	var tasks []*Task
	for rows.Next() {
		t := &Task{}
		var nextNudgeAt sql.NullString
		var done int
		if err := rows.Scan(&t.ID, &t.UserID, &t.Description, &nextNudgeAt, &done); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		if nextNudgeAt.Valid {
			t.NextNudgeAt = nextNudgeAt.String
		}
		t.Done = done != 0
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// --- Prefs ---

func (s *Store) EnsurePrefs(ctx context.Context, userID int64, defaultLang string, defaultInterval int) (*Prefs, error) {
	p, err := s.GetPrefs(ctx, userID)
	if err != nil {
		return nil, err
	}
	if p != nil {
		return p, nil
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO prefs (user_id, language, nudge_interval_m) VALUES (?, ?, ?)`,
		userID, defaultLang, defaultInterval)
	if err != nil {
		return nil, fmt.Errorf("insert prefs: %w", err)
	}
	return &Prefs{UserID: userID, Language: defaultLang, NudgeIntervalM: defaultInterval}, nil
}

func (s *Store) GetPrefs(ctx context.Context, userID int64) (*Prefs, error) {
	p := &Prefs{}
	var lastWakeupAt sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT user_id, language, nudge_interval_m, schedule, last_wakeup_at
		 FROM prefs WHERE user_id = ?`, userID).
		Scan(&p.UserID, &p.Language, &p.NudgeIntervalM, &p.Schedule, &lastWakeupAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan prefs: %w", err)
	}
	if lastWakeupAt.Valid {
		p.LastWakeupAt = lastWakeupAt.String
	}
	return p, nil
}

func (s *Store) SetLanguage(ctx context.Context, userID int64, lang string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE prefs SET language = ? WHERE user_id = ?`, lang, userID)
	return err
}

func (s *Store) SetSchedule(ctx context.Context, userID int64, schedule string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE prefs SET schedule = ? WHERE user_id = ?`, schedule, userID)
	return err
}

func (s *Store) SetLastWakeupAt(ctx context.Context, userID int64, t string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE prefs SET last_wakeup_at = ? WHERE user_id = ?`, t, userID)
	return err
}
