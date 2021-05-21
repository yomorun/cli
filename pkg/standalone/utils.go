package standalone

func TruncateString(str string, maxLength int) string {
	strLength := len(str)
	if strLength <= maxLength {
		return str
	}

	return str[0:maxLength-3] + "..."
}
