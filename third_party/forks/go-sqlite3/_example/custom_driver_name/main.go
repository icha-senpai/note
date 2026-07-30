package main

import (
	"database/sql"

	_ "github.com/icha-senpai/note/third_party/forks/go-sqlite3"
)

func main() {
	for _, driver := range sql.Drivers() {
		println(driver)
	}
}
