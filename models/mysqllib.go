package models

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var Mysql_db *sql.DB

func mysql_error() {
	if err := recover(); err != nil {
		fmt.Println("数据库连接失败", err)
	}
}

//连接当前的数据库
func Mysql_connect() {
	defer mysql_error()
	var err error
	cont := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8", MysqlConf.User, MysqlConf.PassWord, MysqlConf.Host, MysqlConf.Port, MysqlConf.DataBase)
	Mysql_db, err = sql.Open("mysql", cont)
	Mysql_db.SetMaxOpenConns(1000)
	Mysql_db.SetMaxIdleConns(1000)
	Mysql_db.SetConnMaxLifetime(60 * time.Second)
	CheckError(err)
}

//关闭当前的mysql连接
func Mysql_colose() {
	err := Mysql_db.Close()
	CheckError(err)
}

//检查当前连接
func Check_connect() {
	defer mysql_error()
	if Mysql_db == nil {
		Mysql_connect()
		return
	}
	err := Mysql_db.Ping()

	if err != nil {
		fmt.Println(err, "ddd===")
		//当前的数据库连接断开重新连接
		//数据库断开许重连
		Mysql_connect()
	}
}

//写书数据
func Mysql_insert(sqla string) {
	Check_connect()
	_, err := Mysql_db.Exec(sqla)
	CheckError(err)

}

//插叙当前的数据返回数据
func Mysql_query(sql_q string, args ...interface{}) ([]map[string]string, bool) {
	defer mysql_error()
	Check_connect()
	rows, err := Mysql_db.Query(sql_q, args...)
	if err != nil {
		return []map[string]string{}, false
	}
	columns, err := rows.Columns()
	if err != nil {
		panic(err.Error())
	}

	values := make([]sql.RawBytes, len(columns))

	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	var dat []map[string]string = []map[string]string{}
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error())
		}
		mpp := make(map[string]string)
		var value string
		for i, col := range values {
			// Here we can check if the value is nil (NULL value)
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			mpp[columns[i]] = value
		}
		dat = append(dat, mpp)
	}
	defer rows.Close() //关闭当前的查询操作

	if err = rows.Err(); err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	return dat, true //返回当前查询的结果 ，当前的查询的状态
}
