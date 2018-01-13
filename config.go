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
