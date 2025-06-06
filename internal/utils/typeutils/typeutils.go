package typeutils

import (
	"strconv"
)

func ParseSecretInt64(data []byte) int64 {
	v, _ := strconv.ParseInt(string(data), 10, 64)
	return v
}
