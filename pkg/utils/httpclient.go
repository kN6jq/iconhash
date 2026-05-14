package utils

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"
)

// HTTPClient 是一个增强的HTTP客户端
type HTTPClient struct {
	client         *http.Client
	maxRedirs      int
	lastFaviconURL string // 存储最后获取的favicon URL
}

// NewHTTPClient 创建一个新的HTTP客户端
func NewHTTPClient(insecure bool, timeout time.Duration, maxRedirs int) *HTTPClient {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
	return &HTTPClient{
		client:    client,
		maxRedirs: maxRedirs,
	}
}

// GetFavicon 获取网站的favicon
// 如果提供了完整的favicon URL，直接获取
// 否则，尝试从网站获取favicon的URL
func (c *HTTPClient) GetFavicon(urlStr string) ([]byte, string, error) {
	// 处理URL中的 {{favicon}} 模板
	if strings.Contains(urlStr, "{{favicon}}") {
		// 提取协议和主机部分，替换 {{favicon}} 为 /favicon.ico
		u, err := url.Parse(urlStr)
		if err == nil {
			// 替换 {{favicon}} 部分
			path := strings.ReplaceAll(u.Path, "{{favicon}}", "/favicon.ico")
			// 清理双斜杠，但保留协议后的双斜杠
			path = strings.ReplaceAll(path, "//", "/")
			u.Path = path
			urlStr = u.String()
		}
	}

	// 检查是否是直接的favicon URL
	if strings.Contains(urlStr, "favicon") || strings.HasSuffix(urlStr, ".ico") {
		c.lastFaviconURL = urlStr
		data, err := c.downloadFile(urlStr)
		return data, urlStr, err
	}

	// 否则，尝试从网站获取favicon URL
	faviconURL, err := c.findFaviconURL(urlStr)
	if err != nil {
		return nil, "", err
	}

	c.lastFaviconURL = faviconURL
	// 下载favicon
	data, err := c.downloadFile(faviconURL)
	return data, faviconURL, err
}

// GetLastFaviconURL 返回最后获取的favicon URL
func (c *HTTPClient) GetLastFaviconURL() (string, bool) {
	if c.lastFaviconURL == "" {
		return "", false
	}
	return c.lastFaviconURL, true
}

// findFaviconURL 从网站获取favicon的URL
func (c *HTTPClient) findFaviconURL(urlStr string) (string, error) {
	// 确保URL以http或https开头
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "http://" + urlStr
	}

	// 获取网页内容
	body, finalURL, err := c.getWebContent(urlStr, 0)
	if err != nil {
		return "", err
	}

	// 从网页内容中提取favicon URL
	faviconURL := c.getFaviconURL(string(body), finalURL)
	return faviconURL, nil
}

// getWebContent 获取网页内容，支持处理JS重定向
func (c *HTTPClient) getWebContent(urlStr string, redirectCount int) ([]byte, string, error) {
	if redirectCount > c.maxRedirs {
		return nil, "", ErrTooManyRedirects
	}

	resp, err := c.client.Get(urlStr)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	// 检查JS重定向
	jsRedirect := c.checkJSRedirect(string(body), urlStr)
	if jsRedirect != "" {
		return c.getWebContent(jsRedirect, redirectCount+1)
	}

	return body, urlStr, nil
}

