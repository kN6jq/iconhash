# IconHash

IconHash 是一个用于计算网站图标（favicon）哈希值的工具，支持 FOFA 和 Hunter 等安全搜索引擎的哈希格式。

## 功能特点

- 自动发现网站图标（传入网站地址即可，无需手动指定 favicon 路径）
- 支持直接指定 favicon URL
- 支持本地 favicon 图片文件
- 支持 FOFA、Hunter、MD5 三种哈希格式
- 支持 JavaScript 重定向检测

## 安装

### 从源码编译

```bash
git clone https://github.com/kN6jq/iconhash.git
cd iconhash
go build -o iconhash.exe ./cmd/iconhash
```

### 下载预编译版本

从 [Releases](https://github.com/kN6jq/iconhash/releases) 页面下载最新版本。

## 使用方法

```
用法: iconhash [选项]

选项:
  -u=<url>   指定URL地址，支持网站地址或favicon直链
  -f=<file>  指定本地favicon图片路径

示例:
  iconhash -u=http://example.com              自动发现favicon
  iconhash -u=http://example.com/favicon.ico  直接获取favicon
  iconhash -f=/path/to/favicon.ico            本地图片
```

## 输出格式

```
来源: http://example.com
Favicon: http://example.com/favicon.ico
FOFA Hash: -123456789
Hunter Hash: d41d8cd98f00b204e9800998ecf8427e
MD5: d41d8cd98f00b204e9800998ecf8427e
```

## 注意事项

- 默认情况下，工具会自动跳过 TLS 证书验证
- `-u` 和 `-f` 参数不能同时使用

## 许可证

MIT License

