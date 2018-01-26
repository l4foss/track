package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

var (
	confSkel string = `{
    "osu": %s,
    "telegram": %s,
    "broadcast": %d,
    "interval": %d,
	"limit": %d,
    "trackdb": %s,
	"verbose": %d,
	"track": %s
}`
)

type Config struct {
	Osu       string   `json:"osu"`
	Telegram  string   `json:"telegram"`
	Broadcast int64    `json:"broadcast"`
	Interval  int64    `json:"interval"`
	Limit     int      `json:"limit"`
	TrackDB   string   `json:"trackdb"`
	Verbose   int      `json:"verbose"`
	Users     []string `json:"track"`
}

/*
* reads config from a json file into Config struct
 */
func readConfig(filename string) (Config, error) {
	var config Config
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, err
	}

	json.Unmarshal(data, &config)
	return config, nil
}

/*
* generates sample config file
 */
func genConfig() error {
	conf := fmt.Sprintf(confSkel, "[OSU TOKEN]",
		"[TELEGRAM TOKEN]", 11111111, 60, 50, "track.db", 1, "[\"Cookiezi\"]")

	err := ioutil.WriteFile("config.sample", []byte(conf), 0644)
	if err != nil {
		return err
	}
	return nil
}

/*
* this function writes current config to config file
* this could be useful when you change the database at runtime and want
* user list to be synced with config file
 */
func writeConfig() error {
	users, err := getUsers()
	if err != nil {
		return err
	}

	ulist := fmt.Sprintf("%v", users)
	conf := fmt.Sprintf(confSkel, config.Osu, config.Telegram, config.Broadcast, config.Interval,
		config.Limit, config.TrackDB, config.Verbose, ulist)

	err = ioutil.WriteFile(os.Args[2], []byte(conf), 0644)
	if err != nil {
		return err
	}

	return nil
}
