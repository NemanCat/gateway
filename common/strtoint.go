//функция преобразования строки к целому числу

package common

import (
	"strconv"
)

func StrToInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
