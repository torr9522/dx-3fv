package handler

import "testing"

func TestNormalizeSequentialPortConfigValue(t *testing.T) {
	for in, want := range map[string]string{"true": "true", " FALSE ": "false"} {
		got, err := normalizeAndValidateConfigValue(sequentialPortConfigKey, in)
		if err != nil || got != want {
			t.Fatalf("value %q: got %q err=%v, want %q", in, got, err, want)
		}
	}
	if _, err := normalizeAndValidateConfigValue(sequentialPortConfigKey, "maybe"); err == nil {
		t.Fatal("expected invalid value to be rejected")
	}
}

func TestLegacySequentialPortConfigCannotOverrideRandomPolicy(t *testing.T) {
	_, r := setupConfigAccessTestRouter(t)
	h := &Handler{repo: r}
	if h.isSequentialPortEnabled() {
		t.Fatal("missing config must preserve random allocation")
	}
	for _, value := range []string{"true", " TRUE ", "false", "invalid"} {
		seedConfigValue(t, r, sequentialPortConfigKey, value)
		if got := h.isSequentialPortEnabled(); got {
			t.Fatalf("legacy config %q overrode the random policy", value)
		}
	}
}
