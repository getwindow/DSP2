// controllers project controllers.go
package controllers

import (
	"fmt"

	"io/ioutil"

	"DSP2/data/today"
	"DSP2/models"

	"github.com/golang/protobuf/proto"
)

func IM() {
	dat, err := ioutil.ReadFile("bit.req")
	if err != nil {
		fmt.Println(err)
	}

	req := &today.BidRequest{}
	era := proto.Unmarshal(dat, req)
	if era != nil {
		fmt.Println("转化错误")
	}
	tod := models.BidReq{}
	for i := 0; i < 10000; i++ {
		go tod.GetPlan(req)
		fmt.Println("当前运行到", i)
	}
	//TuiHead()
}
