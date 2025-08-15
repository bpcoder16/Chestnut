package utils

func RemoveDuplicates[T comparable](list []T) []T {
	seen := make(map[T]struct{})
	var result []T
	for _, item := range list {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}
