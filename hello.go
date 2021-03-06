package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
)

//go:embed icon.png
var iconData []byte

//go:embed manifest.json
var manifestData []byte

//go:embed bindings.json
var bindingsData []byte

//go:embed send_form.json
var formData []byte

var apiKey string

var weatherCodes = map[int]string{
	200: "thunderstorm with light rain",
	201: "thunderstorm with rain",
	202: "thunderstorm with heavy rain",
	210: "light thunderstorm",
	211: "thunderstorm",
	212: "heavy thunderstorm",
	221: "ragged thunderstorm",
	230: "thunderstorm with light drizzle",
	231: "thunderstorm with drizzle",
	232: "thunderstorm with heavy drizzle",
	300: "light intensity drizzle",
	301: "drizzle",
	302: "heavy intensity drizzle",
	310: "light intensity drizzle rain",
	311: "drizzle rain",
	312: "heavy intensity drizzle rain",
	313: "shower rain and drizzle",
	314: "heavy shower rain and drizzle",
	321: "shower drizzle",
	500: "light rain",
	501: "moderate rain",
	502: "heavy intensity rain",
	503: "very heavy rain",
	504: "extreme rain",
	511: "freezing rain",
	520: "light intensity shower rain",
	521: "shower rain",
	522: "heavy intensity shower rain",
	531: "ragged shower rain",
	600: "light snow",
	601: "Snow",
	602: "Heavy snow",
	611: "Sleet",
	612: "Light shower sleet",
	613: "Shower sleet",
	615: "Light rain and snow",
	616: "Rain and snow",
	620: "Light shower snow",
	621: "Shower snow",
	622: "Heavy shower snow",
	701: "mist",
	711: "Smoke",
	721: "Haze",
	731: "sand/ dust whirls",
	741: "fog",
	751: "sand",
	761: "dust",
	762: "volcanic ash",
	771: "squalls",
	781: "tornado",
	800: "clear sky",
	801: "few clouds",
	802: "scattered clouds",
	803: "broken clouds",
	804: "overcast cloud",
}

// Types autogenerated on https://mholt.github.io/json-to-go/
type Response struct {
	Coord      Coord     `json:"coord"`
	Weather    []Weather `json:"weather"`
	Base       string    `json:"base"`
	Main       Main      `json:"main"`
	Visibility int       `json:"visibility"`
	Wind       Wind      `json:"wind"`
	Clouds     Clouds    `json:"clouds"`
	Dt         int       `json:"dt"`
	Sys        Sys       `json:"sys"`
	Timezone   int       `json:"timezone"`
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Cod        int       `json:"cod"`
}
type Coord struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
}
type Weather struct {
	ID          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}
type Main struct {
	Temp      float64 `json:"temp"`
	FeelsLike float64 `json:"feels_like"`
	TempMin   float64 `json:"temp_min"`
	TempMax   float64 `json:"temp_max"`
	Pressure  int     `json:"pressure"`
	Humidity  int     `json:"humidity"`
}
type Wind struct {
	Speed float64 `json:"speed"`
	Deg   int     `json:"deg"`
	Gust  float64 `json:"gust"`
}
type Clouds struct {
	All int `json:"all"`
}
type Sys struct {
	Type    int    `json:"type"`
	ID      int    `json:"id"`
	Country string `json:"country"`
	Sunrise int    `json:"sunrise"`
	Sunset  int    `json:"sunset"`
}

func (r *Response) ToMessage() string {
	template := "%s\n##### **%s**\n# **%.2f ??C**\n**Feels Like %.2f ??C. %s**\n**Wind:** %.2f m/s from %d??\n**Pressure:**  %d hPa\n**Humidity:** %d"
	return fmt.Sprintf(
		template,
		time.Now().Format("3:04PM, Jan 6"),
		r.Name,
		r.Main.Temp,
		r.Main.FeelsLike,
		weatherCodes[r.Weather[0].ID],
		r.Wind.Speed,
		r.Wind.Deg,
		r.Main.Pressure,
		r.Main.Humidity,
	)
}

func getWeather(appID, city string) (*Response, error) {
	url := "http://api.openweathermap.org/data/2.5/weather?q=" + city + "&units=metric&appid=" + appID
	method := "GET"

	client := &http.Client{
	}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var weather *Response
	if err := json.Unmarshal(body, &weather); err != nil {
		return nil, err
	}

	return weather, nil
}

func main() {
	var exists bool
	apiKey, exists = os.LookupEnv("WEATHER_API_KEY")

	if !exists {
		fmt.Println("Weather API key not found in environment. Add your OpenWeatherMaps API key in `WEATHER_API_KEY` variable")
		return
	}

	// Serve its own manifest as HTTP for convenience in dev. mode.
	http.HandleFunc("/manifest.json", writeJSON(manifestData))

	// Returns the Channel Header and Command bindings for the App.
	http.HandleFunc("/bindings", writeJSON(bindingsData))

	// The form for sending a Hello message.
	http.HandleFunc("/send/form", writeJSON(formData))

	// The main handler for sending a Hello message.
	http.HandleFunc("/send/submit", send)

	// Forces the send form to be displayed as a modal.
	// TODO: ticket: this should be unnecessary.
	http.HandleFunc("/send-modal/submit", writeJSON(formData))

	// Serves the icon for the App.
	http.HandleFunc("/static/icon.png", writeData("image/png", iconData))

	http.ListenAndServe(":8080", nil)
}

func send(w http.ResponseWriter, req *http.Request) {
	c := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&c)

	v, ok := c.Values["message"]
	if ok && v != nil {

		weather, err := getWeather(apiKey, v.(string))
		if err != nil {
			fmt.Println(fmt.Sprintf("Error fetching weather for location: %s, error: %s", v.(string), err.Error()))
			return
		}

		mmclient.AsBot(c.Context).DM(c.Context.ActingUserID, weather.ToMessage())
	}

	json.NewEncoder(w).Encode(apps.CallResponse{})
}

func writeData(ct string, data []byte) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", ct)
		w.Write(data)
	}
}

func writeJSON(data []byte) func(w http.ResponseWriter, r *http.Request) {
	return writeData("application/json", data)
}
