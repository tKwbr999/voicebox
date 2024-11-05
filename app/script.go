package app

import "strings"

func Trim(s string) string {
	s = strings.TrimSpace(s)
	// 全角スペースをトリム
	s = strings.Replace(s, "　", "", -1)

	return s
}
