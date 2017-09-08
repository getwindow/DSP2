package controllers

import (
	"DSP2/data/today"
	"DSP2/models"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"DSP2/utils"
	"io"
	"os"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2/bson"
)

const (
	TJ_FEED_DSP_ID  string = "1756165498"
	TJ_FEED_DSP_KEY string = "a74e696576394976bc694fbd58a2b0d6"
	XQ_FEED_DSP_ID  string = "1756165504"
	XQ_FEED_DSP_KEY string = "c8c2df73229b47378f8eceb9cc12ea1a"
	DZ_FEED_DSP_ID  string = "1756165502"
	DZ_FEED_DSP_KEY string = "f49916e2447f490d93822dff5c345aa3"
	WIN_NOTIFY      string = "1"
	SHOW_NOTIFY     string = "2"
	CLICK_NOTIFY    string = "3"
)

type NotifyDat struct {
	UserId     string
	RequestID  string
	Adid       string
	BidPrce    uint64
	Ip         string
	TimesTamp  string
	Did        string
	PlanId     string
	Ads        string
	Os         string
	Win_num    int
	GroupID    string
	NotifyType string
}

//宏替换的数据

func TodayBidRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	st_time := time.Now().UnixNano()
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("出错了")
			fmt.Println(err)
		}

	}()

	httputil.DumpRequest(r, true)
	bydata, _ := ioutil.ReadAll(r.Body)
	reqa := &today.BidRequest{}

	era := proto.Unmarshal(bydata, reqa)
	if era != nil {
		fmt.Println("转化错误iiiii")
		panic("解析错误")
	}

	//todaymodel := models.Today{}
	//data := todaymodel.GetPlans(reqa) //回去相关的广告计划
	ntt := models.BidReq{}
	data := ntt.GetPlan(reqa)

	end_time := time.Now().UnixNano()
	fmt.Println("\r\n useTime:", (end_time-st_time)/1e6)
	if len(data) != 0 {
		w.Write(data)
		//filename := fmt.Sprintf("bid_responded_%d", os.Getpid()) //获取当前的额进程号
		cont := fmt.Sprintf("DSP|bid_responded|%s|%s|", string(data), time.Now().String())
		//		utils.Logdebug(filename, cont)
		utils.LogInfo(cont)
		//go saveReponseDat((end_time-st_time)/1e6, data)
	}

}

//监测地址数据
var JCdat []map[string]string = make([]map[string]string, 1)

//广告计划的消费汇总

type CostMoney struct {
	Costid  string
	Groupid string
	Planid  string
	Ads     string
	Money   uint64
	AddTime int64
}

func saveData(ntype string, ps url.Values) {

	dat := NotifyDat{}
	dat.Adid = ps.Get("adid")
	dat.BidPrce = Decprice(ps.Get("bid_price"))
	dat.Did = ps.Get("did")
	dat.Ip = ps.Get("ip")
	dat.RequestID = ps.Get("request_id")

	dat.TimesTamp = ps.Get("timestamp")
	dat.UserId = ps.Get("user_id")
	dat.PlanId = ps.Get("plan_id")
	dat.Ads = ps.Get("ads")
	dat.Os = ps.Get("os")
	dat.GroupID = ps.Get("group_id")
	win_num, _ := strconv.Atoi(ps.Get("g_pos"))
	dat.Win_num = win_num
	dat.NotifyType = ntype
	conts := fmt.Sprintf("DSP|%s|%s|%s|%s|%s|%s|%s", "notify_type"+ntype, dat.Adid, dat.Did, dat.RequestID, dat.TimesTamp, dat.PlanId, dat.GroupID)

	//	mongo := models.GetMongoSession().Copy()
	//	defer mongo.Close()
	//	mongo.DB(models.MongodbConf.DataBase).C(collectionName).Insert(dat)
	mgodb := models.NewDataStore()
	mgodb.C("notifyDat").Insert(dat)
	defer mgodb.Close()

	//utils.Logdebug("notify_type"+ntype, conts)
	utils.LogInfo(conts)

	if ntype == WIN_NOTIFY {
		//汇总当前金额
		nowt := time.Now()
		timestamp := time.Date(nowt.Year(), nowt.Month(), nowt.Day(), 0, 0, 0, 0, time.Local) //当日的时间戳

		key := strconv.FormatInt(timestamp.Unix(), 10) + dat.PlanId

		//格式化当前的文档的
		mgodb := models.NewDataStore()

		mtableName := "UseMoney"
		num, _ := mgodb.C(mtableName).Find(bson.M{"costid": key}).Count()
		if num == 0 {
			money := &CostMoney{}
			money.Costid = key
			money.Ads = dat.Ads
			money.Groupid = dat.GroupID
			money.Money = dat.BidPrce
			money.Planid = dat.PlanId
			money.AddTime = time.Now().Unix()
			mgodb.C(mtableName).Insert(money)
		} else {
			//更新限额的
			//Update(bson.M{"_id": bson.ObjectIdHex("5204af979955496907000001")}, bson.M{"$inc": bson.M{ "age": -1, }})
			mgodb.C(mtableName).Update(bson.M{"costid": key}, bson.M{"$inc": bson.M{"money": dat.BidPrce}})
		}
		defer mgodb.Close()
	}
}

