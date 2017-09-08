package models

import (
	"DSP2/data/today"
	"DSP2/utils"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"encoding/json"
	//	"os"

	"github.com/golang/protobuf/proto"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type BidReq struct {
	BID_REQUEST *today.BidRequest
	GET_PLANS   []string        //根据条件获取相关的广告计划
	Gg          map[string]bool //查询到的当前的广告组

}

//根据当前的条件获取相关的
func (self *BidReq) GetPlan(bid *today.BidRequest) []byte {
	self.BID_REQUEST = bid
	//filename := fmt.Sprintf("bid_req_%d", os.Getpid())
	cont := fmt.Sprintf("DSP|bid_request|%s|%s|", bid.String(), time.Now().String())
	//	utils.Logdebug(filename, cont)
	utils.LogInfo(cont)
	//go self.SaveRquestDat(bid)

	if bid == nil {
		return []byte{}
	}

	selectMap := make(map[string]interface{})
	lk := *bid.GetDevice().Geo.City
	if lk != "" {
		selectMap["address"] = lk
	}
	ugender := bid.User.GetGender()
	if ugender.String() != "" {
		selectMap["gender"] = TODAY_USER_GENDER[bid.User.Gender.String()]
	}

	if *bid.Device.Os != "" {
		selectMap["app_type"] = *bid.Device.Os
	}
	yodstr := bid.User.GetYob()
	if yodstr != "" {
		selectMap["yob"] = yodstr
	}
	contype := bid.Device.ConnectionType.String()
	if contype != "" {
		selectMap["connection_type"] = TODAY_NT_ENUM[contype]
	}
	devicestr := bid.Device.GetCarrier()
	if devicestr != "" {
		selectMap["operator"] = devicestr
	}

	sql := "select * from tf_plan where "
	for k, v := range selectMap {
		ty := reflect.TypeOf(v)
		nn := ty.Name()
		if k != "app_type" {
			switch nn {
			case "string":
				sql += fmt.Sprintf("if(%s=-1,1,FIND_IN_SET(%q,%s)) and ", k, v, k)
			case "int":
				sql += fmt.Sprintf("if(%s=-1,1,FIND_IN_SET(%d,%s)) and ", k, v, k)
			}
		} else {
			sql += fmt.Sprintf("if(%s='',1,FIND_IN_SET(%q,app_type)) and ", k, v)
		}

	}

	jk := time.Now()
	sql += " start_time < " + strconv.FormatInt(jk.Unix(), 10) + " and "
	sql += "is_off = 0 and is_group_off = 0 "
	//fmt.Println(sql)
	dat, serr := Mysql_query(sql)

	if serr == false {
		return []byte{} //未能找到正确的对象数据
	}
	//匹配到的计划总数

	pnum := len(dat)

	//检查当前的查询是否完成
	planChan := make(chan bool, pnum) //
	defer close(planChan)

	for _, planv := range dat {
		go self.CheckPlan(planv, planChan) //检查当前的广告计划是否满足查询条件
	}

	//
	for i := 0; i < pnum; i++ {

		<-planChan //阻塞等待
	}

	//查询当前的广告创意
	jh := self.GetAds()
	//	fmt.Println("new==》》》", jh)
	return jh

}

//保存请求的数据

//检查满足条件的广告计划

func (self *BidReq) CheckPlan(plandat map[string]string, pchan chan bool) {
	//当前开启的判读条件的数量
	glen := 3
	//创建条件查询的额组
	showGroup := make(chan bool, glen) // 创建三个长度的缓冲区
	// 当前的函数结束的时候回收当前的
	defer close(showGroup)
	//1 检查当前的时间
	go self.OneDaytimeLimit(plandat, showGroup)
	//2 检查当前的广告计划的消费是否超过
	ptype, _ := strconv.Atoi(plandat["account_type"]) //投放的方式0 日预算 1 总预算
	go self.OneDayMoneyLimit(plandat, showGroup, ptype)

	go self.checkPlatform(plandat, showGroup)

	end_pv := true
	//阻塞等待
	for i := 0; i < glen; i++ {
		pv := <-showGroup
		if pv == false {
			end_pv = false
		}
	}

	if end_pv == true {

		if len(self.GET_PLANS) == 0 {
			self.GET_PLANS = []string{}
		}
		//查找满足条件的广告计划
		self.GET_PLANS = append(self.GET_PLANS, plandat["id"])
		//查询满足条件的广告创意
	}
	pchan <- true

}

//每个广告计划的每天的限额
func (self *BidReq) OneDaytimeLimit(dat map[string]string, bbq chan bool) {

	//当前投放时间
	if dat["tf_time"] == "0" && dat["tf_style"] == "0" {
		//完全的满足不用考虑时间
		bbq <- true
		return
	}

	//投放时间限制
	if dat["tf_time"] == "0" && dat["tf_style"] == "1" {
		//判断办小时的时间
		JK := Stlimit(dat["tf_range"])
		if JK == false {
			//时间检测未能通过
			bbq <- false
			return
		}
	}
	nowt := time.Now()
	timestamp := time.Date(nowt.Year(), nowt.Month(), nowt.Day(), 0, 0, 0, 0, time.Local) //当日的时间戳

	st_unix := strconv.FormatInt(timestamp.Unix(), 10)
	start_unix_time, _ := strconv.Atoi(st_unix)
	//时间对象的控制S
	if dat["tf_time"] == "1" && dat["tf_style"] == "0" {
		end_time, _ := strconv.Atoi(dat["end_time"])
		st_time, _ := strconv.Atoi(dat["start_time"])

		if start_unix_time > end_time || start_unix_time < st_time {
			//出错了少年 不在投放的时间段呢
			bbq <- false
			return
		}
		//判断开始时间，结束时间
	}

	//时间段控制的处理
	if dat["tf_time"] == "1" && dat["tf_style"] == "1" {
		//时间段的判断办小时
		end_time, _ := strconv.Atoi(dat["end_time"])
		st_time, _ := strconv.Atoi(dat["start_time"])
		if start_unix_time > end_time || start_unix_time < st_time {
			//出错了少年 不在投放的时间段呢
			bbq <- false
			return
		}
		JK := Stlimit(dat["tf_range"])
		//时间点的判断
		if JK == false {
			//时间检测未能通过
			bbq <- false
			return
		}
	} else {
		bbq <- true
	}
}

//当前广告计划的消费限额
//获取当前计划的CPM的付费的方式的的计费总额 tp 0 表示按天的限额 1总的限额 计划的限额的
func (self *BidReq) OneDayMoneyLimit(dat map[string]string, bbq chan bool, tp int) {

	nowt := time.Now()
	timestamp := time.Date(nowt.Year(), nowt.Month(), nowt.Day(), 0, 0, 0, 0, time.Local) //当日的时间戳
	start_unix_time := timestamp.Unix()

	planid := dat["id"]
	var mongotable string = "UseMoney"
	defer func() {
		if err := recover(); err != nil {
			//fmt.Println("there had some error 4", err)
		}
	}()

	mongo := NewDataStore()

	job := &mgo.MapReduce{
		Map:    "function() { emit(this.adid,this.bidprce) }",
		Reduce: "function(key, values) { return Array.sum(values) }",
	}
	var result []struct {
		Id    int `bson:"adid"` //广告的ID
		Value int
	}
	sttime := strconv.Itoa(int(start_unix_time)) + planid
	var useMoney struct {
		Money int `bson:"money"`
	}

	if dat["account_type"] == "0" {
		//_, err := mongo.DB(MongodbConf.DataBase).C(mongotable).Find(bson.M{"planid": planid, "timestamp": bson.M{"$gte": strconv.FormatInt(start_unix_time, 10)}}).MapReduce(job, &result)
		err := mongo.C(mongotable).Find(bson.M{"costid": sttime}).One(&useMoney)
		if err != nil {
			//fmt.Println("There had some error 1", err, sttime)
		}
	} else {
		_, err := mongo.C(mongotable).Find(bson.M{"planid": planid}).MapReduce(job, &result)
		if err != nil {
			//fmt.Println("There had some error 2")
		}
	}
	defer mongo.Close()
	if useMoney.Money == 0 {
		bbq <- true
		return
	} else {
		price := dat["account_price"]
		limitMoney := SftoI(price) * 1000
		if useMoney.Money > limitMoney {
			bbq <- false
			return
		} else {
			bbq <- true
			return
		}
	}

	select {
	case <-time.After(time.Microsecond * 20):
		bbq <- false
	}
	// 1000 * 100 获取当前的足额，转化成元为单位的
}

//检查当前的平台
func (self *BidReq) checkPlatform(dat map[string]string, bbq chan bool) {

	//获取当前的投放的额平台信息
	types := self.BID_REQUEST.Device.GetOs()
	//tparr := strings.Split(dat["platform"], ",")
	if dat["type"] == "2" {
		if dat["platform"] == "-1" {
			//当前的投放的平台不限制
			bbq <- true
			return
		}

		//判断当前的平台
		if strings.Contains(dat["platform"], types) {
			bbq <- true
		} else {
			bbq <- false
		}
		return
	}
	fmt.Println("检查当前哦平台")
	bbq <- true

}

//当前广告组的查询

func (self *BidReq) CheckGroup(dat map[string]string, bbq chan bool) {
	//检查当前的广告加护
	group_id := dat["group_id"]
	if self.Gg[group_id] != false {
		bbq <- self.Gg[dat["group_id"]] //  直接返回当前的是否开启当前的广告组的判读
	}

	sql := fmt.Sprintf("select * from tf_ads_group where id = %s", group_id) //查询当前广告组的信息

	_, gerr := Mysql_query(sql)

	if gerr == false {
		//未能查找到相关的组数据返回错误
		bbq <- false
		return
	}

	bbq <- false //获取当前的组的限额
	select {
	case <-time.After(time.Microsecond):
		//当前超时返回false
		bbq <- false
	}

}

//查询当前满足条件的广告创意

func (self *BidReq) GetAds() []byte {
	bid := self.BID_REQUEST
	if len(self.GET_PLANS) == 0 {
		return []byte{}
	}
	str := strings.Join(self.GET_PLANS, ",")
	sql := fmt.Sprintf("SELECT ads.id,ads.type,ads.ad_type,im.*,ads.source,ads.detail_url,ads.plan_id,plan.link_url,plan.group_id from tf_plan as plan inner join tf_ads as ads on plan.id =ads.plan_id INNER JOIN tf_images as im ON ads.id = im.ads_id WHERE ads.plan_id in (%s) AND ads.is_off =0 and plan.is_group_off = 0  AND im.status = %q", str, "normal")

	dat, isdat := Mysql_query(sql) //广告创意集合

	if isdat == false {
		return []byte{}
		//当前的时间限制的
		fmt.Println("mysql not found some message 123")
	}

	//满足条件的创意的合
	var Ads_Array []map[string]string

	adsNum := bid.GetAdslots()
	//广告位置的合计
	//var Arr_Seats []*SeatBid //广告计划
	//var plan_ads_arr []map[string]string = make([]map[string]string, len(dat))

	//用户保存多个，广告位的相求的请求信息

	var SomeSetids map[string][]map[string]string = make(map[string][]map[string]string)
	//广告请求位置
	for _, ads := range adsNum {
		//广告位的adtype
		bidadress := ads.GetId()        //当前广告位的，广告的信息id信息
		for _, vm := range ads.AdType { //当前广告位可接受的，广告类型的遍历
			kv := vm.String()
			knum := today.AdType_value[kv] //转化成对应的数值表示的广告位
			for _, vv := range dat {       //广告创意
				//广告创意
				ch_type := vv["ad_type"]                   //创意 投放的位置
				ad_type_arr := strings.Split(ch_type, ",") //当前广告位支持的位置
				for _, tv := range ad_type_arr {
					tv_num, _ := strconv.Atoi(tv)
					if tv_num == int(knum) { //满足条件的创意
						vv["ads_type"] = tv
						Ads_Array = append(Ads_Array, vv)
						SomeSetids[bidadress] = append(SomeSetids[bidadress], vv)
					}
				}
			}

		}
	}

	return self.CreatAds(SomeSetids) // 生成Probuf的格式
	//安装广告位分开的投放的位置
}

//创建满足条件的广告创意
//生成probuf格式数据
func (self *BidReq) CreatAds(dat map[string][]map[string]string) []byte {

	if len(dat) == 0 {
		mk := []byte{}
		return mk
	}

	var domain string = "http://dspapi.qcwan.com"
	var param string = "notify?user_id={user_id}&request_id={request_id}&adid={adid}&bid_price={bid_price}&ip={ip}&timestamp={timestamp}&did={did}&imei={IMEI}&idfa={IDFA}&os={OS}&g_pos={g_pos}"
	var imgdom string = ""

	adslos := self.BID_REQUEST.Adslots

	var Arr_Seats []*today.SeatBid

	for _, adslocation := range adslos { //广告位每个广告位对象一个set读
		k := adslocation.GetId()
		setAds := dat[k]
		//生成一个新广告
		Seatbids := &today.SeatBid{}
		var Ads_arr []*today.Bid //
		if len(dat) == 0 {
			return []byte{}
		}
		//for key, setAds := range dat { //广告创意数组

		for _, ads := range setAds {
			//

			ads_type := ads["ad_type"]
			ad_type_arr := strings.Split(ads_type, ",")
			//查询当前的计划id
			if len(ads) == 0 {

				continue
			}
			plansql := fmt.Sprintf("select * from tf_plan where id = %s", ads["plan_id"])

			datplam, isdat := Mysql_query(plansql) //满足条件的创意

			if isdat == false {
				//当前的时间限制的
				continue
			}

			//广告创意
			for _, v := range ad_type_arr {

				//创意生成-------------------------------------------------------------
				biddata := &today.Bid{}
				ida_t := time.Now().Unix()
				d_yime := strconv.FormatInt(ida_t, 10)
				biddata.Id = &d_yime //生成唯一的商品的信息
				//创意的ID
				cid, _ := strconv.ParseUint(d_yime, 10, 64)
				//获取当前金额
				tf_money := SftoI(datplam[0]["first_price"])

				biddata.Adid = &cid //参与竞价的广告ID
				umprice := uint32(tf_money)

				//pp := adslocation.GetBidFloor() + 1
				//biddata.Price = &pp
				biddata.Price = &umprice
				//biddata.Price = adslocation.BidFloor + 1
				kk := k
				biddata.AdslotId = &kk //当前广告位对象的ID
				cidstring := strconv.FormatUint(cid, 10)
				biddata.Cid = &cidstring //扩展的广告ID

				modelste := &today.MaterialMeta{} //广告素材对象
				modelste.AdType = getAdtype(v)
				apparam := fmt.Sprintf("%s&plan_id=%s&ads=%s&group_id=%s", param, ads["plan_id"], ads["id"], ads["group_id"])
				nurlstr := fmt.Sprintf("%s/%s/%s", domain, "win", apparam)
				modelste.Nurl = &nurlstr
				titstr := ads["title"]
				modelste.Title = &titstr
				sourcestr := ads["source"]
				modelste.Source = &sourcestr

				//获取传递过来的banner

				//banner := adslots.Banner
				//banner_dat := banner[stk_v]

				//banner的图片信息显示
				imgbanner := &today.MaterialMeta_ImageMeta{}
				width_32, _ := strconv.ParseUint(ads["width"], 10, 32)
				height_32, _ := strconv.ParseUint(ads["height"], 10, 32)
				wid_addr := uint32(width_32)
				height_addr := uint32(height_32)
				imgbanner.Width = &wid_addr
				imgbanner.Height = &height_addr
				ads_title := ads["title"]

				imgbanner.Description = &ads_title

				if ads["type"] != "3" {
					im_url := fmt.Sprintf("%s%s", imgdom, ads["img_url"])
					imgbanner.Url = &im_url
					imgbanner.Urls = []string{fmt.Sprintf("%s%s", imgdom, ads["img_url"])}
				} else {
					var imgsA []string
					dsd := strings.Split(ads["img_url"], ",")
					for _, v := range dsd {
						imgsA = append(imgsA, fmt.Sprintf("%s%s", imgdom, v))
					}
					imgurl := imgsA[0]
					imgbanner.Url = &imgurl
					imgbanner.Urls = imgsA

				}
				modelste.ImageBanner = imgbanner
				//设置当前的操作流程

				dsp_external := &today.MaterialMeta_ExternalMeta{}
				detailurl := dat[k][0]["detail_url"]
				//exter_url := ads["link_url"]
				dsp_external.Url = &detailurl
				modelste.External = dsp_external
				biddata.Creative = modelste

				modelste.ShowUrl = []string{fmt.Sprintf("%s/%s/%s", domain, "show", apparam)}
				if len(ads["show_url"]) > 0 {
					modelste.ShowUrl = append(modelste.ShowUrl, ads["show_url"])
				}
				modelste.ClickUrl = []string{fmt.Sprintf("%s/%s/%s", domain, "click", apparam)}
				if len(ads["click_url"]) > 0 {
					modelste.ClickUrl = append(modelste.ClickUrl, ads["click_url"])
				}
				//视频
				if ads["type"] == "4" {
					vurl := new(string)
					*vurl = ads["img_url"]
					modelste.VideoUrl = vurl
					//视频监测地址
				}

				bjk := []string{"3", "4", "7", "9", "18", "4"}
				//获取当前系统
				for _, vtype := range bjk {
					if v == vtype {
						os := self.BID_REQUEST.Device.Os
						switch *os {
						case "ios":
							iosDat := &today.MaterialMeta_IosApp{}
							ios_downurl := datplam[0]["link_url"]
							iosDat.DownloadUrl = &ios_downurl
							ios_appname := datplam[0]["app_name"]
							iosDat.AppName = &ios_appname
							modelste.IosApp = iosDat
						case "android":
							android_url := datplam[0]["link_url"]
							android := &today.MaterialMeta_AndroidApp{}
							android_appname := datplam[0]["app_name"]
							android.AppName = &android_appname
							android.DownloadUrl = &android_url
							android.WebUrl = &detailurl
							modelste.AndroidApp = android
						}
					}
				}
				//创意生成结束------------------------------------------------------------
				Ads_arr = append(Ads_arr, biddata) //创意数组

			}
		}

		//广告位重置
		Seatbids.Ads = Ads_arr
		//生成新的广告位
		Arr_Seats = append(Arr_Seats, Seatbids)
	}

	//}

	res := &today.BidResponse{}
	res.Seatbids = Arr_Seats

	res.RequestId = self.BID_REQUEST.RequestId
	data, err := proto.Marshal(res)

	if err != nil {
		return []byte{}
	}
	return data

	//return []byte{}
}

//保存当前的数据对象

func SaveDataM(tab string, dat map[string]string) {
	keys := []string{}
	values := []string{}
	for k, v := range dat {
		keys = append(keys, k)
		values = append(values, v)
	}
	sql := fmt.Sprintf("insert into %s (%s) values(%s)", tab, strings.Join(keys, ","), strings.Join(values, ","))

	Mysql_insert(sql)
}

//当前的分钟的级别的投放的控制
func Stlimit(selecTime string) bool {
	var jj [][]string
	json.Unmarshal([]byte(selecTime), &jj)
	nowt := time.Now()
	dl := nowt.Weekday().String()
	daytime := TODAY_WEEK_NUM[dl]
	htime := nowt.Hour()         //获取当前时
	mintime := nowt.Minute()     //获取当前的分钟
	skip := mintime / 30         //当前的时偏移量
	hanftindex := htime*2 + skip //半小时索引
	//fmt.Println("是否处于投放时间", jj[daytime][hanftindex]) //获取当前是否投放广告 tf_style 投放的方式全天
	//当前的如果不满组条件的时候创建一个新的文件
	//	tp, _ := strconv.Atoi( jj[daytime][hanftindex]))
	//检查当前的投放是不是全天的投放的模式
	if jj[daytime][hanftindex] == "0" {
		return false
	} else {
		return true
	}
}

func getAdtype(num string) *today.AdType {

	//	return modelste
	n, _ := strconv.ParseUint(num, 10, 32)
	var th today.AdType = today.AdType(n)
	return &th
}
