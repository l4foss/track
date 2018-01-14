package main

/*
* database scheme
* |group|playername|oldAcc|oldPP|oldRank|oldCRank|
* |bool | string   |float |float|int    |int     |
 */
import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	gosu "github.com/thehowl/go-osuapi"
	"log"
)

type User struct {
	group      int
	playername string
	id         int
	oldAcc     float64
	oldPP      float64
	oldRank    int
	oldCRank   int
}

/*
* create new database in case of first time
 */
func initDB() error {
	db, err = sql.Open("sqlite3", config.TrackDB)
	if err != nil {
		return err
	}

	//try to get all players
	players, err := getPlayers()

	if err != nil || len(players) == 0 {
		/*
		* create new
		 */
		stmt := "CREATE TABLE \"track\" ( `group` INTEGER, `username` TEXT, `id` INTEGER, `oldAcc` float, `oldPP` float, `oldRank` INTEGER, `oldCRank` INTEGER, PRIMARY KEY(`username`) )"
		log.Println("Database does not exist, creating new ...")
		_, err := db.Exec(stmt)

		if err != nil {
			return err
		}
	}
	return nil
}

/*
* get all players
 */
func getPlayers() ([]string, error) {
	rows, err := db.Query(`SELECT playername FROM track`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []string
	var user string
	for rows.Next() {
		//err = rows.Scan(&user.group, &user.playername, &user.id,
		//	&user.oldAcc, &user.oldPP, &user.oldRank, &user.oldCRank)

		err = rows.Scan(&user)
		if err == nil {
			result = append(result, user)
		} else {
			return nil, err
		}
	}
	return result, nil
}

/*
* get all players who are not in group
* that means we don't include them in group ranking
 */
func getUsers() ([]string, error) {
	rows, err := db.Query(`SELECT playername FROM track WHERE group=$`, 1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []string
	var user string
	for rows.Next() {
		//err = rows.Scan(&user.group, &user.playername, &user.id,
		//	&user.oldAcc, &user.oldPP, &user.oldRank, &user.oldCRank)

		err = rows.Scan(&user)
		if err == nil {
			result = append(result, user)
		} else {
			return nil, err
		}
	}
	return result, nil
}

/*
* returns info of user
 */
func getInfo(playername string) (User, error) {
	var user User
	row := db.QueryRow(`SELECT * FROM track WHERE playername=$1`, playername)

	if err != nil {
		return user, err
	}

	err = row.Scan(&user.group, &user.playername, &user.id,
		&user.oldAcc, &user.oldPP, &user.oldRank, &user.oldCRank)

	if err != nil {
		return user, err
	} else {
		return user, nil
	}
}

/*
 * updates user info
 */
func updateInfo(user *gosu.User) error {
	stmt, err := db.Prepare(`UPDATE track
		SET id=?,
		oldAcc=?,
		oldPP=?,
		oldRank=?,
		oldCRank=?
		WHERE playername=?`)

	if err != nil {
		return err
	}

	_, err = stmt.Exec(user.UserID,
		user.Accuracy,
		user.PP,
		user.Rank,
		user.CountryRank,
		user.Username)
	if err != nil {
		return err
	}
	return nil
}
