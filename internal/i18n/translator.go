package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Translator handles internationalization
type Translator struct {
	mu           sync.RWMutex
	locale       string
	fallback     string
	translations map[string]map[string]string
}

// New creates a new Translator
func New(locale, fallback string) *Translator {
	return &Translator{
		locale:       locale,
		fallback:     fallback,
		translations: make(map[string]map[string]string),
	}
}

// LoadTranslations loads translation files from a directory
func (t *Translator) LoadTranslations(dir string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read translations dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := filepath.Ext(name)
		if ext != ".json" {
			continue
		}

		locale := strings.TrimSuffix(name, ext)
		filePath := filepath.Join(dir, name)

		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read %s: %w", filePath, err)
		}

		var translations map[string]string
		if err := json.Unmarshal(data, &translations); err != nil {
			return fmt.Errorf("parse %s: %w", filePath, err)
		}

		t.translations[locale] = translations
	}

	return nil
}

// T translates a key to the current locale
func (t *Translator) T(key string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Try current locale
	if msg, ok := t.translations[t.locale]; ok {
		if val, ok := msg[key]; ok {
			return val
		}
	}

	// Try fallback locale
	if msg, ok := t.translations[t.fallback]; ok {
		if val, ok := msg[key]; ok {
			return val
		}
	}

	// Return key as fallback
	return key
}

// Tf translates a key with format arguments
func (t *Translator) Tf(key string, args ...interface{}) string {
	msg := t.T(key)
	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}
	return msg
}

// SetLocale changes the current locale
func (t *Translator) SetLocale(locale string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.locale = locale
}

// GetLocale returns the current locale
func (t *Translator) GetLocale() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.locale
}

// GetAvailableLocales returns all loaded locales
func (t *Translator) GetAvailableLocales() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	locales := make([]string, 0, len(t.translations))
	for locale := range t.translations {
		locales = append(locales, locale)
	}
	return locales
}

// HasLocale checks if a locale is loaded
func (t *Translator) HasLocale(locale string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	_, ok := t.translations[locale]
	return ok
}

// GetTranslations returns all translations for a locale
func (t *Translator) GetTranslations(locale string) map[string]string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.translations[locale]
}
