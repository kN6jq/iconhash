package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/kN6jq/iconhash/internal/hash"
	"github.com/kN6jq/iconhash/internal/web"
	"github.com/kN6jq/iconhash/pkg/utils"
)

var (
	mode     = flag.String("mode", "terminal", "运行模式: terminal 或 web")
	port     = flag.String("port", "8080", "Web 模式下的端口号")
	insecure = flag.Bool("insecure", true, "是否跳过 TLS 证书验证")
)

// 终端模式
func terminalMode(httpClient *http.Client, cache hash.Cache) {
	if len(os.Args) < 2 {
		fmt.Println("用法: iconhash <favicon_url>")
		fmt.Println("示例: iconhash http://example.com/favicon.ico")
		os.Exit(1)
	}

	url := os.Args[len(os.Args)-1]
	data, err := hash.DownloadFavicon(url, httpClient, cache)
	if err != nil {
		log.Fatal(err)
	}

	fofaHash := hash.CalcFofaHash(data)
	hunterHash := hash.CalcHunterHash(data)
	pureMD5 := hash.CalcPureMD5Hash(data)

	fofaLink := hash.GenerateFofaLink(fofaHash)
	hunterLink := hash.GenerateHunterLink(hunterHash)

	fmt.Printf("%s\n", fofaHash)
	fmt.Printf("FOFA 搜索链接: %s\n\n", fofaLink)

	fmt.Printf("%s\n", hunterHash)
	fmt.Printf("Hunter 搜索链接: %s\n\n", hunterLink)

	fmt.Printf("纯 MD5: %s\n", pureMD5)
}

func main() {
	flag.Parse()

	// 创建 HTTP 客户端
	httpClient := utils.CreateHTTPClient(*insecure)

	// 创建缓存
	cache, err := utils.CreateCache(100)
	if err != nil {
		log.Fatalf("创建缓存失败: %v", err)
	}

	switch strings.ToLower(*mode) {
	case "terminal":
		terminalMode(httpClient, cache)
	case "web":
		// 获取静态资产
		assetsFS, err := GetAssetsFS()
		if err != nil {
			log.Fatalf("获取静态资产失败: %v", err)
		}

		// 创建Web处理器
		handler, err := web.NewHandler(httpClient, cache, assetsFS)
		if err != nil {
			log.Fatalf("创建Web处理器失败: %v", err)
		}

		// 设置路由
		handler.SetupRoutes()

		// 启动Web服务器
		log.Printf("Web 服务器启动在 :%s", *port)
		log.Printf("访问: http://localhost:%s", *port)
		if *insecure {
			log.Printf("警告: TLS 证书验证已禁用")
		}
		log.Fatal(http.ListenAndServe(":"+*port, nil))
	default:
		log.Fatalf("无效的模式: %s，支持的模式: terminal, web", *mode)
	}
}
