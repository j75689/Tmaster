package parser

func RemoveDoubleQuotes(s string) string {
	if len(s) > 2 && s[:1] == `"` && s[len(s)-1:] == `"` {
		return s[1 : len(s)-1]
	}
	return s
}
