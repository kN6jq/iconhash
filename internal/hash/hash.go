package hash

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/twmb/murmur3"
)

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

// CalcFofaHash 计算 favicon 的 murmur3 hash (FOFA)
func CalcFofaHash(data []byte) string {
	// 按照 FOFA 的方式进行 base64 编码
	b64 := standBase64(data)

	// 计算 Murmur3 哈希值并转换为有符号整数
	hash := murmur3.Sum32(b64)

	// 返回有符号整数格式
	return fmt.Sprintf("icon_hash=\"%d\"", int32(hash))
}

// CalcHunterHash 计算 favicon 的 md5 hash (Hunter)
func CalcHunterHash(data []byte) string {
	hash := md5.Sum(data)
	md5hex := fmt.Sprintf("%x", hash)
	// 返回带前缀的格式
	return fmt.Sprintf("web.icon=\"%s\"", md5hex)
}

// CalcPureMD5Hash 计算纯 MD5 哈希值（不带前缀）
func CalcPureMD5Hash(data []byte) string {
	hash := md5.Sum(data)
	return fmt.Sprintf("%x", hash)
}

// Cache 简单的LRU缓存接口
type Cache interface {
	Get(key interface{}) (interface{}, bool)
	Add(key, value interface{})
}

// DownloadFavicon 从 URL 下载 favicon
func DownloadFavicon(url string, httpClient *http.Client, cache Cache) ([]byte, error) {
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

// GenerateFofaLink 生成 FOFA 搜索链接
func GenerateFofaLink(hash string) string {
	// 从 hash 中提取纯数字部分
	hashValue := strings.TrimPrefix(strings.TrimSuffix(hash, "\""), "icon_hash=\"")
	// 构造查询语句
	query := fmt.Sprintf("icon_hash=%s", hashValue)
	// Base64 编码
	queryBase64 := base64.StdEncoding.EncodeToString([]byte(query))
	// URL 编码并返回完整链接
	return fmt.Sprintf("https://fofa.info/result?qbase64=%s", url.QueryEscape(queryBase64))
}

// GenerateHunterLink 生成 Hunter 搜索链接
func GenerateHunterLink(hash string) string {
	// 从 hash 中提取纯 MD5 部分
	hashValue := strings.TrimPrefix(strings.TrimSuffix(hash, "\""), "web.icon=\"")
	// 构造查询语句
	query := fmt.Sprintf("web.icon=%s", hashValue)
	// Base64 编码并返回完整链接
	return fmt.Sprintf("https://hunter.qianxin.com/list?search=%s", url.QueryEscape(base64.StdEncoding.EncodeToString([]byte(query))))
}
