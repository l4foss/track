package main

import (
	"database/sql"
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/mattn/go-sqlite3"
	gosu "github.com/thehowl/go-osuapi"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)

const VERSION string = "0.1beta"

/*
* global variables
 */
var (
	osu    *gosu.Client
	bot    *tg.BotAPI
	db     *sql.DB
	lock   = &sync.Mutex{}
	err    error
	config Config
)

/*
* some formatted strings
 */
var (
	msg string = `New <a href="%s">#%d</a> for <a href=https://osu.ppy.sh/u/%d>%s</a> on %v
Map: <a href="http://osu.ppy.sh/b/%d">%s</a> [%s]
Star: <b>%d</b> BPM: <b>%d</b>
Mods: <b>%s</b> Acc: <b>%.2f%%</b> Rank: <b>%v</b>
Combo: <b>%dx/%dx</b> PP: <b>%.2fpp</b>
-- Group ranking --
<pre>
%s
</pre>`

	//				  #01 | SH  | rinq0 | HDHRDTSONF | 111x/222x  | 95.35%
	msgUser string = `#%2d| %3s | %10s  | %10s       | %4dx/%4dx  | %2.2f%%`
)

func initTrack() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Printf("Tracking: %v\n", config.Users)
	err = initDB()

	if err != nil {
		log.Fatal(err)
	}
	/*
	* creates telegram client and osu client
	 */
	osu = gosu.NewClient(config.Osu)
	bot, err = tg.NewBotAPI(config.Telegram)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Done initializing the track! bot")
}

func genGroupRanking(mapid int) string {
	//get only group members
	users, err := getUsers()
	if err != nil {
		return "--ERROR--"
	}

	var result []gosu.GSScore
	opts := gosu.GetScoresOpts{
		BeatmapID: mapid,
		Username:  "someone",
		Mode:      gosu.ModeOsu,
		Limit:     1,
	}

	for _, user := range users {
		opts.Username = user
		score, err := osu.GetScores(opts)

		if err != nil {
			continue
		}

		result = append(result, score[0])
	}

	return "-.-"

}

func calcAccuracy(play *gosu.GUSScore) float64 {
	//generates accuracy
	total := play.MaxCombo * 300
	got := play.Count300*300 + play.Count100*100 + play.Count50*50

	return float64((got * 100) / total)
}

func genMessage(play *gosu.GUSScore, playername string, index int) {
	opts := gosu.GetBeatmapsOpts{
		BeatmapID: play.BeatmapID,
	}

	beatmap, err := osu.GetBeatmaps(opts)
	if err != nil {
		log.Println("Could not fetch beatmap due to ", err.Error())
		return
	}
	bm := beatmap[0]
	thumb := fmt.Sprintf("https://b.ppy.sh/thumb/%dl.jpg", bm.BeatmapSetID)

	message := fmt.Sprintf(msg, thumb, index+1,
		play.Score.UserID, playername, play.Score.Date,
		play.BeatmapID, bm.Title, bm.Artist,
		bm.DifficultyRating, bm.BPM,
		play.Mods)

	/*
	* sends the message using telegram
	* since this runs concurrently, we should use mutex to lock it
	 */
	lock.Lock()
	resp := tg.NewMessage(config.Broadcast, message)
	bot.Send(resp)

	lock.Unlock()
}

func getTop(playername string, wg *sync.WaitGroup) {
	fmt.Printf("Fetching new score for %s\n", playername)
	defer wg.Done()
	opts := gosu.GetUserScoresOpts{
		Username: playername,
		Mode:     gosu.ModeOsu,
		Limit:    50,
	}

	top, err := osu.GetUserBest(opts)
	if err != nil {
		log.Println(err)
		return
	}

	for index, play := range top {
		/*
		* TODO: calculate and add delay time
		 */
		//t := time.Now().Sub(play.Score.Date.GetTime())
		//fmt.Printf("%v\n", t)

		if time.Now().Sub(play.Score.Date.GetTime()) <
			(time.Duration(config.Interval) * time.Second) {
			genMessage(&play, playername, index)
		}
	}
}

func track() {
	users, err := getUsers()

	if err != nil {
		log.Fatal(err)
	}
	var wg sync.WaitGroup
	for {
		for _, user := range users {
			wg.Add(1)
			go getTop(user, &wg)
		}

		wg.Wait()
		//after this all jobs are done
		time.Sleep(time.Duration(config.Interval) * time.Second)
	}
}

func usage() {
	var txt string = `Track v%s
A bot that tracks for osu! players
--
Usage: %s [option]
Options:
	--conf [config file]         runs the bot with config file
	--init                       creates sample config
	--version                    shows version
	--help                       shows this help
`
	fmt.Printf(txt, VERSION, os.Args[0])
	os.Exit(0)
}

func genConfig() error {
	var conf string = `{
    "osu": "OSU TOKEN",
    "telegram": "TELEGRAM TOKEN",
    "broadcast": BROADCAST_CHATID,
    "interval": 60,
    "trackdb": "data.db",
	"track": ["Cookiezi", "Rafis"]
}`

	err := ioutil.WriteFile("config.sample", []byte(conf), 0644)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	if len(os.Args) == 1 {
		usage()
	}

	switch os.Args[1] {
	case "--help":
		{
			usage()
		}
	case "--init":
		{
			err := genConfig()
			if err != nil {
				log.Println(err)
				os.Exit(-1)
			}
			fmt.Println(`Config file "config.sample" created`)
		}
	case "--version":
		{
			fmt.Printf("Track v%s\nUsing Go %s\n", VERSION, runtime.Version())
		}
	case "--conf":
		{
			if len(os.Args) != 3 {
				fmt.Printf("%s --conf [config file]\n", os.Args[0])
				os.Exit(-1)
			}
			config, err = readConfig(os.Args[2])
			if err != nil {
				log.Fatal(err)
			}
			initTrack()
			track()

		}
	default:
		usage()
	}
}
