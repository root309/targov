package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// EventHandler - EventHandler Structure
type EventHandler struct{}

// messageCreate - Event handler when a new message is received
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Check if the message begins with "! Check to see if the message begins with
	if len(m.Content) > 1 && m.Content[0] == '!' {
		// Get item name (remove leading "!") is removed)
		itemName := m.Content[1:]

		// Pass item name
		price, err := getPriceOfItem(itemName)
		if err != nil {
			fmt.Println("Error getting price:", err)
			s.ChannelMessageSend(m.ChannelID, "Error retrieving item price.")
			return
		}

		// Send item price information to the channel
		_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Price of %s: %s", itemName, price))
		if err != nil {
			fmt.Println("Error sending message:", err)
		}
	}
}


// ready - bot connection event
func ready(s *discordgo.Session, event *discordgo.Ready) {
	fmt.Printf("%s is connected!\n", event.User.Username)
}

func main() {
	godotenv.Load()

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		fmt.Println("No token provided")
		return
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session:", err)
		return
	}

	dg.AddHandler(messageCreate)
	dg.AddHandler(ready)

	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session:", err)
	}

	fmt.Println("Bot is now running. Press CTRL+C to exit.")
	select {} Waiting in an infinite loop
}

// getPriceOfItem - Function to get the price of a given item
func getPriceOfItem(itemName string) (string, error) {
	query := fmt.Sprintf(`
	{
		items(name: "%s") {
			id
			name
			traderPrices {
				trader {
					name
				}
				price
				currency
			}
		}
	}
	`, itemName)
	body := map[string]string{"query": query}
	bodyJSON, _ := json.Marshal(body)

	resp, err := http.Post("https://api.tarkov.dev/graphql", "application/json", bytes.NewBuffer(bodyJSON))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	responseBody, _ := ioutil.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(responseBody, &result)

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Invalid response format")
	}

	items, ok := data["items"].([]interface{})
	if !ok || len(items) == 0 {
		return "", fmt.Errorf("No items found")
	}

	traderPrices, ok := items[0].(map[string]interface{})["traderPrices"].([]interface{})
	if !ok {
		return "", fmt.Errorf("No trader prices found")
	}

	var prices []string
	for _, tp := range traderPrices {
		traderPrice := tp.(map[string]interface{})
		trader := traderPrice["trader"].(map[string]interface{})
		price := traderPrice["price"].(float64)
		currency := traderPrice["currency"].(string)
		prices = append(prices, fmt.Sprintf("Trader: %s, Price: %.0f %s", trader["name"], price, currency))
	}

	return fmt.Sprintf("%s", prices), nil
}

