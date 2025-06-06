package utils

import (
	"fmt"
	"strings"
)

func SanitizePrefix(prefix string) string {
	trimmedPrefix := strings.TrimPrefix(prefix, "/")
	trimmedPrefix = strings.TrimSuffix(trimmedPrefix, "/")
	trimmedPrefix = fmt.Sprintf("%s/", trimmedPrefix)

	return trimmedPrefix
}
