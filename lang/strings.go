package lang

import "fmt"

var strings = map[string]map[string]string{
	"en": {
		"welcome":          "Hello! I'm Nudgent. Tell me what you need to do.",
		"tasks_empty":      "No active tasks.",
		"tasks_header":     "Active tasks:",
		"message_too_long": "Message is too long (max %d characters).",
		"help":             "/tasks — list active tasks\n/help — show this message\n\nOr just tell me what you need.",
	},
	"it": {
		"welcome":          "Ciao! Sono Nudgent. Dimmi cosa devi fare.",
		"tasks_empty":      "Nessun compito attivo.",
		"tasks_header":     "Compiti attivi:",
		"message_too_long": "Messaggio troppo lungo (max %d caratteri).",
		"help":             "/tasks — lista compiti attivi\n/help — mostra questo messaggio\n\nOppure dimmi quello che ti serve.",
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
