package main
import(
	"fmt"
	"strings"
	"os"
	"bufio"
)

func cleanInput(text string) []string {
	text = strings.ToLower(text)
	words := strings.Fields(text)
	
	return words
}














func main(){
	scanner := bufio.NewScanner(os.Stdin)


	for{
		fmt.Print("Pokedex > ")
		scanner.Scan()
		input := scanner.Text()
		cleaned := cleanInput(input)
		if len(cleaned) == 0 {continue}
		fmt.Printf("Your command was: %s\n",cleaned[0])








	}



	return
}

