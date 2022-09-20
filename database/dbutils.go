package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var Db *sql.DB

type Config struct {
	Database struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		DB_name  string `json:"dbname"`
	} `json:"db"`

	Bot struct {
		Token string `json:"token"`
	} `json:"bot"`
}

func LoadConfiguration(file string) Config {
	var config Config
	configFile, err := os.Open(file)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer configFile.Close()
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}

func Init() {
	config := LoadConfiguration("config/config.json")
	URL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", config.Database.User, config.Database.Password, config.Database.Host, config.Database.Port, config.Database.DB_name)
	var err error
	Db, err = sql.Open("postgres", URL)
	if err != nil {
		log.Fatal(err)
	}
}
