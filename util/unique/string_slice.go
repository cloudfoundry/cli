package unique

func StringSlice(input []string) []string {
	result := make([]string, 0, len(input))
	seen := make(map[string]struct{})
	for _, v := range input {
		if _, ok := seen[v]; !ok {
			result = append(result, v)
			seen[v] = struct{}{}
		}
	}
	return result
}
