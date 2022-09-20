package game

import (
	"fmt"

	db "github.com/dzendos/Turing/database"
)

func Add_messages(player *Player, id_session int64) {
	var role string
	switch player.Role {
	case 2:
		role = "host"
	case 3:
		role = "knave"
	case 4:
		role = "knight"
	default:
		// print some error TODO
		return
	}
	for i := 0; i < len(player.History); i++ {
		sql_insert_message := fmt.Sprintf("INSERT INTO messages (id_session, id_player, time_from_start, message, role) VALUES (%d, %d, %d, '%s', '%s')", id_session, player.User.ID, player.History[i].TimeFromTheBeg, player.History[i].Message, role)
		_, err := db.Db.Exec(sql_insert_message)
		if err != nil {
			panic(err)
		}
	}
}

func UploadGame(host, knigth, knave *Player) {
	gamestate := host.State
	dab := db.Db
	// user_id host.user.ID
	// time of beggining gamestate.BegginingDate
	// messages list is player.History[i]
	// There are timeFromTheBeg
	date := gamestate.BegginingDate.Format("2006 01 02")
	time := gamestate.BegginingDate.Format("15:04")
	sql_insert_statement := fmt.Sprintf("INSERT INTO game_session (host_id, knight_id, knave_id, date_start, time_start, was_succesfull, was_finished) VALUES (%d, %d, %d, '%s', '%s', %t, %t) returning id", host.User.ID, knave.User.ID, knave.User.ID, date, time, gamestate.WasGameSuccesfull, gamestate.WasGameFinished)
	var id_session int64
	err := dab.QueryRow(sql_insert_statement).Scan(&id_session)

	if err != nil {
		panic(err)
	}

	Add_messages(host, id_session)
	Add_messages(knave, id_session)
	Add_messages(knigth, id_session)
}
