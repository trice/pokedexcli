package main

import (
	"bufio"
	"fmt"
	"os"
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
	scanner := bufio.NewScanner(os.Stdin)
	for ;; {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		words := cleanInput(scanner.Text())
		
		if len(words) > 0 {
			fmt.Printf("Your command was: %s\n", words[0])
		}
	}
}
