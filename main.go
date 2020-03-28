package main

import (
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"log"
	"net/http"
	"os"
)

type Application struct {
	Db Db
}

type Db struct {
	AddrMysql     string
	UserMysql     string
	PasswordMysql string
	Database      string
}

var conf = &Application{}

type Database struct {
	db *sqlx.DB
}

func New(user string, password string, addr string, database string) (*Database, error) {
	dataSourceName := user + ":" + password + "@tcp(" + addr + ")/" + database + "?parseTime=true"

	db, err := sqlx.Open("mysql", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	return &Database{db: db}, nil
}

func main() {

	file, err := os.Open("/Users/natalizhylina/src/github.com/natalizhy/list_exemple/config.json")
	if err != nil {
		fmt.Println(err)
	}
	decoder := json.NewDecoder(file)

	err = decoder.Decode(conf)
	if err != nil {
		fmt.Println(err)
	}

	database, err := New(conf.Db.UserMysql, conf.Db.PasswordMysql, conf.Db.AddrMysql, conf.Db.Database)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(database)
	http.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "About Page")
	})
	http.HandleFunc("/contact", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Contact Page")
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//fmt.Fprint(w, "Index Page")
		http.ServeFile(w, r, "index.html")

	})
	fmt.Println("Server is listening...")
	http.ListenAndServe(":8181", nil)
}
