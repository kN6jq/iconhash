package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/kN6jq/iconhash/internal/hash"
	"github.com/kN6jq/iconhash/pkg/utils"
)

var (
	file    = flag.String("file", "", "包含URL列表的文件路径")
	output  = flag.String("output", "", "输出结果的文件路径")
	threads = flag.Int("threads", 10, "并发线程数")
	url     = flag.String("u", "", "指定单个URL")
)

func main() {
	flag.Parse()

	// 创建增强版HTTP客户端，默认跳过TLS证书验证
	httpClient := utils.CreateEnhancedHTTPClient(true)

	// 创建缓存
	cache, err := utils.CreateCache(100)
	if err != nil {
		log.Fatalf("创建缓存失败: %v", err)
	}

	// 处理输入
	var urls []string
	if *file != "" {
		// 批量处理文件中的URL
		urls, err = readURLsFromFile(*file)
		if err != nil {
			log.Fatalf("读取文件失败: %v", err)
		}
	} else if *url != "" {
		// 使用-u参数指定的URL
		urls = []string{*url}
	} else if len(flag.Args()) > 0 {
		// 处理命令行参数中的URL
		urls = flag.Args()
	} else {
		// 显示帮助信息
		fmt.Println("用法: iconhash [选项] <URL>")
		fmt.Println("选项:")
		fmt.Println("  -u=<url>             指定单个URL")
		fmt.Println("  -file=<filepath>      包含URL列表的文件路径")
		fmt.Println("  -output=<filepath>    输出结果的文件路径")
		fmt.Println("  -threads=<number>     并发线程数 (默认: 10)")
		fmt.Println("\n示例:")
		fmt.Println("  iconhash http://example.com")
		fmt.Println("  iconhash -u=http://example.com")
		fmt.Println("  iconhash -file=urls.txt -output=results.txt")
		os.Exit(1)
	}

	// 准备输出文件
	var outputFile *os.File
	if *output != "" {
		outputFile, err = os.Create(*output)
		if err != nil {
			log.Fatalf("创建输出文件失败: %v", err)
		}
		defer outputFile.Close()

		// 写入UTF-8 BOM，确保Windows系统能正确显示中文
		outputFile.Write([]byte{0xEF, 0xBB, 0xBF})
	}

	// 如果只有一个URL，直接处理
	if len(urls) == 1 {
		processURL(urls[0], httpClient, cache, outputFile)
		return
	}

	// 批量处理多个URL
	fmt.Printf("开始处理 %d 个URL，并发数: %d\n", len(urls), *threads)

	// 使用通道控制并发
	urlChan := make(chan string, len(urls))
	var wg sync.WaitGroup

	// 启动工作协程
	for i := 0; i < *threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for url := range urlChan {
				processURL(url, httpClient, cache, outputFile)
			}
		}()
	}

	// 发送URL到通道
	for _, url := range urls {
		urlChan <- url
	}
	close(urlChan)

	// 等待所有URL处理完成
	wg.Wait()
	fmt.Println("所有URL处理完成")
}

// 处理单个URL
func processURL(url string, httpClient *utils.HTTPClient, cache hash.Cache, outputFile *os.File) {

	// 获取favicon数据
	data, err := httpClient.GetFavicon(url)
	if err != nil {
		fmt.Printf("错误 [%s]: %v\n", url, err)
		return
	}

	// 获取实际的favicon URL
	faviconURL, _ := httpClient.GetLastFaviconURL()

	// 计算哈希
	fofaHash := hash.CalcFofaHash(data)
	hunterHash := hash.CalcHunterHash(data)

	// 提取哈希值（移除引号）
	fofaHashValue := strings.Split(fofaHash, "=")[1]
	fofaHashValue = strings.Trim(fofaHashValue, "\"")

	hunterHashValue := strings.Split(hunterHash, "=")[1]
	hunterHashValue = strings.Trim(hunterHashValue, "\"")

	// 准备输出结果（简化格式）
	result := fmt.Sprintf("%s | %s | %s | %s\n",
		url, faviconURL, fofaHashValue, hunterHashValue)

	// 输出结果
	if outputFile != nil {
		// 输出到文件
		fmt.Fprint(outputFile, result)
	} else {
		// 输出到控制台
		fmt.Print(result)
	}
}

// 从文件读取URL列表
func readURLsFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if url != "" && !strings.HasPrefix(url, "#") {
			urls = append(urls, url)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}
