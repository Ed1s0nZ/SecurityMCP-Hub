# ZoomEye MCP 服务

ZoomEye MCP 是一个基于 Model Context Protocol (MCP) 的 ZoomEye 资产搜索服务，使用 Go 语言实现。

## 功能特性

- ✅ **自主检索**：所有查询参数、翻页、返回数量等完全由大模型自主配置，无硬编码限制
- ✅ **灵活查询**：支持 ZoomEye 所有查询语法和参数
- ✅ **多种工具**：提供用户信息查询和资产搜索两种工具
- ✅ **独立部署**：可独立编译和运行，不依赖其他服务

## 工具说明

### 1. zoomeye_userinfo - 用户信息查询

获取 ZoomEye 用户信息，包括用户名、邮箱、订阅计划、积分等详细信息。

**参数说明：**
- 无需参数

**返回信息：**
- 用户基本信息（用户名、邮箱、电话、创建时间）
- 订阅信息（计划类型、结束日期、普通积分、权益积分）

**示例：**
```json
{
  "name": "zoomeye_userinfo",
  "arguments": {}
}
```

### 2. zoomeye_search - 资产搜索

在 ZoomEye 中搜索网络资产，支持自定义所有查询参数。

**参数说明：**
- `query` (必需): ZoomEye 查询语句，例如：`title="cisco vpn"` 或 `app="nginx" && country="CN"`。查询语句会自动进行 Base64 编码
- `page` (可选): 页码，从1开始，默认为1。支持任意页码翻页
- `pagesize` (可选): 每页返回数量，范围1-10000，默认为10。支持任意数量设置
- `fields` (可选): 返回字段，逗号分隔。默认：`ip,port,domain,update_time`
- `sub_type` (可选): 数据类型，支持 `v4`（IPv4）、`v6`（IPv6）和 `web`（Web资产），默认为 `v4`
- `facets` (可选): 统计项，如果有多个，用逗号分隔。支持：`country`, `subdivisions`, `city`, `product`, `service`, `device`, `os`, `port`
- `ignore_cache` (可选): 是否忽略缓存，默认为 false。支持商业版及以上用户

**支持的返回字段：**

**基础字段：**
- `ip`, `port`, `domain`, `url`, `hostname`, `os`, `service`, `title`, `version`, `device`, `rdns`, `product`, `banner`, `update_time`

**地理位置字段：**
- `continent.name`, `country.name`, `province.name`, `city.name`, `lon`, `lat`, `zipcode`

**网络信息字段：**
- `asn`, `protocol`, `isp.name`, `organization.name`

**SSL/TLS 字段：**
- `ssl`, `ssl.jarm`, `ssl.ja3s`

**HTTP 信息字段：**
- `header`, `header_hash`, `body`, `body_hash`, `header.server.name`, `header.server.version`

**其他字段：**
- `iconhash_md5`, `robots_md5`, `security_md5`, `idc`, `honeypot`, `primary_industry`, `sub_industry`, `rank`

**注意：** 字段权限取决于您的 ZoomEye 账号版本（免费版、专业版、商业版等），超出权限的字段将返回空值。

**示例：**
```json
{
  "name": "zoomeye_search",
  "arguments": {
    "query": "title=\"cisco vpn\"",
    "page": 1,
    "pagesize": 100,
    "fields": "ip,port,domain,title,country.name",
    "sub_type": "v4"
  }
}
```

**带统计的示例：**
```json
{
  "name": "zoomeye_search",
  "arguments": {
    "query": "app=\"nginx\"",
    "page": 1,
    "pagesize": 10,
    "fields": "ip,port,domain",
    "facets": "country,product,port"
  }
}
```

## 快速开始

### 1. 获取 ZoomEye API Key

1. 访问 [ZoomEye 官网](https://www.zoomeye.org)
2. 注册/登录账号
3. 在 [个人资料页面](https://www.zoomeye.org/profile) 下方找到 API-KEY

### 2. 配置环境变量

复制 `env.example` 为 `.env` 并填入您的 API Key：

```bash
cp env.example .env
# 编辑 .env 文件，填入您的 ZOOMEYE_API_KEY
```

或者直接设置环境变量：

```bash
export ZOOMEYE_API_KEY=your_api_key_here
```

### 3. 编译和运行

```bash
# 进入目录
cd servers/zoomeye-mcp

# 编译
go build -o zoomeye-mcp server.go

# 运行（通过 stdio 进行 JSON-RPC 通信）
./zoomeye-mcp
```

### 4. 在 MCP 客户端中配置

在您的 MCP 客户端配置文件中添加：

```json
{
  "mcpServers": {
    "zoomeye": {
      "command": "/path/to/zoomeye-mcp",
      "env": {
        "ZOOMEYE_API_KEY": "your_api_key_here"
      }
    }
  }
}
```

## 项目结构

```
zoomeye-mcp/
├── README.md           # 本文件
├── go.mod              # Go 模块定义
├── server.go           # MCP 服务器主文件
├── config.yaml         # 配置文件（可选）
├── env.example         # 环境变量示例
└── src/                # 源代码目录
    └── zoomeye_client.go  # ZoomEye API 客户端实现
```

## 开发说明

### 代码结构

- `server.go`: MCP 服务器主文件，实现 JSON-RPC over stdio 协议
- `src/zoomeye_client.go`: ZoomEye API 客户端，封装所有 API 调用

### 自主检索实现

所有查询参数都通过工具参数传入，没有任何硬编码限制：

- 查询语句：完全由用户/大模型构建，自动进行 Base64 编码
- 分页：支持任意页码，通过 `page` 参数控制
- 返回数量：支持 1-10000 的任意数量，通过 `pagesize` 参数控制
- 返回字段：支持任意字段组合，通过 `fields` 参数控制
- 数据类型：支持 v4、v6、web 三种类型
- 统计功能：支持通过 `facets` 参数进行统计

### Base64 编码说明

ZoomEye API 要求查询语句必须进行 Base64 编码。本工具会自动处理编码，用户只需提供原始查询语句即可。

例如：
- 输入查询：`title="cisco vpn"`
- 自动编码为：`dGl0bGU9ImNpc2NvIHZwbiIK`（Base64）

### 扩展开发

如需添加新功能：

1. 在 `src/zoomeye_client.go` 中添加新的 API 方法
2. 在 `server.go` 中注册新的工具
3. 实现工具处理函数

## API 参考

- [ZoomEye API 文档](https://www.zoomeye.org/doc)
- [Model Context Protocol 规范](https://modelcontextprotocol.io)

## 许可证

本项目采用 MIT 许可证。

