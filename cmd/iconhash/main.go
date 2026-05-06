package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/kN6jq/iconhash/internal/hash"
	"github.com/kN6jq/iconhash/pkg/utils"
)

var (
	urlFlag  = flag.String("u", "", "指定URL地址（网站地址或favicon直链）")
	fileFlag = flag.String("f", "", "指定本地favicon图片路径")
)

func main() {
	flag.Usage = func() {
		fmt.Println("用法: iconhash [选项]")
		fmt.Println("选项:")
		fmt.Println("  -u=<url>   指定URL地址，支持网站地址或favicon直链")
		fmt.Println("  -f=<file>  指定本地favicon图片路径")
		fmt.Println("\n示例:")
		fmt.Println("  iconhash -u=http://example.com              自动发现favicon")
		fmt.Println("  iconhash -u=http://example.com/favicon.ico  直接获取favicon")
		fmt.Println("  iconhash -f=/path/to/favicon.ico")
	}

	flag.Parse()

	// 检查参数
	if *urlFlag == "" && *fileFlag == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *urlFlag != "" && *fileFlag != "" {
		fmt.Println("错误: -u 和 -f 不能同时使用")
		flag.Usage()
		os.Exit(1)
	}

	var data []byte
	var source string
	var faviconURL string

	if *urlFlag != "" {
		// 从URL获取favicon
		httpClient := utils.CreateEnhancedHTTPClient(true)
		var err error
		data, err = httpClient.GetFavicon(*urlFlag)
		if err != nil {
			log.Fatalf("获取favicon失败: %v", err)
		}
		source = *urlFlag
		if u, ok := httpClient.GetLastFaviconURL(); ok {
			faviconURL = u
		}
	} else {
		// 从本地文件读取
		var err error
		data, err = os.ReadFile(*fileFlag)
		if err != nil {
			log.Fatalf("读取文件失败: %v", err)
		}
		source = *fileFlag
	}

	// 计算哈希
	fofaHash := hash.CalcFofaHash(data)
	hunterHash := hash.CalcHunterHash(data)
	md5Hash := hash.CalcPureMD5Hash(data)

	// 提取哈希值
	fofaHashValue := strings.Trim(strings.Split(fofaHash, "=")[1], "\"")
	hunterHashValue := strings.Trim(strings.Split(hunterHash, "=")[1], "\"")

	// 输出结果
	fmt.Printf("来源: %s\n", source)
	if faviconURL != "" && faviconURL != source {
		fmt.Printf("Favicon: %s\n", faviconURL)
	}
	fmt.Printf("FOFA Hash: %s\n", fofaHashValue)
	fmt.Printf("Hunter Hash: %s\n", hunterHashValue)
	fmt.Printf("MD5: %s\n", md5Hash)
}
