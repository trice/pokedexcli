package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/trice/pokedexcli/internal/pokecache"
)

type cliCommand struct {
	name        string
	description string
	callback    func(arg *commandConfig, c *pokecache.Cache) error
}

type commandConfig struct {
	Next string
	Previous *string
}

type queryResponse struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous *string`json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

func getCommands() map[string]cliCommand {

	return map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help":	{
			name: "help",
			description: "Displays a help message",
			callback: commandHelp,
		},
		"map": {
			name: "map",
			description: "Displays 20 location areas in the Pokemon world",
			callback: commandMap,
		},
		"mapb": {
			name: "mapb",
			description: "Displays 20 previous location areas in the Pokemon world",
			callback: commandMapBack,
		},
	}
}

func makeApiCall(apiUrl string, c *pokecache.Cache) (queryResponse, error) {
	response := queryResponse{}
	tmp := []byte{}
	
	if cr, ok := c.Get(apiUrl); ok {
		tmp = cr
	} else {
		r, err := http.Get(apiUrl)
		if err != nil {
			return queryResponse{}, err
		}
		dat, err := io.ReadAll(r.Body)
		if err != nil {
			return queryResponse{}, err
		}
		tmp = dat
		c.Add(apiUrl, dat)
	}

	err := json.Unmarshal(tmp, &response)
	if err != nil {
		return queryResponse{}, err
	}

	return response, nil
}

func commandMap(arg *commandConfig, c *pokecache.Cache) error {
	resp, err := makeApiCall(arg.Next, c)
	if err != nil {
		return fmt.Errorf("map error: %w", err)
	}

	for _, loc := range resp.Results {
		fmt.Println(loc.Name)
	}

	arg.Next = resp.Next
	arg.Previous = resp.Previous

	return nil
}

func commandMapBack(arg *commandConfig, c *pokecache.Cache) error {
	if arg.Previous == nil {
		fmt.Println("You are already on the first page")
		return nil
	}

	resp, err := makeApiCall(*arg.Previous, c)
	if err != nil {
		return fmt.Errorf("mapb error: %w", err)
	}

	for _, loc := range resp.Results {
		fmt.Println(loc.Name)
	}

	arg.Next = resp.Next
	arg.Previous = resp.Previous

	return nil
}

func commandHelp(arg *commandConfig, c *pokecache.Cache) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Printf("Usage:\n\n")
	for _, cmd := range getCommands() {
		fmt.Printf("%s\t%s\n", cmd.name, cmd.description)
	}
	return nil
}

func commandExit(arg *commandConfig, c *pokecache.Cache) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

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
	cmdCfg := commandConfig { "https://pokeapi.co/api/v2/location-area/", nil }
	theCache := pokecache.NewCache(5 * time.Second)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		words := cleanInput(scanner.Text())
		if cmd, ok := getCommands()[words[0]]; ok {
			err := cmd.callback(&cmdCfg, &theCache)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println("Unknown command")
		}
	}
}
