package models

import (
	"regexp"
)

type RegexCheck struct {
}

//匹配数字开头的节点
func (ic *RegexCheck) IsInteger(str ...string) bool {
	var b bool
	for _, s := range str {
		b, _ = regexp.MatchString("^[0-9]+$|^-[0-9]+$", s)
		if false == b {
			return b
		}
	}
	return b
}