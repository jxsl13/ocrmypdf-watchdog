package config

func contains(element int, in []int) bool {
	for _, s := range in {
		if s == element {
			return true
		}
	}
	return false
}
