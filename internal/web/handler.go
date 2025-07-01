package web

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/kN6jq/iconhash/internal/hash"
)

// Handler 定义Web模式处理器
type Handler struct {
	httpClient *http.Client
	cache      hash.Cache
	staticFS   fs.FS
}

// NewHandler 创建一个新的Web处理器
func NewHandler(httpClient *http.Client, cache hash.Cache, staticFS fs.FS) (*Handler, error) {
	return &Handler{
		httpClient: httpClient,
		cache:      cache,
		staticFS:   staticFS,
	}, nil
}

// SetupRoutes 设置HTTP路由
func (h *Handler) SetupRoutes() {
	// 设置静态文件服务器
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(h.staticFS))))

	// 设置主页和计算处理器
	http.HandleFunc("/", h.HomeHandler)
	http.HandleFunc("/calculate", h.CalcHandler)
}

// HomeHandler 处理主页请求
func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("index").Parse(HTMLTemplate)
	if err != nil {
		http.Error(w, "内部错误", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// CalcHandler 处理计算请求
func (h *Handler) CalcHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "请提供 favicon URL 参数", http.StatusBadRequest)
		return
	}

	data, err := hash.DownloadFavicon(url, h.httpClient, h.cache)
	if err != nil {
		http.Error(w, "下载或处理 favicon 失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fofaHash := hash.CalcFofaHash(data)
	hunterHash := hash.CalcHunterHash(data)
	pureMD5 := hash.CalcPureMD5Hash(data)

	fofaLink := hash.GenerateFofaLink(fofaHash)
	hunterLink := hash.GenerateHunterLink(hunterHash)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "%s\n%s\n%s\n%s\n%s\n",
		fofaHash, fofaLink,
		hunterHash, hunterLink,
		pureMD5)
}
