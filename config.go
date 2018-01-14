package main

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Osu       string   `json:"osu"`
	Telegram  string   `json:"telegram"`
	Broadcast int64    `json:"broadcast"`
	Interval  int64    `json:"interval"`
	Limit     int      `json:"limit"`
	TrackDB   string   `json:"trackdb"`
	Users     []string `json:"track"`
}

func readConfig(filename string) (Config, error) {
	var config Config
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, err
	}

	json.Unmarshal(data, &config)
	return config, nil
}

func genConfig() error {
	var conf string = `{
    "osu": "OSU TOKEN",
    "telegram": "TELEGRAM TOKEN",
    "broadcast": BROADCAST_CHATID,
    "interval": 60,
	"limit": 50,
    "trackdb": "data.db",
	"track": ["Cookiezi", "Rafis"]
}`

	err := ioutil.WriteFile("config.sample", []byte(conf), 0644)
	if err != nil {
		return err
	}
	return nil
}
