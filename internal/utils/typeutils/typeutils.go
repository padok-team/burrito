package typeutils

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseSecretInt64(data []byte) int64 {
	v, _ := strconv.ParseInt(string(data), 10, 64)
	return v
}

func SanitizePrefix(prefix string) string {
	trimmedPrefix := strings.TrimPrefix(prefix, "/")
	trimmedPrefix = strings.TrimSuffix(trimmedPrefix, "/")
	trimmedPrefix = fmt.Sprintf("%s/", trimmedPrefix)

	return trimmedPrefix
}
