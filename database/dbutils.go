package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	gs "github.com/dzendos/Turing/game"
	_ "github.com/lib/pq"
)

var db *sql.DB

type Config struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DB_name  string `json:"db_name"`
}

func LoadConfiguration(file string) Config {
	var config Config
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}

func init() {
	config := LoadConfiguration("config/database_config/config.json")
	URL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", config.User, config.Password, config.Host, config.Port, config.DB_name)
	var err error
	db, err = sql.Open("postgres", URL)
	if err != nil {
		log.Fatal(err)
	}
}

func add_messages(player *gs.Player, id_session int64) {
	var role string
	switch player.Role {
	case 2:
		role = "host"
		break
	case 3:
		role = "knave"
		break
	case 4:
		role = "knight"
		break
	default:
		// print some error TODO
		return
	}
	for i := 0; i < len(player.History); i++ {
		sql_insert_message := fmt.Sprintf("INSERT INTO messages (id_session, id_player, time_from_start, message, role) VALUES (%d, %d, %d, '%s', '%s')", id_session, player.User.ID, player.History[i].TimeFromTheBeg, player.History[i].Message, role)
		_, err := db.Exec(sql_insert_message)
		if err != nil {
			panic(err)
		}
	}
}

func uploadGame(host, knigth, knave *gs.Player) {
	gamestate := host.State
	// user_id host.user.ID
	// time of beggining gamestate.BegginingDate
	// messages list is player.History[i]
	// There are timeFromTheBeg
	date := gamestate.BegginingDate.Format("yyyy MM dd")
	time := gamestate.BegginingDate.Format("HH mm ss")
	sql_insert_statement := fmt.Sprintf("INSERT INTO game_session (host_id, knight_id, knave_id, data_start, time_start, was_succesfull, was_finished) VALUES (%d, %d, %d, %s, %s, %t, %t)", host.User.ID, knave.User.ID, knave.User.ID, date, time, gamestate.WasGameSuccesfull, gamestate.WasGameFinished)
	sql_res, err := db.Exec(sql_insert_statement)
	id_session, _ := sql_res.LastInsertId()

	if err != nil {
		panic(err)
	}

	add_messages(host, id_session)
	add_messages(knave, id_session)
	add_messages(knigth, id_session)
}
