package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/joho/godotenv"

	"github.com/gosub/nudgent/agent"
	"github.com/gosub/nudgent/bot"
	"github.com/gosub/nudgent/log"
	"github.com/gosub/nudgent/store"
)

type Config struct {
	TelegramTokenEnv string `toml:"telegram_token_env"`
	OpenRouterKeyEnv string `toml:"openrouter_key_env"`
	AllowedUserIDEnv string `toml:"allowed_user_id_env"`
	Timezone         string `toml:"timezone"`
	Model            string `toml:"model"`
	Language         string `toml:"language"`
	NudgeIntervalM   int    `toml:"nudge_interval_m"`
}

func main() {
	logMode := os.Getenv("LOG_FORMAT") != "json"
	log.Init(logMode)
	logger := log.Logger.With().Str("component", "main").Logger()

	if err := godotenv.Load(); err != nil {
		logger.Warn().Err(err).Msg("no .env file found")
	}

	var cfg Config
	if _, err := toml.DecodeFile("config.toml", &cfg); err != nil {
		logger.Fatal().Err(err).Msg("failed to load config.toml")
	}

	telegramToken := os.Getenv(cfg.TelegramTokenEnv)
	if telegramToken == "" {
		logger.Fatal().Str("var", cfg.TelegramTokenEnv).Msg("missing env var")
	}

	openrouterKey := os.Getenv(cfg.OpenRouterKeyEnv)
	if openrouterKey == "" {
		logger.Fatal().Str("var", cfg.OpenRouterKeyEnv).Msg("missing env var")
	}

	allowedUserIDStr := os.Getenv(cfg.AllowedUserIDEnv)
	if allowedUserIDStr == "" {
		logger.Fatal().Str("var", cfg.AllowedUserIDEnv).Msg("missing env var")
	}
	allowedUserID, err := strconv.ParseInt(allowedUserIDStr, 10, 64)
	if err != nil || allowedUserID == 0 {
		logger.Fatal().Err(err).Msg("invalid ALLOWED_USER_ID")
	}

	if cfg.NudgeIntervalM <= 0 {
		cfg.NudgeIntervalM = 30
	}

	s, err := store.New("nudgent.db")
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to init store")
	}
	defer s.Close()

	if _, err := s.EnsurePrefs(context.Background(), allowedUserID, cfg.Language, cfg.NudgeIntervalM); err != nil {
		logger.Fatal().Err(err).Msg("failed to ensure prefs")
	}

	a := agent.New(openrouterKey, cfg.Model)

	b, err := bot.New(telegramToken, a, s, bot.Config{
		AllowedUserID:  allowedUserID,
		NudgeIntervalM: cfg.NudgeIntervalM,
		Timezone:       cfg.Timezone,
		Language:       cfg.Language,
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to init bot")
	}

	logger.Info().
		Int64("user_id", allowedUserID).
		Str("lang", cfg.Language).
		Int("nudge_interval_m", cfg.NudgeIntervalM).
		Str("timezone", cfg.Timezone).
		Str("model", cfg.Model).
		Msg("starting")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go b.Run(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	logger.Info().Str("signal", sig.String()).Msg("shutting down")
	cancel()
	fmt.Println("Goodbye!")
}
