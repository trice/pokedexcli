package main

import (
	"fmt"
	"strings"
)

func cleanInput(text string) []string {
	if len(text) == 0 {
		return []string{}
	}

	words := strings.Fields(text)

	for i := range words {
		words[i] = strings.ToLower(words[i])
	}

	return words
}

func main()  {
	for _, v := range cleanInput(" HELLO     World ") {
		fmt.Println(v)
	}
}
