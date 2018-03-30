package main

import (
	"database/sql"
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/mattn/go-sqlite3"
	gosu "github.com/thehowl/go-osuapi"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)

const VERSION string = "0.1beta"

var REVISION string = ""

/*
* global variables
 */
var (
	osu      *gosu.Client
	bot      *tg.BotAPI
	db       *sql.DB
	teleLock = &sync.Mutex{}
	err      error
	config   Config
)

/*
* some formatted strings
 */
var (
	msgFmt string = `New <a href="%s">#%d</a> for <a href=https://osu.ppy.sh/u/%d>%s</a> on %v
Map: <a href="http://osu.ppy.sh/b/%d">%s</a> [%s]
Star: <b>%d</b> BPM: <b>%d</b>
Mods: <b>%s</b> Acc: <b>%.2f%%</b> Rank: <b>%v</b>
Combo: <b>%dx/%dx</b> PP: <b>%.2fpp</b>
-- Group ranking --
<pre>
%s
</pre>`

	//	                   #01 | Cookiezi   | 1230x| 99.34%
	scoreFmt string = `#%2d| %10s       | %4dx | %2.2f%%\n`
)

/*
* checks database, creates telegram and osu client
 */
func initTrack() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Printf("Tracking: %v\n", config.Users)
	/*
	* creates telegram client and osu client
	 */
	osu = gosu.NewClient(config.Osu)
	bot, err = tg.NewBotAPI(config.Telegram)
	if err != nil {
		log.Fatal(err)
	}
	err = initDB()

	if err != nil {
		log.Fatal(err)
	}
	if config.Verbose == 1 {
		log.Println("Done initializing the track! bot")
	}
}

/*
* generates group ranking on a map
 */
func genGroupRanking(mapid int) string {
	//get only group members
	users, err := getUsers()
	if err != nil {
		return "ERROR"
	}

	var result []gosu.GSScore
	var output string

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
			//user does not have any score on that map
			continue
		}

		result = append(result, score[0])
	}

	//sort result
	ScoreSort(result)

	for index, play := range result {
		output += fmt.Sprintf(scoreFmt, index+1, play.Username, play.Score.MaxCombo,
			calcAccuracy(&play.Score))
	}
	return output
}

/*
* calculates score accuracy
 */
func calcAccuracy(play *gosu.Score) float64 {
	//generates accuracy
	total := play.MaxCombo * 300
	got := play.Count300*300 + play.Count100*100 + play.Count50*50

	return float64((got * 100) / total)
}

/*
* generates telegram message
 */
func genMessage(play *gosu.GUSScore, username string, index int) {
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

	message := fmt.Sprintf(msgFmt, thumb, index+1,
		play.Score.UserID, username, play.Score.Date,
		play.BeatmapID, bm.Title, bm.Artist,
		bm.DifficultyRating, bm.BPM,
		play.Mods)

	/*
	* sends the message using telegram
	* since this runs concurrently, we should use mutex to lock it
	 */
	teleLock.Lock()
	resp := tg.NewMessage(config.Broadcast, message)
	bot.Send(resp)

	teleLock.Unlock()
}

/*
* checks for new top scores and generates
* telegram message
* called concurrently
 */
func getTop(username string, wg *sync.WaitGroup) {
	if config.Verbose == 1 {
		log.Printf("Fetching new score for %s\n", username)
	}
	defer wg.Done()
	opts := gosu.GetUserScoresOpts{
		Username: username,
		Mode:     gosu.ModeOsu,
		Limit:    config.Limit,
	}

	top, err := osu.GetUserBest(opts)
	if err != nil {
		log.Println(err)
		return
	}

	for index, play := range top {
		/*
		* BUG: since this depends on system time, invalid system time
		* mays cause track to work improperly
		* I was pretty naive :<
		 */
		diff := time.Now().Sub(play.Score.Date.GetTime())
		if diff < (time.Duration(config.Interval) * time.Second) {
			genMessage(&play, username, index)
		}
	}
}

/*
* tracks users for each interval of time
 */
func track() {
	var wg sync.WaitGroup
	for {
		for _, user := range config.Users {
			wg.Add(1)
			go getTop(user, &wg)
		}

		wg.Wait()
		//after this all jobs are done
		time.Sleep(time.Duration(config.Interval) * time.Second)
	}
}

/*
* prints usage
 */
func usage() {
	var txt string = `Track v%s
A bot that tracks osu! users for theirs top scores
--
Usage: %s [option]
Options:
	--conf [config file]         runs the bot with config file
	--genconf                    generates sample config
	--version                    shows version
	--help                       shows this help
`
	fmt.Printf(txt, VERSION, os.Args[0])
	os.Exit(0)
}

/*
* main function
 */
func main() {
	if len(os.Args) == 1 {
		usage()
	}

	switch os.Args[1] {
	case "--help":
		{
			usage()
		}
	case "--genconf":
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
			fmt.Printf("Track v%s-%s\nUsing Go %s\n", VERSION, REVISION, runtime.Version())
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
