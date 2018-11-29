package lib

import (
	"strings"
)

// DistinctStr 去重 逗号连接的 字符串
func DistinctStr(v string) string {
	as := strings.Split(v, ",")

	cp := make([]string, 0)
	for _, a := range as {

		found := false
		for _, i := range cp {
			if i == a {
				found = true
				break
			}
		}

		if !found {
			cp = append(cp, a)
		}
	}

	return strings.Join(cp, ",")
}
