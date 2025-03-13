package main

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/base64"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/twmb/murmur3"
)

var (
	mode       = flag.String("mode", "terminal", "运行模式: terminal 或 web")
	port       = flag.String("port", "8080", "Web 模式下的端口号")
	insecure   = flag.Bool("insecure", true, "是否跳过 TLS 证书验证")
	cache, _   = simplelru.NewLRU(100, nil) // 简单缓存，避免重复下载
	httpClient *http.Client
)

// 初始化 HTTP 客户端
func initHTTPClient() {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: *insecure},
	}
	httpClient = &http.Client{Transport: transport}
}

// 按照 FOFA 的方式进行 base64 编码（每76个字符后添加换行符）
func standBase64(data []byte) []byte {
	// 标准 base64 编码
	bckd := base64.StdEncoding.EncodeToString(data)

	// 每76个字符添加一个换行符
	var buffer bytes.Buffer
	for i := 0; i < len(bckd); i++ {
		ch := bckd[i]
		buffer.WriteByte(ch)
		if (i+1)%76 == 0 {
			buffer.WriteByte('\n')
		}
	}
	buffer.WriteByte('\n')

	return buffer.Bytes()
}

// 计算 favicon 的 murmur3 hash (FOFA)
func calcFofaHash(data []byte) string {
	// 按照 FOFA 的方式进行 base64 编码
	b64 := standBase64(data)

	// 计算 Murmur3 哈希值并转换为有符号整数
	hash := murmur3.Sum32(b64)

	// 返回有符号整数格式
	return fmt.Sprintf("icon_hash=\"%d\"", int32(hash))
}

// 计算 favicon 的 md5 hash (Hunter)
func calcHunterHash(data []byte) string {
	hash := md5.Sum(data)
	md5hex := fmt.Sprintf("%x", hash)
	// 返回带前缀的格式
	return fmt.Sprintf("web.icon=\"%s\"", md5hex)
}

// 计算纯 MD5 哈希值（不带前缀）
func calcPureMD5Hash(data []byte) string {
	hash := md5.Sum(data)
	return fmt.Sprintf("%x", hash)
}

// 从 URL 下载 favicon
func downloadFavicon(url string) ([]byte, error) {
	if val, ok := cache.Get(url); ok {
		return val.([]byte), nil
	}

	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("下载 favicon 失败: %v", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取 favicon 数据失败: %v", err)
	}

	cache.Add(url, data)
	return data, nil
}

// 终端模式
func terminalMode() {
	if len(os.Args) < 2 {
		fmt.Println("用法: favicon-hash <favicon_url>")
		fmt.Println("示例: favicon-hash http://example.com/favicon.ico")
		os.Exit(1)
	}

	url := os.Args[len(os.Args)-1]
	data, err := downloadFavicon(url)
	if err != nil {
		log.Fatal(err)
	}

	fofaHash := calcFofaHash(data)
	hunterHash := calcHunterHash(data)
	pureMD5 := calcPureMD5Hash(data)

	fofaLink := generateFofaLink(fofaHash)
	hunterLink := generateHunterLink(hunterHash)

	fmt.Printf("%s\n", fofaHash)
	fmt.Printf("FOFA 搜索链接: %s\n\n", fofaLink)

	fmt.Printf("%s\n", hunterHash)
	fmt.Printf("Hunter 搜索链接: %s\n\n", hunterLink)

	fmt.Printf("纯 MD5: %s\n", pureMD5)
}

// Web 模式的 HTML 模板
var htmlTemplate = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Favicon Hash 计算器</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
        }
        .input-container {
            margin-bottom: 20px;
        }
        input[type="text"] {
            width: 70%;
            padding: 8px;
            margin-right: 10px;
        }
        button {
            padding: 8px 16px;
            background-color: #4CAF50;
            color: white;
            border: none;
            cursor: pointer;
        }
        button:hover {
            background-color: #45a049;
        }
        .result {
            margin-top: 20px;
            padding: 10px;
            border: 1px solid #ddd;
            display: none;
        }
        .error {
            color: red;
        }
    </style>
