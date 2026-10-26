package provider

import "testing"

func TestCanonicalEndpointSettingBoolAndNumber(t *testing.T) {
	obj := map[string]any{
		"EndpointDlpGlobalSettings": []any{
			map[string]any{"Value": true, "Setting": "AdvancedClassificationEnabled"},
			map[string]any{"setting": "DailyBandwidthLimitInMB", "value": float64(1000)},
		},
	}
	got := endpointDlpGlobalSettingStrings(obj)
	want := []string{
		`{"Setting":"AdvancedClassificationEnabled","Value":"true"}`,
		`{"Setting":"DailyBandwidthLimitInMB","Value":"1000"}`,
	}
	if len(got) != len(want) {
		t.Fatalf("len %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entry %d = %s, want %s", i, got[i], want[i])
		}
	}
}

func TestHashtableEndpointSettingsExpandsArrayString(t *testing.T) {
	raw := []string{
		`[{"Setting":"CloudAppMode","Value":"Allow"},{"Setting":"CloudAppRestrictionList","Value":"*.savex-internal.com"}]`,
	}
	got, err := hashtableEndpointSettings(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("len %d, want 2", len(got))
	}
	first, ok := got[0].(map[string]any)
	if !ok || first["Setting"] != "CloudAppMode" || first["Value"] != "Allow" {
		t.Fatalf("first = %#v", got[0])
	}
}

func TestHashtableEndpointSettingsRejectsText(t *testing.T) {
	if _, err := hashtableEndpointSettings([]string{"Allow"}); err == nil {
		t.Fatal("expected error")
	}
}
