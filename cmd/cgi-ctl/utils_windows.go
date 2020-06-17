package main

import (
	"fmt"
)

func AskPass() ([]byte, error) {
	var input string
	_, err := fmt.Scanln(&input)
	return []byte(input), err
}
