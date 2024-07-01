package main

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type ApiResponse struct {
	ClientIP string `json:"client_ip"`
	Location string `json:"location"`
	Greeting string `json:"greeting"`
}

type IpGeolocationResponse struct {
	IP        string `json:"ip"`
	Continent string `json:"continent_name"`
	Country   string `json:"country_name"`
	State     string `json:"state_prov"`
	City      string `json:"city"`
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
	TimeZone  struct {
		Name   string  `json:"name"`
		Offset float64 `json:"offset"`
	} `json:"time_zone"`
}

type WeatherResponse struct {
	Main struct {
		Temp float64 `json:"temp"`
	} `json:"main"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func helloHandler(c *fiber.Ctx) error {
	visitorName := c.Query("visitor_name", "Guest")

	clientIP := getClientIP(c)
	if clientIP == "127.0.0.1" || clientIP == "::1" {
		clientIP = "8.8.8.8" // Use a mock IP address for testing
	}

	apiKey := os.Getenv("IP_GEOLOCATION_API_KEY")
	apiUrl := fmt.Sprintf("https://api.ipgeolocation.io/ipgeo?apiKey=%s&ip=%s", apiKey, clientIP)

	response, err := http.Get(apiUrl)
	if err != nil {
		log.Printf("Error fetching location data: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "Failed to fetch location data",
		})
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("Error reading location data: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "Failed to read location data",
		})
	}

	var locationData IpGeolocationResponse
	err = json.Unmarshal(body, &locationData)
	if err != nil {
		log.Printf("Error parsing location data: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "Failed to parse location data",
		})
	}

	location := fmt.Sprintf("%s, %s", locationData.City, locationData.Country)

	if locationData.City == "" && locationData.Country == "" {
		log.Println("Location data is empty")
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "Location data is empty",
		})
	}

	weatherApiKey := os.Getenv("WEATHER_API_KEY")
	weatherApiUrl := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric", locationData.City, weatherApiKey)

	weatherResponse, err := http.Get(weatherApiUrl)
	if err != nil {
		log.Printf("Error fetching weather data: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "Failed to fetch weather data",
		})
	}
	defer weatherResponse.Body.Close()

	weatherBody, err := ioutil.ReadAll(weatherResponse.Body)
	if err != nil {
		log.Printf("Error reading weather data: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "Failed to read weather data",
		})
	}

	var weatherData WeatherResponse
	err = json.Unmarshal(weatherBody, &weatherData)
	if err != nil {
		log.Printf("Error parsing weather data: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "Failed to parse weather data",
		})
	}

	temperature := weatherData.Main.Temp

	greeting := fmt.Sprintf("Hello, %s! The temperature is %.2f degrees Celsius in %s", visitorName, temperature, location)

	responseObject := ApiResponse{
		ClientIP: clientIP,
		Location: location,
		Greeting: greeting,
	}

	return c.JSON(responseObject)
}

func homeHandler(c *fiber.Ctx) error {

	clientIP := getClientIP(c)
	message := "Welcome to the hng stage one task using Golang! and your ip is: " + clientIP

	// Log the welcome message
	log.Println(message)

	// Return the JSON response with the welcome message
	return c.JSON(fiber.Map{
		"message": message,
	})
}

func getClientIP(c *fiber.Ctx) string {
	ip := c.IP() // Fiber's built-in method to get the client IP
	return ip
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	app := fiber.New()

	app.Get("/api/hello", helloHandler)
	app.Get("/", homeHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server listening on port", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
