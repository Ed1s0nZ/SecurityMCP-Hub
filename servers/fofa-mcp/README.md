# FOFA MCP 服务

FOFA MCP 是一个基于 Model Context Protocol (MCP) 的 FOFA 资产搜索服务，使用 Go 语言实现。

## 功能特性

- ✅ **自主检索**：所有查询参数、翻页、返回数量等完全由大模型自主配置，无硬编码限制
- ✅ **灵活查询**：支持 FOFA 所有查询语法和参数
- ✅ **多种工具**：提供搜索、统计、主机信息三种工具
- ✅ **独立部署**：可独立编译和运行，不依赖其他服务

## 工具说明

### 1. fofa_search - 资产搜索

在 FOFA 中搜索资产，支持自定义所有查询参数。

**参数说明：**
- `query` (必需): FOFA 查询语句，例如：`app="Apache" && country="CN"`
- `page` (可选): 页码，从1开始，默认为1。支持任意页码翻页
- `size` (可选): 每页返回数量，范围1-10000，默认为100。支持任意数量设置
- `fields` (可选): 返回字段，逗号分隔。可选字段：host,title,ip,domain,port,protocol,server,country,region,city,icp,asn,org,header,body,banner,cert
- `full` (可选): 是否返回全量数据，默认为false
- `is_domain` (可选): 是否为域名查询，默认为false

**示例：**
```json
{
  "query": "app=\"nginx\" && country=\"CN\"",
  "page": 1,
  "size": 100,
  "fields": "host,ip,port,protocol,title"
}
```

### 2. fofa_stats - 统计信息

获取 FOFA 查询结果的统计信息。

**参数说明：**
- `query` (必需): FOFA 查询语句
- `fields` (可选): 要统计的字段，逗号分隔，例如：country,server,protocol

**示例：**
```json
{
  "query": "app=\"nginx\"",
  "fields": "country,server"
}
```

### 3. fofa_host_info - 主机信息

获取指定主机的详细信息。

**参数说明：**
- `host` (必需): 主机地址，可以是IP或域名

**示例：**
```json
{
  "host": "192.168.1.1"
}
```

## 快速开始

### 1. 获取 FOFA API 凭证

1. 访问 [FOFA 官网](https://fofa.info)
2. 注册/登录账号
3. 在 [个人中心](https://fofa.info/userInfo) 获取您的邮箱和 API Key

### 2. 配置环境变量

复制 `.env.example` 为 `.env` 并填入您的凭证：

```bash
cp .env.example .env
# 编辑 .env 文件，填入您的 FOFA_EMAIL 和 FOFA_KEY
```

或者直接设置环境变量：

```bash
export FOFA_EMAIL=your_email@example.com
export FOFA_KEY=your_api_key_here
```

### 3. 编译和运行

```bash
# 进入目录
cd servers/fofa-mcp

# 编译
go build -o fofa-mcp server.go

# 运行（通过 stdio 进行 JSON-RPC 通信）
./fofa-mcp
```

### 4. 在 MCP 客户端中配置

在您的 MCP 客户端配置文件中添加：

```json
{
  "mcpServers": {
    "fofa": {
      "command": "/path/to/fofa-mcp",
      "env": {
        "FOFA_EMAIL": "your_email@example.com",
        "FOFA_KEY": "your_api_key_here"
      }
    }
  }
}
```

## 项目结构

```
fofa-mcp/
├── README.md           # 本文件
├── go.mod              # Go 模块定义
├── server.go           # MCP 服务器主文件
├── config.yaml         # 配置文件（可选）
├── .env.example        # 环境变量示例
└── src/                # 源代码目录
    └── fofa_client.go  # FOFA API 客户端实现
```

## 开发说明

### 代码结构

- `server.go`: MCP 服务器主文件，实现 JSON-RPC over stdio 协议
- `src/fofa_client.go`: FOFA API 客户端，封装所有 API 调用

### 自主检索实现

所有查询参数都通过工具参数传入，没有任何硬编码限制：

- 查询语句：完全由用户/大模型构建
- 分页：支持任意页码，通过 `page` 参数控制
- 返回数量：支持 1-10000 的任意数量，通过 `size` 参数控制
- 返回字段：支持任意字段组合，通过 `fields` 参数控制

### 扩展开发

如需添加新功能：

1. 在 `src/fofa_client.go` 中添加新的 API 方法
2. 在 `server.go` 中注册新的工具
3. 实现工具处理函数

## 参考文档

- [FOFA API 文档](https://fofa.info/api)
- [Model Context Protocol 规范](https://modelcontextprotocol.io)

## 许可证

本项目采用 MIT 许可证。

