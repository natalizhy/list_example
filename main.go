package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"log"
	"os"
	"html/template"
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

type Link struct {
	ID     int    `json:"id"`
	Url    string `json:"url"`
	Status int    `json:"status"`
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


func AddLink(sql *Database, video string) (ID int64, err error) {

	res, err := sql.db.Exec("INSERT INTO `advertising_link` (`url`, `status`) "+
		"VALUES (?, ?)", video, 1)

	if err != nil {
		return 0, fmt.Errorf("AddLink: %w", err)
	}

	ID, err = res.LastInsertId()
	fmt.Println(ID)
	if err != nil {
		return 0, fmt.Errorf("AddLink: %w", err)
	}
	return
}

func GetLink(sql *Database) (user []Link, err error) {
	rows, err := sql.db.Query("SELECT `id`, `url`, `status` "+
		"FROM `advertising_link` WHERE status=?", 1)
	if err != nil {
		err = fmt.Errorf("GetUser: %w", err)
	}

	defer rows.Close()
	user = []Link{}

	for rows.Next() {
		p := Link{}
		err := rows.Scan(&p.ID, &p.Url, &p.Status)
		if err != nil {
			fmt.Println(err)
			continue
		}
		user = append(user, p)
	}
	for _, p := range user {
		fmt.Println(p.ID, p.Url, p.Status, "записи db")
	}
	return
}

func DeleteLink(sql *Database, ID string) (err error) {
	_, err = sql.db.Exec("UPDATE `advertising_link` SET `status`=? "+
		"WHERE `id`=?", 0, ID)

	if err != nil {
		log.Println(err)
	}
	return
}

func UpdateLink(sql *Database, ID, url string) (err error) {
	_, err = sql.db.Exec("UPDATE `advertising_link` SET `url`=? "+
		"WHERE `id`=?", url, ID)
	if err != nil {
		log.Println(err)
	}
	return
}

func GetAdvertising(ctx *gin.Context, database *Database) error {

	if ctx.Request.Method == "POST" {
		url := ctx.PostForm("link")
		fmt.Println(url)
		_, err := AddLink(database, url)
		if err != nil {
			fmt.Println(err)
		}
	}

	link, err := GetLink(database)
	if err != nil {
		fmt.Println(err)
	}
	//var result string
	var arr []Link

	for _, v := range link {
		arr = append(arr, v)
	}

	t, err := template.ParseFiles("index.html")
	if err != nil {
		fmt.Println(err)
	}

	err = t.Execute(ctx.Writer, arr)
	if err != nil {
		fmt.Println(err)
	}

	return nil
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

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.GET("/video/list", func(ctx *gin.Context) {
		err := GetAdvertising(ctx, database)
		if err != nil {
			fmt.Println(err)
		}


	})

	//r.POST("/video/list", func(ctx *gin.Context) {
	//	url := ctx.PostForm("link")
	//	fmt.Println(url)
	//	_, err = AddLink(database, url)
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	//})

	r.POST("/video/list/edit", func(ctx *gin.Context) {
		st, bol := ctx.GetPostForm("url")
		fmt.Println(st, bol, "DateType")
		id := ctx.PostForm("id")
		url := ctx.PostForm("url") // не получаю url
		url2 := ctx.Param("url") // не получаю url
		fmt.Println(id, url, url2, "DateType")
		err = UpdateLink(database, id, url)
		if err != nil {
			fmt.Println(err)
		}
	})
	r.POST("/video/list/delete", func(ctx *gin.Context) {
		id := ctx.PostForm("id")

		err = DeleteLink(database, id)
		if err != nil {
			fmt.Println(err)
		}
	})

	r.GET("/", func(ctx *gin.Context) {

		ctx.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")

		link, err := GetLink(database)
		if err != nil {
			fmt.Println(err)
		}

		for i := 0; i < len(link); i++ {
			var url string
			if i != len(link)-1 {
				url = link[i].Url + ","
			} else {
				url = link[i].Url
			}
			_, err := io.WriteString(ctx.Writer, url)
			if err != nil {
				fmt.Println(err)
			}
		}

	})
	fmt.Println("Server is listening...")
	err = r.Run(":8181")
	if err != nil {
		fmt.Println(err)
	}
}
