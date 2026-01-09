package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"zoomeye-mcp/src"
)

// MCP请求结构
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// MCP响应结构
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// 工具定义
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// 调用工具请求
type CallToolRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// 调用工具结果
type CallToolResult struct {
	Content []map[string]interface{} `json:"content"`
	IsError bool                     `json:"isError,omitempty"`
}

func main() {
	// 从环境变量获取 ZoomEye API Key
	apiKey := os.Getenv("ZOOMEYE_API_KEY")

	if apiKey == "" {
		log.Fatal("请设置环境变量 ZOOMEYE_API_KEY")
	}

	// 创建 ZoomEye 客户端
	zoomeyeClient := src.NewZoomEyeClient(apiKey)

	// 使用标准输入输出进行JSON-RPC通信
	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for scanner.Scan() {
		var request MCPRequest
		if err := json.Unmarshal(scanner.Bytes(), &request); err != nil {
			sendError(encoder, nil, -32700, "Parse error", err.Error())
			continue
		}

		var response MCPResponse
		response.JSONRPC = "2.0"
		response.ID = request.ID

		switch request.Method {
		case "initialize":
			response.Result = map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{},
				},
				"serverInfo": map[string]interface{}{
					"name":    "zoomeye-mcp",
					"version": "1.0.0",
				},
			}

		case "tools/list":
			response.Result = map[string]interface{}{
				"tools": []Tool{
					{
						Name: "zoomeye_userinfo",
						Description: `获取 ZoomEye 用户信息，包括用户名、邮箱、订阅计划、积分等详细信息。

返回信息包括：
- 用户基本信息（用户名、邮箱、电话、创建时间）
- 订阅信息（计划类型、结束日期、普通积分、权益积分）

可用于查询当前账号状态和可用积分。`,
						InputSchema: map[string]interface{}{
							"type":       "object",
							"properties": map[string]interface{}{},
						},
					},
					{
						Name: "zoomeye_search",
						Description: `在 ZoomEye 中搜索网络资产。支持自定义查询语句、分页、返回字段等所有参数。所有参数都可以由大模型自主配置。

支持的功能：
- 自定义查询语句（会自动进行 Base64 编码）
- 分页查询（支持任意页码）
- 自定义返回字段（支持所有 ZoomEye API 字段）
- 数据类型选择（v4、v6、web）
- 统计功能（facets）
- 缓存控制（ignore_cache）

支持的返回字段包括：
- 基础字段：ip, port, domain, url, hostname, os, service, title, version, device, rdns, product, banner, update_time
- 地理位置：continent.name, country.name, province.name, city.name, lon, lat, zipcode
- 网络信息：asn, protocol, isp.name, organization.name
- SSL/TLS：ssl, ssl.jarm, ssl.ja3s
- HTTP 信息：header, header_hash, body, body_hash, header.server.name, header.server.version
- 其他：iconhash_md5, robots_md5, security_md5, idc, honeypot, primary_industry, sub_industry, rank

字段权限取决于 ZoomEye 账号版本（免费版、专业版、商业版等）。`,
						InputSchema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"query": map[string]interface{}{
									"type":        "string",
									"description": "ZoomEye 查询语句，例如：title=\"cisco vpn\" 或 app=\"nginx\" && country=\"CN\"。查询语句会自动进行 Base64 编码，可以根据需要构建任意查询语句",
								},
								"page": map[string]interface{}{
									"type":        "integer",
									"description": "页码，从1开始，默认为1。可以根据需要设置任意页码进行翻页",
									"default":     1,
								},
								"pagesize": map[string]interface{}{
									"type":        "integer",
									"description": "每页返回数量，范围1-10000，默认为10。可以根据需要设置任意数量",
									"default":     10,
								},
								"fields": map[string]interface{}{
									"type": "string",
									"description": `返回字段，逗号分隔，例如：ip,port,domain,update_time。支持所有 ZoomEye API 字段，可根据需要选择任意字段组合。

常用字段：
- 基础：ip, port, domain, url, hostname, os, service, title, version, device, rdns, product, banner, update_time
- 地理位置：continent.name, country.name, province.name, city.name, lon, lat, zipcode
- 网络：asn, protocol, isp.name, organization.name
- SSL/TLS：ssl, ssl.jarm, ssl.ja3s
- HTTP：header, header_hash, body, body_hash, header.server.name, header.server.version
- 其他：iconhash_md5, robots_md5, security_md5, idc, honeypot, primary_industry, sub_industry, rank

字段权限取决于您的 ZoomEye 账号版本，超出权限的字段将返回空值。`,
									"default": "ip,port,domain,update_time",
								},
								"sub_type": map[string]interface{}{
									"type":        "string",
									"description": "数据类型，支持 v4（IPv4）、v6（IPv6）和 web（Web资产），默认为 v4",
									"enum":        []string{"v4", "v6", "web"},
									"default":     "v4",
								},
								"facets": map[string]interface{}{
									"type":        "string",
									"description": "统计项，如果有多个，用逗号分隔。支持：country, subdivisions, city, product, service, device, os, port。例如：country,product,port",
								},
								"ignore_cache": map[string]interface{}{
									"type":        "boolean",
									"description": "是否忽略缓存，默认为 false。支持商业版及以上用户",
									"default":     false,
								},
							},
							"required": []string{"query"},
						},
					},
				},
			}

		case "tools/call":
			var callRequest CallToolRequest
			if err := json.Unmarshal(request.Params, &callRequest); err != nil {
				sendError(encoder, request.ID, -32602, "Invalid params", err.Error())
				continue
			}

			var result CallToolResult
			var err error

			switch callRequest.Name {
			case "zoomeye_userinfo":
				result, err = handleZoomEyeUserInfo(zoomeyeClient)
			case "zoomeye_search":
				result, err = handleZoomEyeSearch(zoomeyeClient, callRequest.Arguments)
			default:
				sendError(encoder, request.ID, -32601, "Method not found", fmt.Sprintf("Unknown tool: %s", callRequest.Name))
				continue
			}

			if err != nil {
				result = CallToolResult{
					Content: []map[string]interface{}{
						{
							"type": "text",
							"text": fmt.Sprintf("错误: %v", err),
						},
					},
					IsError: true,
				}
			}

			response.Result = result

		default:
			sendError(encoder, request.ID, -32601, "Method not found", fmt.Sprintf("Unknown method: %s", request.Method))
			continue
		}

		if err := encoder.Encode(response); err != nil {
			log.Printf("编码响应失败: %v", err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func sendError(encoder *json.Encoder, id interface{}, code int, message, data string) {
	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
	}
	if data != "" {
		response.Error.Message = fmt.Sprintf("%s: %s", message, data)
	}
	encoder.Encode(response)
}

func handleZoomEyeUserInfo(client *src.ZoomEyeClient) (CallToolResult, error) {
	result, err := client.GetUserInfo()
	if err != nil {
		return CallToolResult{}, err
	}

	response := map[string]interface{}{
		"success": true,
		"code":    result.Code,
		"message": result.Message,
		"data": map[string]interface{}{
			"username":   result.Data.Username,
			"email":      result.Data.Email,
			"phone":      result.Data.Phone,
			"created_at": result.Data.CreatedAt,
			"subscription": map[string]interface{}{
				"plan":          result.Data.Subscription.Plan,
				"end_date":      result.Data.Subscription.EndDate,
				"points":        result.Data.Subscription.Points,
				"zoomeye_points": result.Data.Subscription.ZoomEyePoints,
			},
		},
	}

	responseJSON, _ := json.MarshalIndent(response, "", "  ")

	return CallToolResult{
		Content: []map[string]interface{}{
			{
				"type": "text",
				"text": string(responseJSON),
			},
		},
	}, nil
}

func handleZoomEyeSearch(client *src.ZoomEyeClient, args map[string]interface{}) (CallToolResult, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return CallToolResult{}, fmt.Errorf("query参数是必需的")
	}

	// 对查询语句进行 Base64 编码
	qbase64 := base64.StdEncoding.EncodeToString([]byte(query))

	params := src.SearchParams{
		QBase64:     qbase64,
		Page:        1,
		PageSize:    10,
		SubType:     "v4",
		Fields:      "ip,port,domain,update_time",
		IgnoreCache: false,
	}

	// 解析可选参数
	if page, ok := args["page"].(float64); ok {
		params.Page = int(page)
	}
	if pagesize, ok := args["pagesize"].(float64); ok {
		params.PageSize = int(pagesize)
	}
	if fields, ok := args["fields"].(string); ok && fields != "" {
		params.Fields = fields
	}
	if subType, ok := args["sub_type"].(string); ok && subType != "" {
		params.SubType = subType
	}
	if facets, ok := args["facets"].(string); ok && facets != "" {
		params.Facets = facets
	}
	if ignoreCache, ok := args["ignore_cache"].(bool); ok {
		params.IgnoreCache = ignoreCache
	}

	result, err := client.Search(params)
	if err != nil {
		return CallToolResult{}, err
	}

	response := map[string]interface{}{
		"success": true,
		"code":    result.Code,
		"message": result.Message,
		"total":   result.Total,
		"query":   result.Query,
		"count":   len(result.Data),
		"data":    result.Data,
	}

	responseJSON, _ := json.MarshalIndent(response, "", "  ")

	return CallToolResult{
		Content: []map[string]interface{}{
			{
				"type": "text",
				"text": string(responseJSON),
			},
		},
	}, nil
}

