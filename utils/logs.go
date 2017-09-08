package utils

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

var Logchannel = make(chan map[string]string, 2000) //缓冲区数据对象存储

//使用缓冲区日志
func LogInfo(content string) {
	//判断文件是否存在
	mp := make(map[string]string)
	mp["date"] = string(time.Now().Year())
	mp["content"] = content
	Logchannel <- mp

}

//创建日志接口的实现 很好的完成了日志的记录功能
func LogSave() {
	fmt.Println("开始日志")
	//关闭当前的文件 守护进程一样的，保存日志信息
	time.Sleep(2 * time.Second)
	//nowt := time.Now()
	Path := GetCurrentDirectory()
	//	pp := fmt.Sprintf("%d%d%d", nowt.Year(), nowt.Month(), nowt.Day())
	//	ih, _ := PathExists(Path + "/logs/" + pp)
	//获取当前的进程号
	h := "logs%d"
	file_path := fmt.Sprintf(h, os.Getpid())
	ofile := Path + "/logs/" + file_path + ".log"

	//	if ih == true {
	//		ofile = Path + "/logs/" + pp + "/" + file_path + ".log"
	//	}

	fl, err := os.OpenFile(ofile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0744)
	if err != nil {
		return
	}

	bf := bufio.NewWriterSize(fl, 1024*20)

	defer func() {
		bf.Flush() //缓冲区的数据写入如文件
		fl.Close()
	}()
	tout := time.After(time.Second * 2)
	for {
		select {
		case content, ok := <-Logchannel:
			if ok == true {
				bf.WriteString(content["content"] + "\n")
			}
			bufferid := bf.Buffered()
			if bufferid > 1024*20 {
				bf.Flush()
			}
		case <-tout:
			bf.Flush()
			tout = time.After(time.Second * 2) //每个2秒主动数据同步

		default:
			time.Sleep(time.Microsecond * 1)
		}
	}
}
