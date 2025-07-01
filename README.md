# IconHash

一个简单高效的工具，用于计算网站图标(favicon)的哈希值，支持 FOFA 和 Hunter 等搜索引擎的哈希格式。

## 功能特点

- 计算用于 FOFA 搜索的 Murmur3 哈希值
- 计算用于 Hunter 搜索的 MD5 哈希值
- 支持终端和网页两种使用模式
- 内置缓存，避免重复下载
- 支持跳过 TLS 证书验证（用于处理自签名证书或 IP 访问）
- 资源文件嵌入到可执行程序中，无需外部依赖

## 项目结构

```
/
├── cmd/
│   └── iconhash/         # 入口点
│       ├── main.go       # 主程序
│       ├── assets.go     # 嵌入式资源管理
│       └── assets/       # 嵌入式静态资源
├── internal/
│   ├── hash/             # 哈希计算核心逻辑
│   └── web/              # Web服务相关
├── pkg/
│   └── utils/            # 通用工具函数
├── go.mod
└── README.md
```

## 使用方法

### 终端模式

```bash
# 基本用法
./iconhash http://example.com/favicon.ico

# 显式指定模式
./iconhash -mode=terminal http://example.com/favicon.ico

# 跳过 TLS 证书验证（用于 HTTPS 的 IP 地址访问）
./iconhash -insecure=true https://192.168.1.1/favicon.ico
```

### 网页模式

```bash
# 在默认端口 8080 启动网页服务器
./iconhash -mode=web

# 在自定义端口启动网页服务器
./iconhash -mode=web -port=3000

# 启用不安全模式（跳过 TLS 证书验证）
./iconhash -mode=web -insecure=true
```

然后打开浏览器，访问 `http://localhost:8080`（或您的自定义端口）。

## 构建

```bash
# 从源代码构建
go build -o iconhash ./cmd/iconhash
```

## 依赖项

- github.com/hashicorp/golang-lru/simplelru - LRU 缓存实现
- github.com/twmb/murmur3 - Murmur3 哈希算法实现

## 许可证

MIT 许可证

