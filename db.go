package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" // подключение драйвера PostgreSQL
)

var db *sql.DB // глобальная переменная

func connectDB() {
	var err error
	connStr := "host=localhost port=5432 user=postgres dbname=new sslmode=disable" // подключение к БД
	db, err = sql.Open("postgres", connStr)                                        // подключение с БД
	if err != nil {
		log.Fatal("Ошибка. Соединение с БД не установлено. ", err)
	}

	err = db.Ping() // проверка доступности БД
	if err != nil {
		log.Fatal("Ошибка подключения к БД. ", err)
	}

	log.Println("БД подключена")
}
