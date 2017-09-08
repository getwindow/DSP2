// models project models.go
package models

import (
	"DSP2/data/today"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Today struct {
	BID_REQUEST *today.BidRequest
	GET_PLANS   []map[string]interface{} //根据条件获取相关的广告计划
}

var (
	TODAY_USER_GENDER map[string]int = map[string]int{"UNKNOWN": 3, "FEMALE": 1, "MALE": 2}
	TODAY_NT_ENUM     map[string]int = map[string]int{"Honeycomb": 1, "WIFI": 2, "UNKNOWN": 3, "NT_2G": 4, "NT_4G": 5} //网络编辑
	TODAY_WEEK_NUM    map[string]int = map[string]int{"Monday": 0, "Tuesday": 1, "Wednesday": 2, "Thursday": 3, "Friday": 4, "Saturday": 5, "Sunday": 6}
)

const (
	TF_STYLE_ALL  = 0
	TF_STYLE_TIME = 1
)

//初始化，初始化数据库
func init() {
	initDatabase() //初始化配置信息
	//InitMongo()    //初始化mongo
}

func CheckError(err error) {
	if err != nil {
		fmt.Println("当前出错", err)
	}
}

//mysql config

var MysqlConf struct {
	Host     string
	User     string
	PassWord string
	Port     string
	DataBase string
}

//redis config
var RedisConf struct {
	Host     string
	User     string
	PassWord string
	Port     string
}

//mongodb config
var MongodbConf struct {
	Host     string
	User     string
	PassWord string
	Port     string
	DataBase string
}

/*
获取程序运行路径
*/
func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println("This is an error")
	}
	return strings.Replace(dir, "\\", "/", -1)
}

//初始化mysql
func initDatabase() {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("数据库初始化", err)
			return
		}
	}()
	path := getCurrentDirectory()

	//获取当前文件的配置
	dat, err := ioutil.ReadFile(path + "/configs/config.json")
	//close the file
	if err != nil {
		fmt.Println("为找到配置文件")
		return
	}

	dt := DataChange(string(dat))

	//mysql
	dta := dt["mysql"]
	dap := DataChange(JsonEncodeString(dta))
	MysqlConf.Host = dap["host"].(string)
	MysqlConf.PassWord = dap["password"].(string)
	MysqlConf.Port = dap["port"].(string)
	MysqlConf.User = dap["user"].(string)
	MysqlConf.DataBase = dap["database"].(string)

	//redis
	redisdt := dt["redis"]
	redisdap := DataChange(JsonEncodeString(redisdt))
	RedisConf.Host = redisdap["host"].(string)
	RedisConf.PassWord = redisdap["password"].(string)
	RedisConf.Port = redisdap["port"].(string)
	RedisConf.User = redisdap["user"].(string)

	//mongodb
	mongodt := dt["mongodb"]
	mongodap := DataChange(JsonEncodeString(mongodt))
	MongodbConf.Host = mongodap["host"].(string)
	MongodbConf.PassWord = mongodap["password"].(string)
	MongodbConf.Port = mongodap["port"].(string)
	MongodbConf.User = mongodap["user"].(string)
	MongodbConf.DataBase = mongodap["database"].(string)
	fmt.Println(MysqlConf)
	fmt.Println(MongodbConf)

}

//数据格式转化的操作

func DataChange(data string) map[string]interface{} {
	var dat map[string]interface{}
	json.Unmarshal([]byte(data), &dat)
	return dat
}

// 结构转换成json对象
func JsonEncodeString(data interface{}) string {
	back, err := json.Marshal(data)
	if err != nil {
		return "encode error"
	}
	return string(back)
}

//map的类型转换成！

func MaptoJson(data map[string]interface{}) string {
	configJSON, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return ""
	}
	return string(configJSON) //返回格式化后的字符串的内容0
}

//字符串格式的浮点转整数

func SftoI(tf_money string) int {
	ttf_money, _ := strconv.ParseFloat(tf_money, 10)
	tend_m := ttf_money * 100
	tff_money := strconv.FormatFloat(tend_m, 'f', 0, 64)
	tlimi, _ := strconv.Atoi(tff_money)
	return tlimi
}

//发布预警到dsp
func SendMsgToDsp(content string) {
	contents := fmt.Sprintf(`{"msgtype": "text", "text": {"content": %s}}`, content)
	fmt.Println("sendcontent", contents)
	mk := make(map[string]interface{})
	mk["msgtype"] = "text"
	mk["text"] = &map[string]string{"content": content}
	dd := MaptoJson(mk)

	go Httprequest("https://oapi.dingtalk.com/robot/send?access_token=90d9fd06ec4bc54f32ecb3a609647a4530f73606b02fc229e7bbd794679fa71f", "POST", dd)
}

func Httprequest(requestUrl, requestType, requestData string) {
	defer func() {
		if err := recover(); err != nil {

		}
	}()
	// 设置当前的超时的时间
	client := new(http.Client)
	reqest, err := http.NewRequest(requestType, requestUrl, strings.NewReader(requestData))
	if err != nil {

	}

	reqest.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	reqest.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	reqest.Header.Add("Accept-Encoding", "gzip, deflate")
	reqest.Header.Add("Accept-Language", "zh-cn,zh;q=0.8,en-us;q=0.5,en;q=0.3")
	reqest.Header.Add("Connection", "keep-alive")
	reqest.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:12.0) Gecko/20100101 Firefox/12.0")

	resp, err := client.Do(reqest)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {

	}
	bodyText := string(body)
	status := resp.StatusCode
	backContent := fmt.Sprintf("请求状态：%d 请求的响应时间: %s 请求响应的页面内容：%s", status, resp.Header.Get("Date"), bodyText)

	fmt.Println(backContent)

}
