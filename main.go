package main
import(
	"fmt"
	"strings"
	"os"
	"bufio"
	"encoding/json"
	"net/http"
	"io"
	"time"
	"github.com/johannesalke/pokedex/internal/pokecache"
	"github.com/johannesalke/pokedex/internal/APIparsing"
	"math/rand"

)

type cliCommand struct {
	name        string
	description string
	callback    func(*Config, ...string) error
}

type Config struct {
	mapCount int
	nextMapUrl string
	prevMapUrl string
	cache pokecache.Pokecache
	pokedex map[string]Pokemon
}

type location struct{
	Name string `json:"name"`
	Url string `json:"url"`
}
type locationResponse struct{
	Count int `json:"count"`
	Next string	`json:"next"`
	Previous string	`json:"previous"`
	Results []location	`json:"results"`
}
/*
type Pokedex struct{
	CaughtPokemon 
}
*/
type Pokemon struct{
	Name string
	Height int
	Weight int
	Stats []struct {
		BaseStat int `json:"base_stat"`
		Effort   int `json:"effort"`
		Stat     struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Slot int `json:"slot"`
		Type struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"type"`
	}
}
/*
type Pokestat struct{
		BaseStat int `json:"base_stat"`
		Effort   int `json:"effort"`
		Stat     struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"stat"`
} 

type Poketype struct{
	
	Slot int `json:"slot"`
	Type struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"type"`
	 
}

*/







func cleanInput(text string) []string {
	text = strings.ToLower(text)
	words := strings.Fields(text)
	
	return words
}




func main(){
	cfg := Config{
		nextMapUrl: "http://pokeapi.co/api/v2/location-area/?limit=20&offset=0", 
		prevMapUrl: "",
		cache: pokecache.NewCache(120*time.Second),
		pokedex: make(map[string]Pokemon),
	
	}


	
	
	scanner := bufio.NewScanner(os.Stdin)
	commands := getCommands()

	for{
		fmt.Print("Pokedex > ")
		scanner.Scan()
		rawInput := scanner.Text()
		input := cleanInput(rawInput)
		if len(input) == 0 {continue}
		//fmt.Printf("Your command was: %s\n",input[0])

		

		command, exists := commands[input[0]]
		if exists  {
			err := command.callback(&cfg,input...)
			if err != nil {fmt.Printf("callback error: %v",err)}
		} else {
			fmt.Print("Unknown command\n")
			
		}



	}



	return
}


func getCommands() map[string]cliCommand {
return map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name: "help",
			description: "Displays a help message",
			callback: commandHelp,
		},
		"map": {
			name: "map",
			description: "Lists the next 20 locations in the pokemon world",
			callback: commandMap,
		},
		"mapb": {
			name: "mapb",
			description: "Lists the previous 20 locations in the pokemon world",
			callback: commandMapB,
		},
		"explore": {
			name:"explore <location>",
			description: "Explore a location specified by the second word entered.",
			callback: commandExplore,
		},
		"catch":{
			name: "catch <pokemon>",
			description: "Attempt to catch a specified pokemon.",
			callback: commandCatch,
		},
		"inspect":{
			name: "inspect <pokemon>",
			description: "Inspect the basic properties of a pokemon you have caught",
			callback: commandInspect,
		},
	}
	}



//Callbacks

func commandExit(cfg *Config, args ...string) error {
	fmt.Print("Closing the Pokedex... Goodbye!\n")
	os.Exit(0)
	return nil
}
func commandHelp(cfg *Config, args ...string) error {
	fmt.Print("Welcome to the Pokedex!\nUsage:\n\n")
	commands := getCommands()
	for k := range commands {
	fmt.Printf("%s: %s\n",k,commands[k].description)

	}
	return nil
			
}

func commandMap(cfg *Config, args ...string) error {
	url := cfg.nextMapUrl 
	
	var bodyContents []byte
	bodyContents, exists := cfg.cache.Get(url)
	if exists {

		fmt.Print("=============\nRetrieved from Cache\n===================\n")
	} else { //If it was cached
	res, err := http.Get(url)
	if err != nil {return err}
	
	//fmt.Printf("%v\n",res.Body)
	
	
	defer res.Body.Close()
	bodyContents,err = io.ReadAll(res.Body)
	if err != nil {return err}
	//fmt.Printf("%v\n",string(bodyContents))
	cfg.cache.Add(url,bodyContents)
	} //If it has to be gotten via API.

	var resp locationResponse
	err := json.Unmarshal(bodyContents, &resp)
	if err != nil { return err}
    //fmt.Printf("%v\n",resp)

	locations := resp.Results
	for _,location := range locations{
		fmt.Println(location.Name)//,"\n"
	}
	cfg.nextMapUrl = resp.Next
	cfg.prevMapUrl = resp.Previous

	return nil

}

func commandMapB(cfg *Config, args ...string) error {
	url := cfg.prevMapUrl
	if url == "" {return fmt.Errorf("It's impossible to go further back!\n")}
	
	var bodyContents []byte
	bodyContents, exists := cfg.cache.Get(url)
	if exists {

		fmt.Print("===================\nRetrieved from Cache\n===================\n")
	} else { //If it was cached
	res, err := http.Get(url)
	if err != nil {return err}
	
	//fmt.Printf("%v\n",res.Body)
	
	
	defer res.Body.Close()
	bodyContents,err = io.ReadAll(res.Body)
	if err != nil {return err}
	//fmt.Printf("%v\n",string(bodyContents))
	cfg.cache.Add(url,bodyContents)
	} //If it has to be gotten via API.

	var resp locationResponse
	err := json.Unmarshal(bodyContents, &resp)
	if err != nil { return err}
    //fmt.Printf("%v\n",resp)

	locations := resp.Results
	for _,location := range locations{
		fmt.Println(location.Name)//,"\n"
	}

	cfg.nextMapUrl = resp.Next
	
	cfg.prevMapUrl = resp.Previous
	return nil

}

func commandExplore(cfg *Config, args ...string) error{
	location := args[1]
	baseurl := "https://pokeapi.co/api/v2/location-area/"
	full_url := baseurl + location

	var bodyContents []byte
	bodyContents, exists := cfg.cache.Get(full_url)
	if exists {

		fmt.Print("=============\nRetrieved from Cache\n===================\n")
	} else { //If it was cached
	res, err := http.Get(full_url)
	if err != nil {return err}
	
	//fmt.Printf("%v\n",res.Body)
	
	
	defer res.Body.Close()
	bodyContents,err = io.ReadAll(res.Body)
	if err != nil {return err}
	//fmt.Printf("%v\n",string(bodyContents))
	cfg.cache.Add(full_url,bodyContents)
	} //If it has to be gotten via API.

	var resp APIparsing.ExploreResponse
	err := json.Unmarshal(bodyContents, &resp)
	if err != nil { return err}
    //fmt.Printf("%v\n",resp)
	
	encounters := resp.PokemonEncounters

	

	fmt.Printf("Exploring %s...\n",location)

	for _,encounter := range encounters{
		fmt.Println(encounter.Pokemon.Name)//,"\n"
	}

	return nil
}

func commandCatch(cfg *Config, args ...string) error{
	pokemon := args[1]
	baseurl := "https://pokeapi.co/api/v2/pokemon/"
	full_url := baseurl + pokemon

	var bodyContents []byte
	bodyContents, exists := cfg.cache.Get(full_url)
	if exists {

		fmt.Print("=============\nRetrieved from Cache\n===================\n")
	} else { //If it was cached
	res, err := http.Get(full_url)
	if err != nil {return err}
	
	//fmt.Printf("%v\n",res.Body)
	
	
	defer res.Body.Close()
	bodyContents,err = io.ReadAll(res.Body)
	if err != nil {return err}
	//fmt.Printf("%v\n",string(bodyContents))
	cfg.cache.Add(full_url,bodyContents)
	} //If it has to be gotten via API.

	var resp APIparsing.PokemonResponse
	err := json.Unmarshal(bodyContents, &resp)
	if err != nil { return err}
    //fmt.Printf("%v\n",resp)
	
	baseExperience := resp.BaseExperience

	fmt.Print("Throwing a Pokeball at squirtle...\n")
	if rand.Intn(baseExperience+100) > baseExperience {
		fmt.Printf("You have caught %v!\n",pokemon)
		cfg.pokedex[pokemon] = Pokemon{
			Name: resp.Name,
			Weight: resp.Weight,
			Height: resp.Height,
			Types: resp.Types,
			Stats: resp.Stats,

		}

	} else {
		fmt.Printf("But %v escaped...\n",pokemon)
	}
	


	return nil
}

func commandInspect(cfg *Config, args ...string) error{
	pokemon := args[1]
	pokeData,exists := cfg.pokedex[pokemon]
	if !exists{return fmt.Errorf("You have not yet caught this pokemon!")}

	fmt.Printf("Name: %s\n",pokeData.Name)
	fmt.Printf("Height: %d\n",pokeData.Height)
	fmt.Printf("Weight: %d\n",pokeData.Weight)
	fmt.Print("Stats:\n")
	Stats := pokeData.Stats
	for _,stat := range Stats{
		fmt.Printf("-%s: %d\n",stat.Stat.Name,stat.BaseStat)
	}
	fmt.Print("Types:\n")
	Types := pokeData.Types
	for _,pType := range Types{
		fmt.Printf("-%s\n",pType.Type.Name)
	}



	/*
	Stats []struct {
		BaseStat int `json:"base_stat"`
		Effort   int `json:"effort"`
		Stat     struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"stat"`
	} `json:"stats"`

			Types []struct {
		Slot int `json:"slot"`
		Type struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"type"`
	}
	*/




	


	return nil
}


