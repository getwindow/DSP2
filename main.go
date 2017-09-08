// DSP2 project main.go
package main

import (
	"DSP2/controllers"

	"DSP2/utils"
	"flag"

	"fmt"
	"net/http"
	"runtime"

	"github.com/julienschmidt/httprouter"
)

var port string

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	go utils.LogSave()
	flag.StringVar(&port, "port", ":9090", "port to listen") //监听当前的参数的设置
	flag.Parse()

	controllers.IM() //测试

	router := httprouter.New()
	router.GET("/index", controllers.Index)
	router.POST("/bit/req", controllers.TodayBidRequest)
	router.GET("/win/notify", controllers.WinRequest)
	router.GET("/click/notify", controllers.ClickRequest)
	router.GET("/show/notify", controllers.ShowRequest)

	http.ListenAndServe(port, router)
	fmt.Println("run end")
}
