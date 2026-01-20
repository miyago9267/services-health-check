package format

import "strings"

func DetailsList(details string) string {
	if strings.TrimSpace(details) == "" {
		return "n/a"
	}
	lines := strings.ReplaceAll(details, "; ", "\n")
	lines = strings.ReplaceAll(lines, ";", "\n")
	parts := strings.Split(lines, "\n")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, "- "+p)
	}
	return strings.Join(out, "\n")
}
