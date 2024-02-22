package utils

func ConvertToStringSlice[T ~string](status []T) []string {
	strSlice := make([]string, len(status))
	for i, s := range status {
		strSlice[i] = string(s)
	}
	return strSlice
}
