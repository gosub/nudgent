# Nudgent

Intelligent nudge agent. Tracks your tasks, knows your schedule, and nudges you at the right time.

Runs as a single Go binary, communicates via Telegram, powered by OpenRouter models.

## How it works

You talk to the bot in plain language. It understands what you mean, stores tasks as freeform descriptions, and wakes up periodically to nudge you when something is due. Recurrence, priority, and context all live in the task description — no rigid fields, no forms.

```
You:    add call dentist before end of March, working hours
Bot:    Added: "Call dentist — before end of March, working hours".

You:    done with the gym
Bot:    Marked "Gym" as done.

You:    push the report to Thursday morning
Bot:    Rescheduled "Write report" to Thursday 09:00.

[nudge, unprompted]
Bot:    "Call dentist" is overdue — end of March is tomorrow.
```

## Build

```
git clone https://github.com/gosub/nudgent.git
cd nudgent
make build
```

Produces a single static binary: `./nudgent`.

Copy `.env.example` to `.env` and fill in your secrets:

```
TELEGRAM_TOKEN=your-telegram-bot-token
OPENROUTER_KEY=your-openrouter-api-key
ALLOWED_USER_ID=your-telegram-numeric-id
```

Adjust `config.toml` as needed:

```toml
timezone         = "Europe/Rome"
model            = "openai/gpt-4o-mini"
language         = "en"
nudge_interval_m = 30
```

Then run:

```
./nudgent
```

## Commands

| Command  | Description                                      |
|----------|--------------------------------------------------|
| `/tasks` | List active tasks with next nudge times          |
| `/debug` | Show prefs and full task state (IDs, nudge times)|
| `/help`  | Show usage                                       |

Everything else — adding tasks, marking done, rescheduling, setting your schedule — is plain chat.

## Deployment

To run on a server, copy three files and the binary:

```
nudgent          # compiled binary (make build)
config.toml      # timezone, model, language, nudge interval
.env             # secrets: TELEGRAM_TOKEN, OPENROUTER_KEY, ALLOWED_USER_ID
nudgent.db       # created automatically on first run
```

To run as a systemd user service:

```
cp nudgent.service ~/.config/systemd/user/
systemctl --user daemon-reload
systemctl --user enable --now nudgent
```

The service file expects the binary and config in `~/nudgent/`. Adjust `WorkingDirectory` and `ExecStart` in `nudgent.service` if you use a different path.

## Project Structure

```
nudgent/
├── main.go        # Entry point, config loading, wiring
├── bot/           # Telegram polling, command and chat handlers, nudge engine
├── agent/         # OpenRouter client, system prompt builder, action parser
└── store/         # SQLite: tasks and prefs
```

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/BurntSushi/toml` | TOML config parsing |
| `github.com/joho/godotenv` | .env file loading |
| `github.com/go-telegram-bot-api/telegram-bot-api/v5` | Telegram bot API |
| `github.com/rs/zerolog` | Structured logging |
| `modernc.org/sqlite` | Pure Go SQLite |

## License

GNU General Public License v3.0. See [LICENSE](LICENSE) for details.
