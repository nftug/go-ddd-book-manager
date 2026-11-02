package telemetry

import (
	"net/url"
	"strings"
)

// redactQuery preserves ordinary query values and masks credentials, including repeated keys.
func redactQuery(raw string) string {
	parts := strings.Split(strings.ReplaceAll(raw, ";", "&"), "&")
	for i, part := range parts {
		key, _, _ := strings.Cut(part, "=")
		decoded, err := url.QueryUnescape(key)
		if err != nil {
			return "[REDACTED]"
		}
		switch strings.ToLower(decoded) {
		case "access_token", "token", "password", "client_secret":
			parts[i] = key + "=REDACTED"
		}
	}
	return strings.Join(parts, "&")
}