// downloadFile 下载文件
func (c *HTTPClient) downloadFile(urlStr string) ([]byte, error) {
	resp, err := c.client.Get(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// checkJSRedirect 检查JS重定向
func (c *HTTPClient) checkJSRedirect(htmlContent, baseURL string) string {
	redirectPatterns := []string{
		`window\.parent\.location\.href\s*=\s*['"](.*?)['"];`,
		`window\.location\.href\s*=\s*['"](.*?)['"];`,
		`window\.top\.location\s*=\s*['"](.*?)['"];`,
		`window\.location\s*=\s*['"](.*?)['"];`,
		`location\.href\s*=\s*['"](.*?)['"];`,
		`location\s*=\s*['"](.*?)['"];`,
		`eval\("window\.".*?\.location\s*=\s*['"](.*?)['"]\);`,
		`<meta\s+http-equiv=["']refresh["']\s+content=["']\d*;\s*url=(.*?)["']`,
	}

	for _, pattern := range redirectPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(htmlContent)
		if len(matches) > 1 {
			redirectURL := matches[1]
			if !strings.HasPrefix(redirectURL, "http") {
				baseURLObj, err := url.Parse(baseURL)
				if err != nil {
					continue
				}
				redirectURLObj, err := url.Parse(redirectURL)
				if err != nil {
					continue
				}
				redirectURL = baseURLObj.ResolveReference(redirectURLObj).String()
			}
			return redirectURL
		}
	}
	return ""
}

// getFaviconURL 获取favicon URL
func (c *HTTPClient) getFaviconURL(body, urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	baseURL := u.Scheme + "://" + u.Host

	// 使用更精确的正则表达式提取favicon路径
	// 尝试多种常见的favicon链接模式
	faviconPatterns := []string{
		`<link[^>]*rel=["'](?:shortcut )?icon["'][^>]*href=["']([^"']+)["']`,
		`<link[^>]*href=["']([^"']+)["'][^>]*rel=["'](?:shortcut )?icon["']`,
		`<link[^>]*rel=["']apple-touch-icon["'][^>]*href=["']([^"']+)["']`,
		`<link[^>]*href=["']([^"']+favicon\.[^"']+)["']`,
	}

	for _, pattern := range faviconPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(body)
		if len(matches) > 1 {
			fav := matches[1]
			// 处理相对路径和绝对路径
			switch {
			case strings.HasPrefix(fav, "//"):
				return u.Scheme + ":" + fav
			case strings.HasPrefix(fav, "http"):
				return fav
			case strings.HasPrefix(fav, "/"):
				return baseURL + fav
			default:
				// 处理相对当前目录的路径
				baseDir := ""
				if i := strings.LastIndex(u.Path, "/"); i > 0 {
					baseDir = u.Path[:i+1]
				} else {
					baseDir = "/"
				}
				return baseURL + baseDir + fav
			}
		}
	}

	// 如果没有找到favicon标签，则尝试默认路径
	return baseURL + "/favicon.ico"
}

// supportedTitleMimeTypes 支持提取标题的MIME类型
var supportedTitleMimeTypes = []string{
	"text/html",
	"application/xhtml+xml",
	"application/xml",
	"application/rss+xml",
	"application/atom+xml",
	"application/vnd.wap.xhtml+xml",
}

var (
	reTitle      = regexp.MustCompile(`(?im)<\s*title.*>(.*?)<\s*/\s*title>`)
	reContentType = regexp.MustCompile(`(?im)\s*charset="(.*?)"|charset=(.*?)\s*`)
)

// GetTitle 获取网页标题
func (c *HTTPClient) GetTitle(urlStr string) string {
	// 确保URL以http或https开头
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "http://" + urlStr
	}

	// 处理URL中的 {{favicon}} 模板 - 去掉这部分获取实际页面
	if strings.Contains(urlStr, "{{favicon}}") {
		u, err := url.Parse(urlStr)
		if err == nil {
			// 去掉 {{favicon}} 部分，获取基础URL
			path := strings.ReplaceAll(u.Path, "{{favicon}}", "")
			path = strings.TrimSuffix(path, "/")
			u.Path = path
			urlStr = u.String()
		}
	}

	body, contentType, err := c.getWebContentWithContentType(urlStr, 0)
	if err != nil {
		return ""
	}

	// 解码内容（处理各种编码）
	body = decodeData(body, contentType)

	// 提取标题
	title := extractTitle(string(body))

	return title
}

// getWebContentWithContentType 获取网页内容并返回Content-Type
func (c *HTTPClient) getWebContentWithContentType(urlStr string, redirectCount int) ([]byte, string, error) {
	if redirectCount > c.maxRedirs {
		return nil, "", ErrTooManyRedirects
	}

	resp, err := c.client.Get(urlStr)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	contentType := resp.Header.Get("Content-Type")

	// 检查JS重定向
	jsRedirect := c.checkJSRedirect(string(body), urlStr)
	if jsRedirect != "" {
		return c.getWebContentWithContentType(jsRedirect, redirectCount+1)
	}

	return body, contentType, nil
}

