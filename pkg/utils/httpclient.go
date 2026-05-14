package utils

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
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

// GetTitle 获取网页标题
func (c *HTTPClient) GetTitle(urlStr string) string {
	// 确保URL以http或https开头
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "http://" + urlStr
	}

	body, _, err := c.getWebContent(urlStr, 0)
	if err != nil {
		return ""
	}

	// 提取标题
	titlePatterns := []string{
		`<title[^>]*>([^<]+)</title>`,
		`<TITLE[^>]*>([^<]+)</TITLE>`,
	}

	for _, pattern := range titlePatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(string(body))
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	return ""
}