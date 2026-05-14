package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/kN6jq/iconhash/internal/hash"
	"github.com/kN6jq/iconhash/pkg/utils"
)

var (
	urlFlag     = flag.String("u", "", "指定单个URL地址")
	fileFlag    = flag.String("f", "", "指定URL列表文件")
	concurrency = flag.Int("c", 10, "并发数")
)

type Result struct {
	OriginalURL string
	FaviconURL  string
	Title       string
	FofaHash    string
	HunterHash  string
	MD5         string
	Error       string
}

func main() {
	flag.Usage = func() {
		fmt.Println("用法: iconhash [选项]")
		fmt.Println("选项:")
		fmt.Println("  -u=<url>   指定单个URL地址")
		fmt.Println("  -f=<file>  指定URL列表文件")
		fmt.Println("  -c=<num>   并发数 (默认: 10)")
		fmt.Println("\n示例:")
		fmt.Println("  iconhash -u=http://example.com")
		fmt.Println("  iconhash -f=urls.txt")
		fmt.Println("  iconhash -f=urls.txt -c=20")
	}

	flag.Parse()

	if *urlFlag == "" && *fileFlag == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *urlFlag != "" && *fileFlag != "" {
		fmt.Println("错误: -u 和 -f 不能同时使用")
		flag.Usage()
		os.Exit(1)
	}

	var results []Result

	if *urlFlag != "" {
		// 单个URL模式
		result := processURL(*urlFlag)
		results = append(results, result)
		printResult(result)
	} else {
		// 批量处理模式
		results = processFile(*fileFlag, *concurrency)
	}

	// 自动输出到CSV (和输入文件同目录，扩展名改为.csv)
	if *fileFlag != "" && len(results) > 0 {
		csvPath := filepath.Base(*fileFlag)
		csvPath = strings.TrimSuffix(csvPath, filepath.Ext(csvPath)) + "_iconhash.csv"
		writeCSV(csvPath, results)
		fmt.Printf("\n结果已保存到: %s\n", csvPath)
	}
}

func processFile(filePath string, concurrency int) []Result {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("打开文件失败: %v", err)
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			urls = append(urls, line)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("读取文件失败: %v", err)
	}

	fmt.Printf("开始处理 %d 个URL，并发数: %d\n", len(urls), concurrency)

	var wg sync.WaitGroup
	urlChan := make(chan string, len(urls))
	resultChan := make(chan Result, len(urls))

	// 启动workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for u := range urlChan {
				resultChan <- processURL(u)
			}
		}()
	}

	// 发送所有URL
	go func() {
		for _, u := range urls {
			urlChan <- u
		}
		close(urlChan)
	}()

	// 收集结果
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var results []Result
	for r := range resultChan {
		results = append(results, r)
		printResult(r)
	}

	return results
}

func processURL(urlStr string) Result {
	result := Result{OriginalURL: urlStr}

	httpClient := utils.CreateEnhancedHTTPClient(true)

	// 获取favicon
	data, faviconURL, err := httpClient.GetFavicon(urlStr)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.FaviconURL = faviconURL

	// 获取标题
	result.Title = httpClient.GetTitle(urlStr)

	// 计算哈希
	result.FofaHash = strings.Trim(strings.Split(hash.CalcFofaHash(data), "=")[1], "\"")
	result.HunterHash = strings.Trim(strings.Split(hash.CalcHunterHash(data), "=")[1], "\"")
	result.MD5 = hash.CalcPureMD5Hash(data)

	return result
}

func printResult(r Result) {
	if r.Error != "" {
		fmt.Printf("错误 [%s]: %s\n", r.OriginalURL, r.Error)
	} else {
		titleInfo := ""
		if r.Title != "" {
			titleInfo = " | " + r.Title
		}
		fmt.Printf("%s | %s | %s | %s%s\n", r.OriginalURL, r.FaviconURL, r.FofaHash, r.HunterHash, titleInfo)
	}
}

func writeCSV(filePath string, results []Result) {
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("创建CSV文件失败: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	writer.Write([]string{"URL", "Favicon URL", "Title", "FOFA Hash", "Hunter Hash", "MD5", "Error"})

	// 写入数据
	for _, r := range results {
		writer.Write([]string{
			r.OriginalURL,
			r.FaviconURL,
			r.Title,
			r.FofaHash,
			r.HunterHash,
			r.MD5,
			r.Error,
		})
	}
}