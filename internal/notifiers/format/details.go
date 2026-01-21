package format

import (
	"regexp"
	"strings"
)

func DetailsList(details string) string {
	if strings.TrimSpace(details) == "" {
		return "n/a"
	}
	return buildDetailsList(details, "- ", " - ")
}

func DetailsListForSlack(details string) string {
	return buildDetailsList(details, "• ", "    -")
}

var (
	podPattern       = regexp.MustCompile(`\b[[:alnum:]][[:alnum:]-]*\/[[:alnum:]][[:alnum:]\-\.]*\b`)
	domainParenASCII = regexp.MustCompile(`\([^\)\s]+\)`)
	domainParenZH    = regexp.MustCompile(`（[^）\s]+）`)
	domainPattern    = regexp.MustCompile(`\b[a-zA-Z0-9-]+(?:\.[a-zA-Z0-9-]+)+\b`)
)

func highlightDetails(input string) string {
	input = wrapParenDomains(input, domainParenASCII, "(", ")")
	input = wrapParenDomains(input, domainParenZH, "（", "）")
	input = wrapBareDomains(input)
	input = podPattern.ReplaceAllStringFunc(input, func(match string) string {
		if containsLetter(match) {
			return "`" + match + "`"
		}
		return match
	})
	return input
}

func buildDetailsList(details, bullet, subBullet string) string {
	lines := strings.ReplaceAll(details, "; ", "\n")
	lines = strings.ReplaceAll(lines, ";", "\n")
	lines = strings.ReplaceAll(lines, "； ", "\n")
	lines = strings.ReplaceAll(lines, "；", "\n")
	parts := strings.Split(lines, "\n")
	var out []string
	lastWasExample := false
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if text, ok := trimExamplePrefix(p); ok {
			text = highlightDetails(text)
			if text == "" {
				continue
			}
			out = append(out, subBullet+text)
			lastWasExample = true
			continue
		}
		if lastWasExample && isLikelyPodItem(p) {
			out = append(out, subBullet+highlightDetails(p))
			continue
		}
		lastWasExample = false
		out = append(out, bullet+highlightDetails(p))
	}
	if len(out) == 0 {
		return "n/a"
	}
	return strings.Join(out, "\n")
}

func trimExamplePrefix(input string) (string, bool) {
	trimmed := strings.TrimSpace(input)
	if strings.HasPrefix(trimmed, "例:") {
		return strings.TrimSpace(strings.TrimPrefix(trimmed, "例:")), true
	}
	if strings.HasPrefix(trimmed, "例：") {
		return strings.TrimSpace(strings.TrimPrefix(trimmed, "例：")), true
	}
	return "", false
}

func wrapParenDomains(input string, re *regexp.Regexp, left, right string) string {
	return re.ReplaceAllStringFunc(input, func(match string) string {
		inner := strings.TrimSuffix(strings.TrimPrefix(match, left), right)
		if !strings.Contains(inner, ".") {
			return match
		}
		return left + "`" + inner + "`" + right
	})
}

func wrapBareDomains(input string) string {
	matches := domainPattern.FindAllStringIndex(input, -1)
	if len(matches) == 0 {
		return input
	}
	var b strings.Builder
	last := 0
	for _, idx := range matches {
		start, end := idx[0], idx[1]
		if start < last {
			continue
		}
		skip := (start > 0 && input[start-1] == '`') || (end < len(input) && input[end] == '`')
		b.WriteString(input[last:start])
		token := input[start:end]
		if skip {
			b.WriteString(token)
		} else {
			b.WriteString("`")
			b.WriteString(token)
			b.WriteString("`")
		}
		last = end
	}
	b.WriteString(input[last:])
	return b.String()
}

func isLikelyPodItem(input string) bool {
	if strings.Contains(input, " ") {
		return false
	}
	return podPattern.MatchString(input)
}

func containsLetter(s string) bool {
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return true
		}
	}
	return false
}
