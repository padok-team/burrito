package typeutils

import (
	"strconv"
)

func ParseSecretInt64(data string) int64 {
	v, _ := strconv.ParseInt(data, 10, 64)
	return v
}
