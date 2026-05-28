package assembler

import (
	"regexp"
	"strings"
)

var (
	reComment = regexp.MustCompile(`^(.*?);.*`)
	reDefine  = regexp.MustCompile(`^define\s+(\w+)\s+(\S+)`)
)

// sanitize strips trailing comments and surrounding whitespace from a line.
func sanitize(line string) string {
	if m := reComment.FindStringSubmatch(line); m != nil {
		line = m[1]
	}
	return strings.TrimSpace(line)
}

// preprocess sanitizes every line in place and extracts `define` directives
// into the returned symbol table. Directive lines are blanked.
func preprocess(lines []string) map[string]string {
	table := map[string]string{}
	for i := range lines {
		lines[i] = sanitize(lines[i])
		if m := reDefine.FindStringSubmatch(lines[i]); m != nil {
			if _, exists := table[m[1]]; !exists {
				table[m[1]] = sanitize(m[2])
			}
			lines[i] = ""
		}
	}
	return table
}
