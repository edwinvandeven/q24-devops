package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/edwinvandeven/q24-devops/office_lights/models"
	"github.com/spf13/cobra"
)

var now = time.Now()

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "office_lights",
	Short: "Determine if office lights should be ON or OFF based on sunrise and sunset times",
	Long:  `Determine if office lights should be ON or OFF based on sunrise and sunset times`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		latitude, _ := cmd.Flags().GetString("latitude")
		longitude, _ := cmd.Flags().GetString("longitude")
		tzid, _ := cmd.Flags().GetString("tzid")
		timeStr, _ := cmd.Flags().GetString("time")
		dateStr, _ := cmd.Flags().GetString("date")
		verbose, _ := cmd.Flags().GetBool("verbose")

		// parse date
		date := getDate(dateStr)

		// get location
		location := getLocation(tzid)

		// Check if we're using the current time or a specified time
		lightsTime := getTime(timeStr, date, location)

		if verbose {
			fmt.Printf("Using: lat=%s, lng=%s, tzid=%s, time=%s\n", latitude, longitude, tzid, timeStr)
		}

		apiResponse := callAPI(latitude, longitude, tzid, dateStr, date, verbose)

		sunrise, err := parseAPITime(apiResponse.Results.Sunrise, date, location)
		if err != nil {
			log.Fatalf("Error parsing sunrise time: %v", err)
		}
		sunset, err := parseAPITime(apiResponse.Results.Sunset, date, location)
		if err != nil {
			log.Fatalf("Error parsing sunset time: %v", err)
		}

		// Output whether lights should be ON or OFF
		lightsOn := lightsShouldBeOn(lightsTime, sunrise, sunset, verbose)
		if lightsOn {
			fmt.Println("ON")
		} else {
			fmt.Println("OFF")
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.office_lights.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.Flags().String("latitude", "50.930581", "Latitude for sunrise/sunset calculation")
	rootCmd.Flags().String("longitude", "5.780691", "Longitude for sunrise/sunset calculation")
	rootCmd.Flags().String("tzid", "Europe/Amsterdam", "Timezone for sunrise/sunset calculation")
	rootCmd.Flags().String("time", now.Format("3:04:05 PM"), "Time for sunrise/sunset calculation ('HH:MM:SS AM/PM')")
	rootCmd.Flags().String("date", now.Format("02-01-2006"), "Date for sunrise/sunset calculation (DD-MM-YYYY)")
	rootCmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
}

// lightsShouldBeOn determines if the lights should be ON or OFF based on the provided times.
func lightsShouldBeOn(lightsTime time.Time, sunrise time.Time, sunset time.Time, verbose bool) bool {
	if verbose {
		fmt.Printf("Lights time: %s\n", lightsTime)
		fmt.Printf("Sunrise: %s\n", sunrise)
		fmt.Printf("Sunset: %s\n", sunset)
	}

	lightOn := true
	if lightsTime.After(sunrise) && lightsTime.Before(sunset) {
		lightOn = false
	}

	return lightOn
}

// Parses the time string from the API and combines it with the provided date and location.
func parseAPITime(timeStr string, date time.Time, location *time.Location) (time.Time, error) {
	parsedTime, err := time.Parse("3:04:05 PM", timeStr)
	if err != nil {
		return time.Time{}, err
	}

	// Combine the parsed time with the date and timezone
	year, month, day := date.Date()
	hour, min, sec := parsedTime.Clock()
	// time.Date() returns time.Time
	return time.Date(year, month, day, hour, min, sec, 0, location), nil
}

// Loads and returns the time.Location for the given timezone ID.
func getLocation(tzid string) *time.Location {
	location, err := time.LoadLocation(tzid)
	if err != nil {
		log.Fatalf("Invalid timezone: %v", err)
	}
	return location
}

// getTime parses the time string in 'HH:MM:SS AM/PM' format and returns a value of type time.Time.
// In case of the current time string, it returns the current time.
func getTime(timeStr string, date time.Time, location *time.Location) time.Time {
	lightsTime := now
	// A custom time has been provided
	if timeStr != now.Format("3:04:05 PM") {
		parsedTime, err := time.Parse("3:04:05 PM", timeStr)
		if err != nil {
			log.Fatal(err)
		}
		lightsTime = parsedTime
	}

	// Convert lightsTime to the same timezone for proper comparison
	// var lightsTimeInZone time.Time
	year, month, day := date.Date()
	hour, min, sec := lightsTime.Clock()
	lightsTimeInZone := time.Date(year, month, day, hour, min, sec, lightsTime.Nanosecond(), location)

	return lightsTimeInZone
}

// getDate parses the date string in 'DD-MM-YYYY' format and returns a value of type time.Times.
// In case of an empty string, it returns the current date.
func getDate(dateStr string) time.Time {
	var err error
	date := now
	// A custom date has been provided
	if dateStr != "" {
		date, err = time.Parse("02-01-2006", dateStr)
		if err != nil {
			log.Panic(err)
		}
	}

	return date
}

// callAPI makes the API request to get sunrise and sunset times.
func callAPI(latitude, longitude, tzid, dateStr string, date time.Time, verbose bool) *models.APIResponse {
	// Make API request
	var url string
	if dateStr != "" {
		url = fmt.Sprintf("https://api.sunrise-sunset.org/json?lat=%s&lng=%s&tzid=%s&date=%s", latitude, longitude, tzid, date.Format("2006-01-02"))
	} else {
		url = fmt.Sprintf("https://api.sunrise-sunset.org/json?lat=%s&lng=%s&tzid=%s", latitude, longitude, tzid)
	}

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}

	// Close the response body when done with the function
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var apiResp *models.APIResponse
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		log.Fatalln(err)
	}

	if verbose {
		fmt.Printf("API Response: %+v\n", apiResp)
	}

	return apiResp
}
