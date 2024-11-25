package utils

func Masked(s string, maskedStr string) string {
	if len(maskedStr) == 0 {
		maskedStr = "***"
	}
	runeNames := []rune(s)
	if len(runeNames) <= 1 {
		return s + maskedStr
	} else {
		return string(runeNames[0]) + maskedStr + string(runeNames[len(runeNames)-1])
	}
}
