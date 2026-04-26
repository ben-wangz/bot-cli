package consolecap

import (
	"regexp"
	"strings"
)

var serialStartupBannerPattern = regexp.MustCompile(`^OKstarting serial terminal on interface serial[0-3]$`)

func isOnlySerialStartupBanner(raw string) bool {
	clean := strings.Join(strings.Fields(normalizeSerialText(raw)), " ")
	if clean == "" {
		return false
	}
	return serialStartupBannerPattern.MatchString(clean)
}

func tailText(raw string, max int) string {
	if max <= 0 || len(raw) <= max {
		return raw
	}
	return raw[len(raw)-max:]
}

var serialANSIEscapePattern = regexp.MustCompile("\x1b\\[[0-9;?]*[ -/]*[@-~]|\x1b[@-Z\\-_]")

func normalizeSerialText(raw string) string {
	clean := serialANSIEscapePattern.ReplaceAllString(raw, "")
	clean = strings.ReplaceAll(clean, "\r", "")
	clean = strings.ReplaceAll(clean, "\x00", "")
	return clean
}

func matchesExpect(raw string, expect string) bool {
	if expect == "" {
		return len(raw) > 0
	}
	if strings.Contains(raw, expect) {
		return true
	}
	return strings.Contains(normalizeSerialText(raw), expect)
}
