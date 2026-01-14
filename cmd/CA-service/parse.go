package main

import (
	"fmt"
	"strings"
)

func parseTokenPrincipals(envTokens string) (map[string][]string, error) {
	m := make(map[string][]string)
	if envTokens == "" {
		return m, nil
	}

	entries := strings.Split(envTokens, ";")
	for _, entry := range entries {
		parts := strings.Split(entry, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid token principals format")
		}

		token := parts[0]
		if token == "" {
			return nil, fmt.Errorf("invalid token principals format")
		}
		principals := strings.Split(parts[1], ",")
		m[token] = principals
	}

	return m, nil
}