//获胜后获取的数据github.com/djimenez/iconv-go"

func WinRequest(w http.ResponseWriter, r *http.Request, psa httprouter.Params) {
	fmt.Println("WIN")
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("There had some ER win")
			fmt.Println(err)
		}
	}()
	ps := r.URL.Query()
	saveData(WIN_NOTIFY, ps)
}

//展示的时候请求的数据

func ShowRequest(w http.ResponseWriter, r *http.Request, psa httprouter.Params) {
	fmt.Println("Show ")
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("There had some ER show")
			fmt.Println(err)
		}
	}()
	ps := r.URL.Query()
	//fmt.Println(ps)
	//go SavetoMysql(SHOW_NOTIFY, ps)
	saveData(SHOW_NOTIFY, ps)
}

//点击的时候展示

func ClickRequest(w http.ResponseWriter, r *http.Request, psa httprouter.Params) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("There had some ER click")
			fmt.Println(err)
		}
	}()
	fmt.Println("click")
	ps := r.URL.Query()
	saveData(CLICK_NOTIFY, ps)
	//	clickurl := JCdat[0]["click_url"]
	//Hreaplace(showurl)
	//go senddata(psa, clickurl)
	//go func(ps httprouter.Params) {}(ps)
}

//视频监测地址

func PlayVideoRequest(w http.ResponseWriter, r *http.Request, psa httprouter.Params) {
	fmt.Println("video play")
}

//视频开始播放监测地址
func StartPlayVideo(w http.ResponseWriter, r *http.Request, psa httprouter.Params) {
	fmt.Println("视频 开始播放")
}

//视频播放结束
func EndPlayVideo(w http.ResponseWriter, r *http.Request, psa httprouter.Params) {

}

//落地页打开监测地址
func Lpopen(w http.ResponseWriter, r *http.Request, psa httprouter.Params) {
	fmt.Println("落地页打开")
}

func senddata(ps httprouter.Params, sendurl string) {
	if len(sendurl) == 0 {
		return
	}
	hdat := map[string]string{}
	reg, _ := regexp.Compile(`__\w+__`)
	hdat["__IDFA__"] = ps.ByName("idfa")
	hdat["__IMEI__"] = ps.ByName("imei")
	hdat["__IP__"] = ps.ByName("ip")
	hdat["__TS__"] = ps.ByName("ts")
	hdat["__OS__"] = ps.ByName("os")
	sendstr := reg.ReplaceAllStringFunc(sendurl,
		func(b string) string {
			dh := hdat[b]
			return dh
		})
	Httprequest(sendstr, "GET", "")
}

func Index(w http.ResponseWriter, r *http.Request, psa httprouter.Params) {
	//	fmt.Println("Index header")
	//	for i := 0; i < 1; i++ {
	//		Httprequest("http://127.0.0.1:9090/show/notify?user_id=28370540189&request_id=20170626163122010008059012426E13&adid=1498465882&bid_price=WVDnZwABLexZUOdnAAEt7N7oaCPv_g0c2Rhx-g&ip=117.136.44.188&timestamp=1498474308&did=&imei=869580027816462&idfa=&os=0&g_pos=0&plan_id=14&ads=1018&group_id=10", "GET", "")
	//	}
	w.Write([]byte("欢迎到来到ADX")) //
}

//保存响应的数据

func saveReponseDat(utime int64, data []byte) {
	newTest := &today.BidResponse{}
	err := proto.Unmarshal(data, newTest)
	if err != nil {
		fmt.Println("----->>>")
		return
	}
	btext := newTest.String()
	if len(btext) == 0 {
		return
	}
	dt := map[string]interface{}{"useTime": utime, "content": btext, "addtime": time.Now().Unix()}

	mongo := models.NewDataStore()
	mongo.C("backdat").Insert(dt)
	mongo.Close()
	//	mongo := models.GetMongoSession().Copy()
	//	mongo.DB(models.MongodbConf.DataBase).C("backdat").Insert(dt)
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

//写入当前的测试

func writeFile(dat []byte) {

	path := getCurrentDirectory()
	fl, err := os.OpenFile(path+"/bit.req", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0777)
	n, err := fl.Write(dat)
	fmt.Println(dat)
	if err != nil {
		fmt.Println("打开文件失败")
		return
	}
	defer fl.Close()

	if err == nil && n < len(dat) {
		err = io.ErrShortWrite
	}
}

//saveDatTo mysql
