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
    "math/rand"

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
    CmdArgs []string
    CapturedPokemon *map[string]pokemonExperience
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

type exploreResponse struct {
	EncounterMethodRates []struct {
		EncounterMethod struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"encounter_method"`
		VersionDetails []struct {
			Rate    int `json:"rate"`
			Version struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"encounter_method_rates"`
	GameIndex int `json:"game_index"`
	ID        int `json:"id"`
	Location  struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"location"`
	Name  string `json:"name"`
	Names []struct {
		Language struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"language"`
		Name string `json:"name"`
	} `json:"names"`
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
		VersionDetails []struct {
			EncounterDetails []struct {
				Chance          int   `json:"chance"`
				ConditionValues []any `json:"condition_values"`
				MaxLevel        int   `json:"max_level"`
				Method          struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				} `json:"method"`
				MinLevel int `json:"min_level"`
			} `json:"encounter_details"`
			MaxChance int `json:"max_chance"`
			Version   struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"pokemon_encounters"`
}

type pokemonExperience struct {
    Name string `json:"name"`
    Base_Experience int `json:"base_experience"`
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
        "explore": {
            name: "explore",
            description: "Explore a specified location to learn the Pokemon found there",
            callback: commandExplore,
        },
        "catch": {
            name: "catch",
            description: "Catch a specified Pokemon",
            callback: commandCatch,
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

func makeExploreApiCall(apiUrl string) (exploreResponse, error) {
	response := exploreResponse{}
	bodyBytes := []byte{}

	if cr, err1 := http.Get(apiUrl); err1 == nil {
        bbl, err2 := io.ReadAll(cr.Body)
        if err2 != nil {
            return exploreResponse{}, err2
        }
        bodyBytes = bbl
	} else {
	    return exploreResponse{}, err1
	}

	err := json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return exploreResponse{}, err
	}

	return response, nil
}

func makeCatchApiCall(apiUrl string) (pokemonExperience, error) {
    response := pokemonExperience{}
    bodyBytes := []byte{}

	if cr, err1 := http.Get(apiUrl); err1 == nil {
        bbl, err2 := io.ReadAll(cr.Body)
        if err2 != nil {
            return pokemonExperience{}, err2
        }
        bodyBytes = bbl
	} else {
	    return pokemonExperience{}, err1
	}

    err := json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return pokemonExperience{}, err
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

func commandExplore(arg *commandConfig, c *pokecache.Cache) error  {
    // assume area of exploration is the first CmdArgs because it's slice of inputs
    areaOfExploration := arg.CmdArgs[0]

    if len(areaOfExploration) == 0 {
       return fmt.Errorf("no area provided for exploration")
    }

    apiUrl := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s/", areaOfExploration)

    response, err := makeExploreApiCall(apiUrl)
    if err != nil {
       return err
    } else {
       for _, pokemon := range response.PokemonEncounters {
            fmt.Println(pokemon.Pokemon.Name)
        }
    }
    return nil
}

func commandCatch(arg *commandConfig, c *pokecache.Cache) error {
    apiUrl := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s/", arg.CmdArgs[0])

    response, err := makeCatchApiCall(apiUrl)
    if err != nil {
        return err
    }

    fmt.Printf("Throwing a Pokeball at %s...\n", response.Name)

    if response.Base_Experience < 100 {
        if (rand.Intn(99)+1) >= (100 - response.Base_Experience) {
            fmt.Printf("%s was caught!\n", response.Name)
            (*arg.CapturedPokemon)[response.Name] = response
        } else {
            fmt.Printf("%s escaped\n", response.Name)
        }
    } else if response.Base_Experience >= 100 {
        if (rand.Intn(99)+1) >= (100 - (response.Base_Experience-100)) {
            fmt.Printf("%s was caught!\n", response.Name)
            (*arg.CapturedPokemon)[response.Name] = response
        } else {
            fmt.Printf("%s escaped\n", response.Name)
        }
    }
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
	cmdCfg := commandConfig { "https://pokeapi.co/api/v2/location-area/", nil, nil, nil }

	theCache := pokecache.NewCache(5 * time.Second)
	scanner := bufio.NewScanner(os.Stdin)

    capturedPokemon := map[string]pokemonExperience{}
    cmdCfg.CapturedPokemon = &capturedPokemon

	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		words := cleanInput(scanner.Text())
		if cmd, ok := getCommands()[words[0]]; ok {
            if len(words) > 1 {
                cmdCfg.CmdArgs = words[1:]
            }
			err := cmd.callback(&cmdCfg, &theCache)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println("Unknown command")
		}
	}
}
