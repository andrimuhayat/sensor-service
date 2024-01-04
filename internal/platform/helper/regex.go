package helper

import (
	"regexp"
	"strings"
)

func TrimmedString(string string) string {
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(string, " "))
}
