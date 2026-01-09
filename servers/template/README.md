# MCP 服务模板

这是创建新 MCP 服务的模板。

## 目录结构

```
template/
├── README.md           # 服务文档
├── go.mod              # Go 模块定义（如果使用 Go）
├── server.go           # MCP 服务器主文件
├── config.yaml         # 配置文件（可选）
├── .env.example        # 环境变量示例
└── src/                # 源代码目录
    └── client.go       # API 客户端实现
```

## 创建新服务步骤

1. 复制此模板目录
2. 重命名为你的服务名（如 `your-service-mcp`）
3. 修改 `server.go` 实现你的服务逻辑
4. 更新 `README.md` 文档
5. 添加必要的配置文件和依赖

## 部署要求

所有服务必须：
- 通过 stdio 进行 JSON-RPC 通信
- 实现 MCP 协议标准
- 支持环境变量配置
- 可独立编译和运行

