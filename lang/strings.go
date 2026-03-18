package lang

import "fmt"

var strings = map[string]map[string]string{
	"en": {
		"welcome":           "Welcome back. Where are you in the plan?",
		"status_header":     "Your current status:",
		"status_phase":      "Phase: %d",
		"status_goals":      "Goals: %s",
		"status_rejections": "Rejections logged: %d",
		"status_tone":       "Tone: %s",
		"status_lang":       "Language: %s",
		"rejection_logged":  "Logged! Total: %d. Keep going.",
		"goal_added":        "Goal added: %s",
		"goal_list":         "Your goals:",
		"goal_none":         "No goals set yet. Use /goal add <goal> to add one.",
		"goal_completed":    "Goal completed: %s",
		"goal_invalid":      "Invalid goal index. Use /goal list to see numbered goals.",
		"lang_current":      "Current language: %s",
		"lang_switched":     "Language switched to: %s",
		"tone_current":      "Current tone: %s",
		"tone_options":      "Available tones: warm, direct, drill-sergeant",
		"tone_switched":     "Tone switched to: %s",
		"checkin_skipped":   "Check-in skipped for today.",
		"checkin_msg":       "Day %d of your agency journey. Where are you at?",
		"help_header":       "Available commands:",
		"help_start":        "/start - Welcome message",
		"help_status":       "/status - Show current status",
		"help_rejection":    "/rejection - Log a rejection",
		"help_goal":         "/goal add <text> - Add a goal",
		"help_goal_list":    "/goal list - List goals",
		"help_goal_done":    "/goal done <number> - Complete a goal",
		"help_skip":         "/skip - Skip today's check-in",
		"help_lang":         "/lang [it|en] - Show/set language",
		"help_tone":         "/tone [warm|direct|drill-sergeant] - Show/set tone",
		"help_reset":        "/reset - Clear conversation context",
		"help_help":         "/help - Show this message",
		"goal_too_long":     "Goal is too long (max %d characters).",
		"message_too_long":  "Message is too long (max %d characters).",
	},
	"it": {
		"welcome":           "Bentornato. Dove sei nel piano?",
		"status_header":     "Il tuo stato attuale:",
		"status_phase":      "Fase: %d",
		"status_goals":      "Obiettivi: %s",
		"status_rejections": "Rifiuti registrati: %d",
		"status_tone":       "Tono: %s",
		"status_lang":       "Lingua: %s",
		"rejection_logged":  "Registrato! Totale: %d. Continua così.",
		"goal_added":        "Obiettivo aggiunto: %s",
		"goal_list":         "I tuoi obiettivi:",
		"goal_none":         "Nessun obiettivo ancora. Usa /goal add <obiettivo> per aggiungerne uno.",
		"goal_completed":    "Obiettivo completato: %s",
		"goal_invalid":      "Indice obiettivo non valido. Usa /goal list per vedere gli obiettivi numerati.",
		"lang_current":      "Lingua attuale: %s",
		"lang_switched":     "Lingua cambiata in: %s",
		"tone_current":      "Tono attuale: %s",
		"tone_options":      "Toni disponibili: warm, direct, drill-sergeant",
		"tone_switched":     "Tono cambiato in: %s",
		"checkin_skipped":   "Check-in di oggi saltato.",
		"checkin_msg":       "Giorno %d del tuo percorso di agency. Come te la passi?",
		"help_header":       "Comandi disponibili:",
		"help_start":        "/start - Messaggio di benvenuto",
		"help_status":       "/status - Mostra stato attuale",
		"help_rejection":    "/rejection - Registra un rifiuto",
		"help_goal":         "/goal add <testo> - Aggiungi un obiettivo",
		"help_goal_list":    "/goal list - Lista obiettivi",
		"help_goal_done":    "/goal done <numero> - Completa un obiettivo",
		"help_skip":         "/skip - Salta il check-in di oggi",
		"help_lang":         "/lang [it|en] - Mostra/cambia lingua",
		"help_tone":         "/tone [warm|direct|drill-sergeant] - Mostra/cambia tono",
		"help_reset":        "/reset - Cancella contesto conversazione",
		"help_help":         "/help - Mostra questo messaggio",
		"goal_too_long":     "Obiettivo troppo lungo (max %d caratteri).",
		"message_too_long":  "Messaggio troppo lungo (max %d caratteri).",
	},
}

func Get(lang, key string) string {
	if ls, ok := strings[lang]; ok {
		if s, ok := ls[key]; ok {
			return s
		}
	}
	if ls, ok := strings["en"]; ok {
		if s, ok := ls[key]; ok {
			return s
		}
	}
	return key
}

func Getf(lang, key string, args ...interface{}) string {
	return fmt.Sprintf(Get(lang, key), args...)
}

func IsValidLang(lang string) bool {
	_, ok := strings[lang]
	return ok
}
