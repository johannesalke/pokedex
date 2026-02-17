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

func cleanInput(text string) []string {
	text = strings.ToLower(text)
	words := strings.Fields(text)
	
	return words
}



 



func main(){
	cfg := Config{
		nextMapUrl: "http://pokeapi.co/api/v2/location-area/?limit=20&offset=0", 
		prevMapUrl: "",
		cache: pokecache.NewCache(120*time.Second)}


	
	
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