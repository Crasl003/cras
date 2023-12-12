package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gohutool/boot4go-util/db"
	_ "github.com/mattn/go-sqlite3"
)

/**
* golang-sample源代码，版权归锦翰科技（深圳）有限公司所有。
* <p>
* 文件名称 : _db.go
* 文件路径 :
* 作者 : DavidLiu
× Email: david.liu@ginghan.com
*
* 创建日期 : 2022/5/12 21:23
* 修改历史 : 1. [2022/5/12 21:23] 创建文件 by LongYong
*/

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

type Algorithm struct {
	Name        string
	Description string
}

func InitDB() {

	_, err := dbPlus.GetDB().Exec(sql_table)
	if err != nil {
		panic(err)
	}

	InitAdminUser()
}

func AllDatabaseAlgorithms(c *gin.Context) {
	// 连接数据库
	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	// 执行全部算法数据
	rows, err := db.Query("SELECT algorithm FROM algorithms")
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	// 遍历查询结果
	for rows.Next() {
		var algorithm string
		err = rows.Scan(&algorithm)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println("Running algorithm:", algorithm)
		// 在此处执行具体的算法操作
	}

	fmt.Println("所有数据库算法数据已运行完毕")
}

func GetAlgorithmDetails(c *gin.Context) {
	id := c.PostForm("id")

	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	row := db.QueryRow("SELECT name, details FROM algorithms WHERE id = ?", id)
	var name, details string
	err = row.Scan(&name, &details)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Algorithm Name: %s\n", name)
	fmt.Printf("Algorithm Details: %s\n", details)
}
