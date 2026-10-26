package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// applyEndpointDlpGlobalSettings writes EndpointDlpGlobalSettings and nothing else.
//
// Set-PolicyConfig replaces that list. Sending the rest of the read-back object
// (ReservedForFutureUse and the group JSON blobs) returns HTTP 400 Invalid
// Operation. The parameter type is PswsHashtable[], so each configured JSON
// object is sent as a map, not as a string.
func (r *policyConfigResource) applyEndpointDlpGlobalSettings(ctx context.Context, plan policyConfigModel, state *policyConfigModel, diags *diag.Diagnostics) error {
	if state != nil && plan.EndpointDlpGlobalSettings.Equal(state.EndpointDlpGlobalSettings) {
		return nil
	}
	raw := toStringSlice(ctx, plan.EndpointDlpGlobalSettings, diags)
	if diags.HasError() {
		return fmt.Errorf("endpoint dlp global settings could not be read from the plan")
	}
	settings, err := hashtableEndpointSettings(raw)
	if err != nil {
		return err
	}
	if len(settings) == 0 {
		return fmt.Errorf("refusing to clear EndpointDlpGlobalSettings")
	}
	identity := plan.Identity.ValueString()
	if state != nil && state.Identity.ValueString() != "" {
		identity = state.Identity.ValueString()
	}
	_, err = r.client.Admin.Invoke(ctx, "Set-PolicyConfig", map[string]any{
		"Identity":                  identity,
		"EndpointDlpGlobalSettings": settings,
	})
	return err
}

func hashtableEndpointSettings(raw []string) ([]any, error) {
	out := make([]any, 0, len(raw))
	for _, item := range raw {
		parsed, err := parseEndpointSettingElements(item)
		if err != nil {
			return nil, err
		}
		for _, element := range parsed {
			object, ok := element.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("endpoint dlp setting must be an object, got %T", element)
			}
			out = append(out, object)
		}
	}
	return out, nil
}

func endpointDlpGlobalSettingStrings(obj map[string]any) []string {
	value, ok := obj["EndpointDlpGlobalSettings"]
	if !ok || value == nil {
		return nil
	}
	return canonicalEndpointSettingElements(value)
}

func canonicalEndpointSettingElements(value any) []string {
	switch typed := value.(type) {
	case string:
		parsed, err := parseEndpointSettingElements(typed)
		if err != nil {
			return []string{typed}
		}
		return canonicalizeEndpointElements(parsed)
	case []any:
		return canonicalizeEndpointElements(typed)
	case map[string]any:
		if encoded, ok := canonicalEndpointSetting(typed); ok {
			return []string{encoded}
		}
	}
	return nil
}

func canonicalizeEndpointElements(elements []any) []string {
	out := make([]string, 0, len(elements))
	for _, element := range elements {
		switch typed := element.(type) {
		case string:
			out = append(out, canonicalEndpointSettingElements(typed)...)
		case map[string]any:
			if encoded, ok := canonicalEndpointSetting(typed); ok {
				out = append(out, encoded)
			}
		}
	}
	return out
}

func parseEndpointSettingElements(raw string) ([]any, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("endpoint dlp setting is empty")
	}
	var parsed any
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, fmt.Errorf("endpoint dlp setting is not JSON: %w", err)
	}
	switch typed := parsed.(type) {
	case []any:
		return typed, nil
	case map[string]any:
		return []any{typed}, nil
	default:
		return nil, fmt.Errorf("endpoint dlp setting must be an object or array, got %T", parsed)
	}
}

func canonicalEndpointSetting(setting map[string]any) (string, bool) {
	normalized := make(map[string]string, len(setting))
	for key, value := range setting {
		name := key
		if strings.EqualFold(key, "Setting") {
			name = "Setting"
		} else if strings.EqualFold(key, "Value") {
			name = "Value"
		}
		text, ok := endpointSettingScalar(value)
		if !ok {
			return "", false
		}
		normalized[name] = text
	}
	encoded, err := json.Marshal(normalized)
	if err != nil {
		return "", false
	}
	return string(encoded), true
}

func endpointSettingScalar(value any) (string, bool) {
	switch typed := value.(type) {
	case string:
		return typed, true
	case bool:
		if typed {
			return "true", true
		}
		return "false", true
	case float64:
		if typed == math.Trunc(typed) {
			return strconv.FormatInt(int64(typed), 10), true
		}
		return strconv.FormatFloat(typed, 'f', -1, 64), true
	default:
		return "", false
	}
}
