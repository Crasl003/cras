package db

import (
	"database/sql"

	"github.com/gohutool/boot4go-util/db"
	_ "github.com/mattn/go-sqlite3"
)

var dbPlus db.DBPlus

func init() {
	db1, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		panic(err)
	}
	dbPlus = db.DBPlus{DB: db1}
}

var sql_table = `CREATE TABLE if not exists "t_user" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "userid" VARCHAR(64) NULL,
    "username" VARCHAR(64),
	"password" VARCHAR(128),
    "createtime" TIMESTAMP default (datetime('now', 'localtime'))
);

CREATE TABLE if not exists "t_db" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "dbid" VARCHAR(64) NULL,
    "endpoint" VARCHAR(64),
	"username" VARCHAR(64),
	"password" VARCHAR(64),
    "createtime" TIMESTAMP default (datetime('now', 'localtime'))
);

CREATE TABLE if not exists "t_repos" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "reposid" VARCHAR(64) NULL,
    "name" VARCHAR(255),
    "description" text,
    "endpoint" VARCHAR(255),
	"username" VARCHAR(255),
	"password" VARCHAR(255),
    "createtime" TIMESTAMP default (datetime('now', 'localtime'))
);

CREATE TABLE if not exists "t_uploads" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "Algorithm_Name" VARCHAR(64) ,
    "Downloads" VARCHAR(64),
	"Algorithm_Version" VARCHAR(64),
    "Author_Name" VARCHAR(64),
    "Introduction" VARCHAR(64),
    "Function" VARCHAR(64),
    "Space" VARCHAR(64),
    "createtime" TIMESTAMP default (datetime('now', 'localtime'))
);

CREATE TABLE if not exists "t_orchestrator" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "orchestratorid" VARCHAR(64) NULL,
    "name" VARCHAR(64),
    "description" text,
    "json" text,
    "userid" VARCHAR(64),
    "createtime" TIMESTAMP default (datetime('now', 'localtime'))
);
`

func InitDB() {

	_, err := dbPlus.GetDB().Exec(sql_table)
	if err != nil {
		panic(err)
	}

	InitAdminUser()
}
