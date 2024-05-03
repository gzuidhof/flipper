package hetzner

import "strconv"

func resourceIndexFromLabel(labels map[string]string) int {
	if labels == nil {
		return -1
	}

	if index, ok := labels["resource_index"]; ok {
		v, err := strconv.ParseInt(index, 10, 64)
		if err != nil {
			return -1
		}
		return int(v)
	}

	return -1
}
