package tools

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/truenas/truenas-mcp/mcp"
	"github.com/truenas/truenas-mcp/truenas"
)

// certificateFields are the certificate.query fields passed through when present. Certificate, key and
// CSR PEM material is never in this list.
var certificateFields = []string{
	"name", "cert_type", "common", "san", "from", "until", "key_type", "key_length", "digest_algorithm",
}

// certificateTimeLayouts are the date formats accepted for from/until. The API documents RFC 3339, but
// 25.10.1 was observed returning ctime style text such as "Thu Dec 24 00:43:08 2026" (read as UTC).
var certificateTimeLayouts = []string{time.RFC3339, time.ANSIC}

// registerCertificateTools registers the read-only certificate expiry view.
func registerCertificateTools(r *Registry) {
	r.tools["query_certificates"] = Tool{
		Definition: mcp.Tool{
			Name: "query_certificates",
			Description: "Read-only certificate list (certificate.query): name, type, common name, SANs, validity dates, " +
				"days left until expiry and an expired flag. Certificate, private key and CSR contents are not returned",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		Handler: handleQueryCertificates,
	}
}

// parseCertificateTime reads a certificate date given as an RFC 3339 or ctime style string or as a {"$date": milliseconds}
// object. It reports false for null, missing or unparseable values.
func parseCertificateTime(value interface{}) (time.Time, bool) {
	switch typed := value.(type) {
	case string:
		for _, layout := range certificateTimeLayouts {
			if parsed, err := time.Parse(layout, typed); err == nil {
				return parsed, true
			}
		}
		return time.Time{}, false
	case map[string]interface{}:
		if millis, ok := typed["$date"].(float64); ok {
			return time.UnixMilli(int64(millis)).UTC(), true
		}
	}
	return time.Time{}, false
}

// summarizeCertificates keeps the allowlisted certificate fields and adds days_left (whole days until
// "until", negative once past, nil when until cannot be parsed) and expired. When until cannot be parsed
// the API's own expired flag is used. Entries are sorted by name.
func summarizeCertificates(raw []map[string]interface{}, now time.Time) []map[string]interface{} {
	summary := make([]map[string]interface{}, 0, len(raw))
	for _, certificate := range raw {
		entry := make(map[string]interface{}, len(certificateFields)+2)
		for _, field := range certificateFields {
			if value, ok := certificate[field]; ok {
				entry[field] = value
			}
		}
		if until, ok := parseCertificateTime(certificate["until"]); ok {
			entry["days_left"] = int(math.Floor(until.Sub(now).Hours() / 24))
			entry["expired"] = !until.After(now)
		} else {
			entry["days_left"] = nil
			reported, _ := certificate["expired"].(bool)
			entry["expired"] = reported
		}
		summary = append(summary, entry)
	}
	sort.SliceStable(summary, func(i, j int) bool {
		left, _ := summary[i]["name"].(string)
		right, _ := summary[j]["name"].(string)
		return left < right
	})
	return summary
}

func handleQueryCertificates(client *truenas.Client, args map[string]interface{}) (string, error) {
	certificates, err := callClientList(client, "certificate.query")
	if err != nil {
		return "", err
	}
	formatted, err := json.MarshalIndent(summarizeCertificates(certificates, time.Now()), "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format certificates: %w", err)
	}
	return string(formatted), nil
}
