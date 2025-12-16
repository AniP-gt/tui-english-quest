package i18n

import "testing"

func TestT_JapaneseTownPrompt(t *testing.T) {
	SetLang("ja")
	v := T("town_menu_prompt")
	if v == "Where do you want to go?" || v == "[town_menu_prompt]" {
		t.Fatalf("expected Japanese town_menu_prompt, got: %s", v)
	}
}

func TestSetLangNormalization(t *testing.T) {
	SetLang(" JA ")
	if lang != "ja" {
		t.Fatalf("expected lang to normalize to 'ja', got '%s'", lang)
	}
	SetLang("unknown")
	if lang != "en" {
		t.Fatalf("expected unknown lang to fallback to 'en', got '%s'", lang)
	}
}
