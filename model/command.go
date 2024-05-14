package model

type Command struct {
	ID      int    `json:"id"`
	Command string `json:"command"`
	Result  string `json:"result"`
}
