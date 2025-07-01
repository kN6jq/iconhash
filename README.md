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
用法: iconhash [选项] <URL>

选项:
  -u=<url>             指定单个URL
  -file=<filepath>      包含URL列表的文件路径
  -output=<filepath>    输出结果的文件路径
  -threads=<number>     并发线程数 (默认: 10)

示例:
  iconhash http://example.com
  iconhash -u=http://example.com
  iconhash -file=urls.txt -output=results.txt
```

## 输出格式

输出格式为简化的一行文本，包含以下信息，以竖线分隔：

```
URL | 图标URL | FOFA哈希值 | Hunter哈希值
```

例如：

```
http://example.com | http://example.com/favicon.ico | d41d8cd98f00b204e9800998ecf8427e | f2769a8bb67b9348a93a1a8fb2ea3f7e
```

## 批量处理

创建一个包含多个URL的文本文件（每行一个URL），然后使用 `-file` 参数：

```bash
iconhash -file=urls.txt -output=results.txt
```

## 注意事项

- 默认情况下，工具会自动跳过 TLS 证书验证
- 处理大量URL时，可以使用 `-threads` 参数调整并发数
- 输出结果可以使用 `-output` 参数保存到文件

## 许可证

MIT License

