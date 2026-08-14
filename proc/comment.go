package proc

import (
	"strings"
)

// CommentLines comment the line like `// xxxx`
type CommentLines []string

func NewCommentLines(s string) CommentLines {
	if strings.TrimSpace(s) == "" {
		return CommentLines{}
	}
	if !strings.HasPrefix(s, "//") {
		s = "// " + strings.TrimSpace(s)
	}
	return strings.Split(strings.TrimSuffix(s, "\n"), "\n")
}

// Derives all match the derive and remaining comment lines.
func (c CommentLines) Derives() ([]*Derive, CommentLines) {
	remain := make(CommentLines, 0, len(c))
	ret := make([]*Derive, 0, len(c))
	for _, s := range c {
		if m, err := Match(strings.TrimSpace(strings.TrimPrefix(s, "//"))); err != nil {
			remain = append(remain, s)
		} else {
			ret = append(ret, m)
		}
	}
	return ret, remain
}

// FindDerives find `identity` match the annotation and remaining comment lines.
func (c CommentLines) FindDerives(identity string) ([]*Derive, CommentLines) {
	remain := make(CommentLines, 0, len(c))
	ret := make([]*Derive, 0, len(c))
	for _, s := range c {
		if m, err := Match(trimRedundance(s)); err != nil {
			remain = append(remain, s)
		} else if m.Identity == identity {
			ret = append(ret, m)
		}
	}
	return ret, remain
}

func (c *CommentLines) Append(s string) CommentLines {
	if !strings.HasPrefix(s, "//") {
		s = "// " + strings.TrimSpace(s)
	}
	*c = append(*c, s)
	return *c
}

// String formats the comments by inserting // to the start of each line,
// ensuring that there is a trailing newline.
// An empty comment is formatted as an empty string.
func (c CommentLines) String() string {
	if len(c) == 0 {
		return ""
	}
	b := &strings.Builder{}
	for i, line := range c {
		b.WriteString(line)
		if i+1 < len(c) {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// LineString one line string.
// String formats the comments by remove the start //,
// the next multi line to trim the start //, join with `, ` .
// An empty comment is formatted as an empty string.
func (c CommentLines) LineString() string {
	if len(c) == 0 {
		return ""
	}
	b := &strings.Builder{}
	for i, line := range c {
		b.WriteString(trimRedundance(line))
		if i+1 < len(c) {
			b.WriteString(", ")
		}
	}
	return b.String()
}

// trimRedundance prefix `//` and then trimRedundance leading and trailing whitespace.
func trimRedundance(s string) string {
	return strings.TrimSpace(strings.TrimPrefix(s, "//"))
}
