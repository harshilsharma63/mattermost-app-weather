# Mattermost App Weather

#### All you weather queries answered by a bot.

## Setup

#### Obtaining OpenWeatherMap API Key

The app uses free API provided [OpenWeatherMap](https://openweathermap.org/) to obtain weather data. You need to generate a license key in order to use this API.

Follow the steps below.
1. Login or signup for a free account on OpenWeatherMap [here](https://home.openweathermap.org/users/sign_in).
1. Once logged in, nagivate to API Keys page [here](https://home.openweathermap.org/api_keys).
1. You should have a pre-generated API key in there. If not, you can generate a new one by entering a suitable name and clicking on **Generate** button.

#### Setting API key

Export the API key in an environment variable-

```shell
export WEATHER_API_KEY=<api-key>
```

replacing `<api-key>` with the API key you generated in the previous step.

#### Running the Service

Execute the command to get the app up and running-

```shell
go run hello.go
```

#### Installing App in Mattermost

Once the app is up and running after following the previous steps, we need to install it in Matttermost.

Run the following slash commands in any channel-

```
/apps debug-add-manifest --url http://localhost:8080/manifest.json
```

```
/apps install --app-id weatherbot
```

You can add anything in **App-Secret**.

You will be able to use the app now as-

```
/weather city --message <city-name>
```

for example-

```
/weather city --message Delhi
```

You'll get teh weather DM'ed you by the weatherbot -

![](https://i.imgur.com/bq0xZdl.png)


