package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	apiURL     = "https://api.coingecko.com/api/v3/coins/bitcoin/history"
	timeStep   = 2070356 // seconds
	initialUSD = 1000
)

var apiKey string

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file")
		os.Exit(1)
	}

	apiKey = os.Getenv("COINGECKO_API_KEY")
	if apiKey == "" {
		fmt.Println("API key not found in .env file")
		os.Exit(1)
	}
}

func getPrice(date string) (string, string, error) {
	url := fmt.Sprintf("%s?date=%s", apiURL, date)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", err
	}
	request.Header.Set("accept", "application/json")
	request.Header.Set("Authorization", "Bearer "+apiKey)

	// fmt.Printf("Getting price\n")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", "", err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusTooManyRequests {
		retryAfter, err := strconv.Atoi(response.Header.Get("Retry-After"))
		if err == nil {
			fmt.Printf("Rate limit exceeded. Waiting %d seconds...\n", retryAfter)
			time.Sleep(time.Duration(retryAfter) * time.Second)
			fmt.Println("Resuming after waiting...")
			return getPrice(date)
		}
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", "", err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", "", err
	}

	currency := "BTC" // CoinGecko API response doesn't include the currency symbol
	price := fmt.Sprintf("%v", data["market_data"].(map[string]interface{})["current_price"].(map[string]interface{})["usd"])

	return currency, price, nil
}

func formatTimestamp(timestamp int64) string {
	return time.Unix(timestamp, 0).Format("02-01-2006")
}

func main() {
	// Set initial timestamp
	timestamp := int64(1612796400) // Replace with your initial timestamp

	totalBTC := 0.0
	usdTotal := 0.0
	now := time.Now().Unix()
	lastPrice := 0.0
	for {
		// Format timestamp in "dd-mm-yyyy" for CoinGecko API
		date := formatTimestamp(timestamp)

		currency, price, err := getPrice(date)
		if err != nil {
			fmt.Println("Error fetching data:", err)
			break
		}

		// Calculate how much 1000 USD would get in Bitcoin
		btcAmount, _ := strconv.ParseFloat(price, 64)
		usdEquivalent := initialUSD / btcAmount
		totalBTC += usdEquivalent

		usdTotal += initialUSD

		fmt.Printf("Timestamp: %s  Currency: %s  Price: %s - 1k USD Equivalent: %.8f BTC\n\n",
			formatTimestamp(timestamp), currency, price, usdEquivalent)

		time.Sleep(time.Duration(5) * time.Second)

		// Increment timestamp by timeStep
		timestamp += timeStep

		// fmt.Printf("End loop\n")

		// Break the loop if the timestamp is older than the current time
		if timestamp > now {
			lastPrice = totalBTC / btcAmount
			break
		}
	}

	// Print the total BTC equivalent for 1000 USD
	fmt.Printf("\nTotal BTC equivalent for $%d: %.8f BTC\n - VALUE: %.2f", int(usdTotal), totalBTC, lastPrice)
}
