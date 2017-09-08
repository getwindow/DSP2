package models

import (
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var MongoDb *mgo.Database
var Hsession *mgo.Session

type Adt struct {
	Gameid   string
	App_type string
	Channel  int
	Imei     string
}

func GetMongoSession() *mgo.Session {
	var con_str string
	if MongodbConf.User == "" && MongodbConf.PassWord == "" {
		con_str = fmt.Sprintf("%s:%s", MongodbConf.Host, MongodbConf.Port)
	} else {
		con_str = fmt.Sprintf("%s:%s@%s:%s", MongodbConf.User, MongodbConf.PassWord, MongodbConf.Host, MongodbConf.Port)
	}
	session, err := mgo.Dial(con_str)
	//mgo.DialWithTimeout(con_str, time.Second*60)
	session.SetMode(mgo.Monotonic, true)
	//fmt.Println("数据库连接状态===>>><<><>", mgo.GetStats())
	if err != nil {
		fmt.Println("mongo connect err!!")
	}

	return session
}

//初始化当前的额mongo

func InitMongo() {
	var con_str string
	var err error
	if MongodbConf.User == "" && MongodbConf.PassWord == "" {
		con_str = fmt.Sprintf("%s:%s", MongodbConf.Host, MongodbConf.Port)
	} else {
		con_str = fmt.Sprintf("%s:%s@%s:%s", MongodbConf.User, MongodbConf.PassWord, MongodbConf.Host, MongodbConf.Port)
	}

	Hsession, err = mgo.Dial(con_str)
	//mgo.DialWithTimeout(con_str, time.Second*60)
	Hsession.SetMode(mgo.Monotonic, true)
	//fmt.Println("数据库连接状态===>>><<><>", mgo.GetStats())
	if err != nil {
		fmt.Println("mongo connect err!!")
	}
}

func MongoTest() {

	mon := NewDataStore()

	var useMoney struct {
		Money int `bson:"money"`
	}
	defer mon.Close()

	if mon != nil {
		err := mon.C("UseMoney").Find(bson.M{"costid": "149849280014"}).One(&useMoney)
		if err != nil {
		}

	} else {
		fmt.Println("Mongo connect error!")
	}
}

type (
	DataStore struct {
		session *mgo.Session
	}
)

//关闭当前
func (d *DataStore) Close() {
	d.session.Close()
}

//Returns a collection from the database.
func (d *DataStore) C(name string) *mgo.Collection {
	return d.session.DB("dsp").C(name)
}

//Create a new DataStore object for each HTTP request
func NewDataStore() *DataStore {
	//检查当前的额mongo连接情况
	if Hsession == nil {
		InitMongo()
	}
	err := Hsession.Ping()
	if err != nil {
		InitMongo()
	}
	ds := &DataStore{
		session: Hsession.Copy(),
	}
	return ds
}
