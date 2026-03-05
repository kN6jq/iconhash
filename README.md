# IconHash

IconHash 是一个用于计算网站图标（favicon）哈希值的工具，支持 FOFA 和 Hunter 等安全搜索引擎的哈希格式。

## 功能特点

- 自动寻找网站图标（无需提供完整的 favicon URL）
- 支持 JavaScript 重定向检测
- 支持从文件批量处理 URL
- 支持并发处理，提高效率
- 简化输出格式，便于分析和处理

## 安装

### 从源码编译

```bash
git clone https://github.com/kN6jq/iconhash.git
cd iconhash
go build -o iconhash.exe cmd/iconhash/main.go
```

### 下载预编译版本

从 [Releases](https://github.com/kN6jq/iconhash/releases) 页面下载最新版本。

## 使用方法

```
用法: iconhash [选项]

选项:
  -u=<url>   指定favicon的URL地址
  -f=<file>  指定本地favicon图片路径

示例:
  iconhash -u=http://example.com/favicon.ico
  iconhash -f=/path/to/favicon.ico
```

## 输出格式

```
来源: http://example.com/favicon.ico
FOFA Hash: -123456789
Hunter Hash: d41d8cd98f00b204e9800998ecf8427e
MD5: d41d8cd98f00b204e9800998ecf8427e
```

## 注意事项

- 默认情况下，工具会自动跳过 TLS 证书验证
- `-u` 和 `-f` 参数不能同时使用

## 许可证

MIT License

