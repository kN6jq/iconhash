package utils

import (
	"crypto/tls"
	"net/http"

	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/kN6jq/iconhash/internal/hash"
)

// CreateHTTPClient 创建 HTTP 客户端，可选是否跳过证书验证
func CreateHTTPClient(insecure bool) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	}
	return &http.Client{Transport: transport}
}

// CreateCache 创建LRU缓存
func CreateCache(size int) (hash.Cache, error) {
	cache, err := simplelru.NewLRU(size, nil)
	if err != nil {
		return nil, err
	}

	return &lruCacheWrapper{cache: cache}, nil
}

// LRU缓存的封装，实现了hash.Cache接口
type lruCacheWrapper struct {
	cache *simplelru.LRU
}

func (w *lruCacheWrapper) Get(key interface{}) (interface{}, bool) {
	return w.cache.Get(key)
}

func (w *lruCacheWrapper) Add(key, value interface{}) {
	w.cache.Add(key, value)
}
