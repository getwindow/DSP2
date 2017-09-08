// utils project utils.go
package utils

import (
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

//int 类型转为 uint8
func Unit8FormInt(x int) (uint8, error) {
	if 0 <= x && x <= math.MaxUint8 {
		return uint8(x), nil
	}
	return 0, fmt.Errorf("%d is out of the uint8 range")
}

var Path string = GetCurrentDirectory()

//写入日志
func Logdebug(h, c string) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("日志写入错误")
		}
	}()

	lock := sync.RWMutex{}
	lock.Lock()
	nowt := time.Now()
	pp := fmt.Sprintf("%d%d%d", nowt.Year(), nowt.Month(), nowt.Day())
	ih, _ := PathExists(Path + "/logs/" + pp)
	ofile := Path + "/logs/" + h + ".log"
	if ih == true {
		ofile = Path + "/logs/" + pp + "/" + h + ".log"
	}
	fl, err := os.OpenFile(ofile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0744)
	if err != nil {
		fmt.Println("cuowu", err)
		return
	}

	n, err := fl.Write([]byte(c + "\r" + "\n"))
	if err == nil && n < len(c) {
		err = io.ErrShortWrite
	}
	defer fl.Close()
	lock.Unlock()
}

/*
获取程序运行路径
*/
func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println("This is an error")
	}
	return strings.Replace(dir, "\\", "/", -1)
}

//判断文件
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		er := os.MkdirAll(path, os.ModePerm)
		if er != nil {
			return false, nil
		} else {
			return true, nil
		}
	}
	return false, err
}
