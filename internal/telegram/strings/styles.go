package strings

import (
	"html"
)

func Bold(s string) string {
	return "<b>" + s + "</b>"
}

func Quote(s string) string {
	return "<blockquote>" + html.EscapeString(s) + "</blockquote>"
}