</head>
<body>
    <h1>Favicon Hash 计算器</h1>
    <div class="input-container">
        <input type="text" id="urlInput" placeholder="请输入 favicon URL (如 http://example.com/favicon.ico)">
        <button onclick="calculateHash()">计算</button>
    </div>
    <div id="result" class="result">
        <h3>计算结果:</h3>
        <p><strong>FOFA:</strong> <span id="fofaHash"></span></p>
        <p><strong>FOFA 搜索链接:</strong> <a id="fofaLink" href="#" target="_blank">点击跳转到 FOFA 搜索</a></p>
        <p><strong>Hunter:</strong> <span id="hunterHash"></span></p>
        <p><strong>Hunter 搜索链接:</strong> <a id="hunterLink" href="#" target="_blank">点击跳转到 Hunter 搜索</a></p>
    </div>
    <div id="error" class="error"></div>

    <script>
        async function calculateHash() {
            const url = document.getElementById("urlInput").value.trim();
            const resultDiv = document.getElementById("result");
            const errorDiv = document.getElementById("error");
            const fofaSpan = document.getElementById("fofaHash");
            const hunterSpan = document.getElementById("hunterHash");
            const fofaLink = document.getElementById("fofaLink");
            const hunterLink = document.getElementById("hunterLink");

            // 清空之前的错误和结果
            errorDiv.textContent = "";
            resultDiv.style.display = "none";

            if (!url) {
                errorDiv.textContent = "请输入有效的 favicon URL";
                return;
            }

            try {
                const response = await fetch("/calculate?url=" + encodeURIComponent(url));
                if (!response.ok) {
                    throw new Error(await response.text());
                }
                const text = await response.text();
                const lines = text.split("\n");
                
                // 设置哈希值和链接
                fofaSpan.textContent = lines[0];
                fofaLink.href = lines[1];
                
                hunterSpan.textContent = lines[2];
                hunterLink.href = lines[3];
                
                resultDiv.style.display = "block";
            } catch (error) {
                errorDiv.textContent = "错误: " + error.message;
            }
        }
    </script>
</body>
</html>
`

// Web 主页处理器
func webHomeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("index").Parse(htmlTemplate)
	if err != nil {
		http.Error(w, "内部错误", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// Web 计算处理器
func webCalcHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "请提供 favicon URL 参数", http.StatusBadRequest)
		return
	}

	data, err := downloadFavicon(url)
	if err != nil {
		http.Error(w, "下载或处理 favicon 失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fofaHash := calcFofaHash(data)
	hunterHash := calcHunterHash(data)

	fofaLink := generateFofaLink(fofaHash)
	hunterLink := generateHunterLink(hunterHash)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "%s\n%s\n%s\n%s\n",
		fofaHash, fofaLink,
		hunterHash, hunterLink)
}

// 生成 FOFA 搜索链接
func generateFofaLink(hash string) string {
	// 从 hash 中提取纯数字部分
	hashValue := strings.TrimPrefix(strings.TrimSuffix(hash, "\""), "icon_hash=\"")
	// 构造查询语句
	query := fmt.Sprintf("icon_hash=%s", hashValue)
	// Base64 编码
	queryBase64 := base64.StdEncoding.EncodeToString([]byte(query))
	// URL 编码并返回完整链接
	return fmt.Sprintf("https://fofa.info/result?qbase64=%s", url.QueryEscape(queryBase64))
}

// 生成 Hunter 搜索链接
func generateHunterLink(hash string) string {
	// 从 hash 中提取纯 MD5 部分
	hashValue := strings.TrimPrefix(strings.TrimSuffix(hash, "\""), "web.icon=\"")
	// 构造查询语句
	query := fmt.Sprintf("web.icon=%s", hashValue)
	// Base64 编码并返回完整链接
	return fmt.Sprintf("https://hunter.qianxin.com/list?search=%s", url.QueryEscape(base64.StdEncoding.EncodeToString([]byte(query))))
}

func main() {
	flag.Parse()

	// 初始化 HTTP 客户端
	initHTTPClient()

	switch strings.ToLower(*mode) {
	case "terminal":
		terminalMode()
	case "web":
		http.HandleFunc("/", webHomeHandler)
		http.HandleFunc("/calculate", webCalcHandler)
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
