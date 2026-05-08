package main

import "strings"

func ParseDomain(domain string) (string, string) {
	strs := strings.Split(domain, ".")

	// e.g. example.com
	if len(strs) == 2 {
		return domain, "@"
	}

	return strings.Join(strs[len(strs)-2:], "."), strings.Join(strs[0:len(strs)-2], ".")
}