// extractTitle 从HTML中提取标题
func extractTitle(htmlContent string) string {
	// 尝试用DOM方式解析
	titleDom, err := getTitleWithDom(htmlContent)
	if err != nil {
		// 回退到正则
		for _, match := range reTitle.FindAllString(htmlContent, -1) {
			title := trimTitleTags(match)
			title = html.UnescapeString(strings.TrimSpace(title))
			return title
		}
		return ""
	}

	// 用DOM方式提取
	title := renderNode(titleDom)
	title = trimTitleTags(title)
	title = html.UnescapeString(strings.TrimSpace(title))

	// 清理换行符等
	title = strings.Trim(title, "\n\t\v\f\r")
	title = strings.ReplaceAll(title, "\n", " ")
	title = strings.ReplaceAll(title, "\t", " ")
	title = strings.ReplaceAll(title, "\r", " ")

	return strings.TrimSpace(title)
}

// getTitleWithDom 使用DOM方式提取标题
func getTitleWithDom(htmlContent string) (*html.Node, error) {
	var title *html.Node
	var crawler func(*html.Node)
	crawler = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "title" {
			title = node
			return
		}
		for child := node.FirstChild; child != nil && title == nil; child = child.NextSibling {
			crawler(child)
		}
	}

	htmlDoc, err := html.Parse(bytes.NewReader([]byte(htmlContent)))
	if err != nil {
		return nil, err
	}
	crawler(htmlDoc)
	if title != nil {
		return title, nil
	}
	return nil, fmt.Errorf("title not found")
}

// renderNode 渲染HTML节点为字符串
func renderNode(n *html.Node) string {
	var buf bytes.Buffer
	w := io.Writer(&buf)
	html.Render(w, n)
	return buf.String()
}

// trimTitleTags 去掉title标签
func trimTitleTags(title string) string {
	begin := strings.Index(title, ">")
	end := strings.Index(title, "</")
	if end < 0 || begin < 0 || end <= begin {
		return title
	}
	return title[begin+1 : end]
}

// decodeData 根据Content-Type解码内容
func decodeData(data []byte, contentType string) []byte {
	contentType = strings.ToLower(contentType)

	// 先检查HTTP头的Content-Type
	if strings.Contains(contentType, "charset=gb2312") || strings.Contains(contentType, "charset=gbk") {
		return decodeGBK(data)
	}
	if strings.Contains(contentType, "euc-kr") {
		return decodeKorean(data)
	}
	if strings.Contains(contentType, "big5") {
		return decodeBig5(data)
	}

	// 再检查HTML中的meta charset
	match := reContentType.FindSubmatch(data)
	if len(match) > 0 {
		for i := 1; i < len(match); i++ {
			charset := string(match[i])
			charset = strings.ToLower(strings.TrimSpace(charset))
			if strings.Contains(charset, "gb2312") || strings.Contains(charset, "gbk") {
				return decodeGBK(data)
			}
			if strings.Contains(charset, "big5") {
				return decodeBig5(data)
			}
		}
	}

	return data
}

// decodeGBK 解码GBK为UTF-8
func decodeGBK(data []byte) []byte {
	reader := transform.NewReader(bytes.NewReader(data), simplifiedchinese.GBK.NewDecoder())
	result, err := io.ReadAll(reader)
	if err != nil {
		return data
	}
	return result
}

// decodeBig5 解码Big5为UTF-8
func decodeBig5(data []byte) []byte {
	reader := transform.NewReader(bytes.NewReader(data), traditionalchinese.Big5.NewDecoder())
	result, err := io.ReadAll(reader)
	if err != nil {
		return data
	}
	return result
}

// decodeKorean 解码Korean为UTF-8
func decodeKorean(data []byte) []byte {
	reader := transform.NewReader(bytes.NewReader(data), korean.EUCKR.NewDecoder())
	result, err := io.ReadAll(reader)
	if err != nil {
		return data
	}
	return result
}