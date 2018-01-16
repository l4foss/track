package main

/*
* database scheme
* |group|username|oldAcc|oldPP|oldRank|oldCRank|
* |bool | string   |float |float|int    |int     |
 */
import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	gosu "github.com/thehowl/go-osuapi"
	"log"
	"sync"
	"fmt"
)

var (
	writeLock = &sync.Mutex{}
)

type User struct {
	group      int
	username string
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

	//check if table exists
	exist := existTable("track")

	if exist == true {
		//try to get all users
		users, err := getUsers()
		if err != nil {
			return err
		}

		log.Println("Checking for users in database")
		for _, user := range config.Users {
			log.Println(user)
			exist := existUser(user)

			if exist == false {
				//create new entry in the database
				fmt.Println("User does not exist in database, creating new ...")
			}
		}


		log.Printf("users: %v\n", users)
		return nil
	} else {
		/*
		* create new
		 */
		log.Println("Table does not exist, creating new ...")
		stmt := "CREATE TABLE \"track\" ( `group` INTEGER, `username` TEXT, `id` INTEGER, `oldAcc` float, `oldPP` float, `oldRank` INTEGER, `oldCRank` INTEGER, PRIMARY KEY(`username`) )"
		_, err = db.Exec(stmt)

		if err != nil {
			return err
		}
		return nil
	}
}

/*
* check if a table exists
 */
func existTable(tblname string) bool {
	var name string
	err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, tblname).Scan(&name)

	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		return false
	default:
		return true
	}
}

/*
* check if a user exist
*/
func existUser(username string) bool {
	var name string
	err := db.QueryRow(`SELECT username FROM track WHERE name=?`, username).Scan(&name)

	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		return false
	default:
		return true
	}
}

/*
* get all users
 */
func getUsers() ([]string, error) {
	rows, err := db.Query(`SELECT username FROM track`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []string
	var user string
	for rows.Next() {
		//err = rows.Scan(&user.group, &user.username, &user.id,
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
* add new user to db
* safe to call simultaneously
 */
func addUser(name string, group int) error {
	stmt, err := db.Prepare(`INSERT INTO track (group, username, id, oldAcc, oldPP, oldRank, oldCRank)
	VALUES (?, ?, ?, ?, ?, ?, ?)`)

	/*
	* get info from osu!
	 */
	opts := gosu.GetUserOpts{
		Username:  name,
		Mode:      gosu.ModeOsu,
		EventDays: 7,
	}

	user, err := osu.GetUser(opts)
	if err != nil {
		return err
	}

	writeLock.Lock()
	stmt.Exec(group,
		user.Username,
		user.UserID,
		user.Accuracy,
		user.PP,
		user.Rank,
		user.CountryRank)
	writeLock.Unlock()
	return nil
}

/*
* get all users who are in group
 */
func getGroupUsers() ([]string, error) {
	rows, err := db.Query(`SELECT username FROM track WHERE group=$1`, 1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []string
	var user string
	for rows.Next() {
		//err = rows.Scan(&user.group, &user.username, &user.id,
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
func getInfo(username string) (User, error) {
	var user User
	row := db.QueryRow(`SELECT * FROM track WHERE username=$1`, username)

	if err != nil {
		return user, err
	}

	err = row.Scan(&user.group, &user.username, &user.id,
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
		WHERE username=?`)

	if err != nil {
		return err
	}

	writeLock.Lock()
	_, err = stmt.Exec(user.UserID,
		user.Accuracy,
		user.PP,
		user.Rank,
		user.CountryRank,
		user.Username)
	writeLock.Unlock()
	if err != nil {
		return err
	}
	return nil
}
