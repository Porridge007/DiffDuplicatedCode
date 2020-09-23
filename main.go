package main

import (
	"DiffCode/util"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"strings"
)

// 全局变量，path1基金和债券路径，path2是股指路径
var path1, path2 string

// 读取配置文件，获得kline所在位置
func init() {
	viper.SetConfigName("conf")   //设置配置文件的名字
	viper.AddConfigPath("config") //添加配置文件所在的路径
	viper.SetConfigType("json")   //设置配置文件类型，可选
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("config file error: %s\n", err)
		os.Exit(1)
	}
	kline := viper.Get("kline").(string)
	if strings.Trim(kline," ") == ""{
		panic("No kline field in conf.json")
	}
	// 特性： 屏蔽路径有无“/”差异
	if strings.HasSuffix(kline, "/") == false {
		kline = kline + "/"
	}
	path1 = kline + "fundAndBond/bfq/"
	path2 = kline + "stockAndIndex/bfq/"
}

func main() {
	// 清除之前旧日志，第一次跑报错 忽略
	if err :=  os.Remove("duplicatedCodes.txt"); err != nil{
		fmt.Println("No Need to Clear the previous log, error:", err)
	}
	// 创建日志文件
	f, err := os.Create("duplicatedCodes.txt")
	if err != nil {
		fmt.Println("Cannot Create the LogFile: duplicatedCodes.txt", err)
	}
	defer f.Close()
	// 获得级别
	levels := GetFilename(path1)
	for _, level := range levels {
		specPath1 := path1 + level
		specPath2 := path2 + level
		// 获取每个级别的文件名
		Code1 := GetFilename(specPath1)
		Code2 := GetFilename(specPath2)
		// 获得重复的代码
		duplicatedCodes := GetDuplicated(Code1, Code2)
		// 写日志
		WriteLog(level,duplicatedCodes)
		// 删除基金债券下的重复代码
		DeleteCodeUponLevel(level, duplicatedCodes)
	}
}

// 读取指定目录下，所有文件名
func GetFilename(path string) []string {
	nameList := make([]string, 0)
	files, err := ioutil.ReadDir(path)
	if err!= nil{
		panic("Cannot Get FileNames Under the Given Dir")
	}
	for _, f := range files {
		nameList = append(nameList, f.Name())
	}
	return nameList
}


// 得到重复代码
func GetDuplicated(Code1 []string, Code2 []string) []string {
	// 集合AB 重复元素为A-(A+B-B)
	// CodeSum = A + B
	codeSum := util.NewStringSet()
	codeSum.Add(Code1...)
	codeSum.Add(Code2...)
	// A+B-B
	codeSum.Remove(Code2...)
	// A-(A+B-B)
	codeA := util.NewStringSet()
	codeA.Add(Code1...)
	codeSumList := codeSum.List()
	codeA.Remove(codeSumList...)
	return codeA.List()
}

// 写日志到duplicatedCodes.txt
func WriteLog(level string, codes []string) {
	if f, err := os.OpenFile("duplicatedCodes.txt", os.O_WRONLY|os.O_APPEND, 666); err != nil {
		fmt.Println("Cannot Open the LogFile: duplicatedCodes.txt")
	} else {
		defer f.Close()
		_, err := f.WriteString(fmt.Sprintf("[%s]\n", level))
		if err != nil {
			fmt.Println(err)
		}
		for _, code := range codes {
			if _, err := f.WriteString(code+" ");err!= nil{
				fmt.Println("Write String to Log Failure," ,err)
			}

		}
		if _, err :=f.WriteString("\n");err!= nil{
			fmt.Println("Write String to Log Failure," ,err)
		}
	}
}

// 同级别 把基金债券底下重复的代码删除
func DeleteCodeUponLevel(level string, codes []string)  {
	for _, code := range codes {
		if err := os.Remove(path1 + level + "/" + code); err !=nil {
			fmt.Println("Remove the duplicated code:", code ," failure:", err)
			return
		}
	}
}
