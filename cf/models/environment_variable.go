package models

import "strings"

type EnvironmentVariable struct {
	Name  string
	Value string
}

type EnvironmentVariableList []EnvironmentVariable

func (evl EnvironmentVariableList) Len() int {
	return len(evl)
}

func (evl EnvironmentVariableList) Swap(i, j int) {
	evl[i], evl[j] = evl[j], evl[i]
}

func (evl EnvironmentVariableList) Less(i, j int) bool {
	return strings.Compare(strings.ToLower(evl[i].Name), strings.ToLower(evl[j].Name)) == -1
}
