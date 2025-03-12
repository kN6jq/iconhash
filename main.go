package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/twmb/murmur3"
)

var (
	mode     = flag.String("mode", "terminal", "运行模式: terminal 或 web")
	port     = flag.String("port", "8080", "Web 模式下的端口号")
	cache, _ = simplelru.NewLRU(100, nil) // 简单缓存，避免重复下载
)

// 计算 favicon 的 murmur3 hash (FOFA)
func calcFofaHash(data []byte) string {
	hash := murmur3.Sum32(data)
	return fmt.Sprintf("icon_hash=\"%d\"", int32(hash)) // 添加 icon_hash= 前缀
}

// 计算 favicon 的 md5 hash (Hunter)
func calcHunterHash(data []byte) string {
	hash := md5.Sum(data)
	return fmt.Sprintf("web.icon=\"%x\"", hash) // 添加 web.icon= 前缀
}

// 从 URL 下载 favicon
func downloadFavicon(url string) ([]byte, error) {
	if val, ok := cache.Get(url); ok {
		return val.([]byte), nil
	}

	resp, err := http.Get(url)
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

	url := os.Args[1]
	data, err := downloadFavicon(url)
	if err != nil {
		log.Fatal(err)
	}

	fofaHash := calcFofaHash(data)
	hunterHash := calcHunterHash(data)

	fmt.Printf("%s\n", fofaHash)
	fmt.Printf("%s\n", hunterHash)
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
        <p><strong>Hunter:</strong> <span id="hunterHash"></span></p>
    </div>
    <div id="error" class="error"></div>

    <script>
        async function calculateHash() {
            const url = document.getElementById("urlInput").value.trim();
            const resultDiv = document.getElementById("result");
            const errorDiv = document.getElementById("error");
            const fofaSpan = document.getElementById("fofaHash");
            const hunterSpan = document.getElementById("hunterHash");

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
                fofaSpan.textContent = lines[0];
                hunterSpan.textContent = lines[1];
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

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "%s\n%s\n", fofaHash, hunterHash)
}

func main() {
	flag.Parse()

	switch strings.ToLower(*mode) {
	case "terminal":
		terminalMode()
	case "web":
		http.HandleFunc("/", webHomeHandler)
		http.HandleFunc("/calculate", webCalcHandler)
		log.Printf("Web 服务器启动在 :%s", *port)
		log.Printf("访问: http://localhost:%s", *port)
		log.Fatal(http.ListenAndServe(":"+*port, nil))
	default:
		log.Fatalf("无效的模式: %s，支持的模式: terminal, web", *mode)
	}
}
