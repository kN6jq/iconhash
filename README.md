# Favicon Hash 计算器

一个简单的工具，用于计算 FOFA 和 Hunter 搜索引擎的 favicon 哈希值。

## 功能特点

- 计算用于 FOFA 搜索的 Murmur3 哈希值
- 计算用于 Hunter 搜索的 MD5 哈希值
- 支持终端和网页界面两种模式
- 内置缓存，避免重复下载

## 使用方法

### 终端模式

```bash
# 基本用法
./favicon-hash http://example.com/favicon.ico

# 显式指定模式
./favicon-hash -mode=terminal http://example.com/favicon.ico
```

### 网页模式

```bash
# 在默认端口 8080 启动网页服务器
./favicon-hash -mode=web

# 在自定义端口启动网页服务器
./favicon-hash -mode=web -port=3000
```

然后打开浏览器，访问 `http://localhost:8080`（或您的自定义端口）。

## 构建

```bash
go build -o favicon-hash main.go
```

## 依赖项

- github.com/hashicorp/golang-lru/simplelru
- github.com/twmb/murmur3

## 许可证

[MIT 许可证](LICENSE)

