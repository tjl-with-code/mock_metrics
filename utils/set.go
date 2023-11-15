package utils

import "sort"

type Set map[string]struct{}

func (s *Set) Add(value string) {
	temp := *s
	if _, ok := temp[value]; !ok {
		temp[value] = struct{}{}
	}
}

func (s *Set) List() []string {
	res := make([]string, 0)
	for key, _ := range *s {
		res = append(res, key)
	}
	sort.Strings(res)
	return res
}
