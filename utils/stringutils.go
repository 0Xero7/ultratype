package utils

func PadRight(s string, width int) string {
	k := string(s)
	for len(k) < width {
		k += " "
	}
	return k
}
