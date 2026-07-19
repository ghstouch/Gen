package i18n

import (
	"testing"
)

func TestNew(t *testing.T) {
	tr := New("en", "en")
	if tr == nil {
		t.Fatal("New() returned nil")
	}
	if tr.GetLocale() != "en" {
		t.Errorf("GetLocale() = %q, want %q", tr.GetLocale(), "en")
	}
}

func TestT(t *testing.T) {
	tr := New("en", "en")
	tr.translations["en"] = map[string]string{
		"hello":  "Hello",
		"format": "My name is %s",
	}

	// Simple translation
	if tr.T("hello") != "Hello" {
		t.Errorf("T(hello) = %q, want %q", tr.T("hello"), "Hello")
	}

	// Translation with args
	tr.translations["en"]["formatKey"] = "My name is %s"
	result := tr.Tf("formatKey", "Gen")
	if result != "My name is Gen" {
		t.Errorf("Tf(formatKey, Gen) = %q, want %q", result, "My name is Gen")
	}

	// Missing key returns key
	if tr.T("missing") != "missing" {
		t.Errorf("T(missing) = %q, want %q", tr.T("missing"), "missing")
	}
}

func TestTFallback(t *testing.T) {
	tr := New("vi", "en")
	tr.translations["en"] = map[string]string{
		"hello": "Hello",
	}
	tr.translations["vi"] = map[string]string{
		"world": "Thế giới",
	}

	// Should use Vietnamese
	if tr.T("world") != "Thế giới" {
		t.Errorf("T(world) = %q, want %q", tr.T("world"), "Thế giới")
	}

	// Should fallback to English
	if tr.T("hello") != "Hello" {
		t.Errorf("T(hello) = %q, want %q", tr.T("hello"), "Hello")
	}
}

func TestSetLocale(t *testing.T) {
	tr := New("en", "en")
	tr.SetLocale("vi")

	if tr.GetLocale() != "vi" {
		t.Errorf("GetLocale() = %q, want %q", tr.GetLocale(), "vi")
	}
}

func TestGetAvailableLocales(t *testing.T) {
	tr := New("en", "en")
	tr.translations["en"] = map[string]string{}
	tr.translations["vi"] = map[string]string{}
	tr.translations["zh"] = map[string]string{}

	locales := tr.GetAvailableLocales()
	if len(locales) != 3 {
		t.Errorf("GetAvailableLocales count = %d, want 3", len(locales))
	}
}

func TestHasLocale(t *testing.T) {
	tr := New("en", "en")
	tr.translations["en"] = map[string]string{}

	if !tr.HasLocale("en") {
		t.Error("HasLocale(en) should be true")
	}
	if tr.HasLocale("vi") {
		t.Error("HasLocale(vi) should be false")
	}
}

func TestGetTranslations(t *testing.T) {
	tr := New("en", "en")
	tr.translations["en"] = map[string]string{
		"hello": "Hello",
	}

	translations := tr.GetTranslations("en")
	if translations["hello"] != "Hello" {
		t.Errorf("GetTranslations(en)[hello] = %q, want %q", translations["hello"], "Hello")
	}
}

func TestLoadTranslations(t *testing.T) {
	tr := New("en", "en")

	// This will fail if locales directory doesn't exist
	// But we can test the error handling
	err := tr.LoadTranslations("nonexistent")
	if err == nil {
		t.Error("LoadTranslations should fail for nonexistent directory")
	}
}

func TestConcurrentAccess(t *testing.T) {
	tr := New("en", "en")
	tr.translations["en"] = map[string]string{
		"hello": "Hello",
	}

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_ = tr.T("hello")
			_ = tr.GetLocale()
			_ = tr.GetAvailableLocales()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}