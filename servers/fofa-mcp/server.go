package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"fofa-mcp/src"
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
	// 从环境变量获取FOFA凭证
	email := os.Getenv("FOFA_EMAIL")
	key := os.Getenv("FOFA_KEY")

	if email == "" || key == "" {
		log.Fatal("请设置环境变量 FOFA_EMAIL 和 FOFA_KEY")
	}

	// 创建FOFA客户端
	fofaClient := src.NewFofaClient(email, key)

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
					"name":    "fofa-mcp",
					"version": "1.0.0",
				},
			}

		case "tools/list":
			response.Result = map[string]interface{}{
				"tools": []Tool{
					{
						Name: "fofa_search",
						Description: `在FOFA中搜索资产。支持自定义查询语句、分页、返回字段等所有参数。所有参数都可以由大模型自主配置，包括查询语句、页码、每页数量、返回字段等。

支持50个返回字段，包括基础字段（ip,port,host等）、地理位置字段（country,region,city等）、证书字段（cert.*）、协议字段（banner,protocol等）、产品字段（product,product.version等）等。字段权限取决于FOFA账号版本。

重要限制：当fields参数包含cert或banner字段时，size参数最大值自动限制为2000（而非10000）。`,
						InputSchema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"query": map[string]interface{}{
									"type":        "string",
									"description": "FOFA查询语句，例如：app=\"Apache\" && country=\"CN\"。可以根据需要构建任意查询语句",
								},
								"page": map[string]interface{}{
									"type":        "integer",
									"description": "页码，从1开始，默认为1。可以根据需要设置任意页码进行翻页",
									"default":     1,
								},
								"size": map[string]interface{}{
									"type":        "integer",
									"description": "每页返回数量，范围1-10000，默认为100。可以根据需要设置任意数量。重要限制：当fields参数包含cert或banner字段时，size最大值限制为2000",
									"default":     100,
								},
								"fields": map[string]interface{}{
									"type": "string",
									"description": `返回字段，逗号分隔，例如：host,ip,port,protocol,title。支持所有FOFA API字段，可根据需要选择任意字段组合。

完整字段列表（共50个字段）：
【无权限字段（1-33）】：ip,port,protocol,country,country_name,region,city,longitude,latitude,asn,org,host,domain,os,server,icp,title,jarm,header,banner,cert,base_protocol,link,cert.issuer.org,cert.issuer.cn,cert.subject.org,cert.subject.cn,tls.ja3s,tls.version,cert.sn,cert.not_before,cert.not_after,cert.domain
【个人版及以上（34-36）】：header_hash,banner_hash,banner_fid
【专业版及以上（37-40）】：cname,lastupdatetime,product,product_category
【商业版本及以上（41-47）】：product.version,icon_hash,cert.is_valid,cname_domain,body,cert.is_match,cert.is_equal
【企业会员（48-50）】：icon,fid,structinfo

重要提示：
- 当查询包含cert或banner字段时，size参数值最大为2000
- 字段权限取决于您的FOFA账号版本，超出权限的字段将返回空值
- 可以根据实际需求灵活组合任意字段`,
									"default": "host,ip,port,protocol",
								},
								"full": map[string]interface{}{
									"type":        "boolean",
									"description": "是否返回全量数据，默认为false",
									"default":     false,
								},
								"is_domain": map[string]interface{}{
									"type":        "boolean",
									"description": "是否为域名查询，默认为false",
									"default":     false,
								},
							},
							"required": []string{"query"},
						},
					},
					{
						Name:        "fofa_stats",
						Description: "获取FOFA查询结果的统计信息。支持自定义查询语句和统计字段。",
						InputSchema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"query": map[string]interface{}{
									"type":        "string",
									"description": "FOFA查询语句",
								},
								"fields": map[string]interface{}{
									"type":        "string",
									"description": "要统计的字段，逗号分隔，例如：country,server,protocol。可以根据需要选择任意字段进行统计",
								},
							},
							"required": []string{"query"},
						},
					},
					{
						Name:        "fofa_host_info",
						Description: "获取指定主机的详细信息，包括IP、ASN、组织、国家、协议等。",
						InputSchema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"host": map[string]interface{}{
									"type":        "string",
									"description": "主机地址，可以是IP或域名",
								},
							},
							"required": []string{"host"},
						},
					},
				},
			}

		case "tools/call":
			var callRequest struct {
				Name      string                 `json:"name"`
				Arguments map[string]interface{} `json:"arguments"`
			}
			if err := json.Unmarshal(request.Params, &callRequest); err != nil {
				sendError(encoder, request.ID, -32602, "Invalid params", err.Error())
				continue
			}

			var result CallToolResult
			var err error

			switch callRequest.Name {
			case "fofa_search":
				result, err = handleFofaSearch(fofaClient, callRequest.Arguments)
			case "fofa_stats":
				result, err = handleFofaStats(fofaClient, callRequest.Arguments)
			case "fofa_host_info":
				result, err = handleFofaHostInfo(fofaClient, callRequest.Arguments)
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

func handleFofaSearch(client *src.FofaClient, args map[string]interface{}) (CallToolResult, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return CallToolResult{}, fmt.Errorf("query参数是必需的")
	}

	params := src.QueryParams{
		Query:    query,
		Page:     1,
		Size:     100,
		Fields:   "host,ip,port,protocol",
		Full:     false,
		IsDomain: false,
	}

	// 解析可选参数
	if page, ok := args["page"].(float64); ok {
		params.Page = int(page)
	}
	if size, ok := args["size"].(float64); ok {
		params.Size = int(size)
	}
	if fields, ok := args["fields"].(string); ok && fields != "" {
		params.Fields = fields
	}
	if full, ok := args["full"].(bool); ok {
		params.Full = full
	}
	if isDomain, ok := args["is_domain"].(bool); ok {
		params.IsDomain = isDomain
	}

	result, err := client.Search(params)
	if err != nil {
		return CallToolResult{}, err
	}

	response := map[string]interface{}{
		"success": true,
		"query":   result.Query,
		"page":    result.Page,
		"size":    result.Size,
		"mode":    result.Mode,
		"total":   len(result.Results),
		"results": result.Results,
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

func handleFofaStats(client *src.FofaClient, args map[string]interface{}) (CallToolResult, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return CallToolResult{}, fmt.Errorf("query参数是必需的")
	}

	fields := ""
	if f, ok := args["fields"].(string); ok {
		fields = f
	}

	result, err := client.Stats(query, fields)
	if err != nil {
		return CallToolResult{}, err
	}

	response := map[string]interface{}{
		"success":  true,
		"distinct": result.Distinct,
		"aggs":     result.Aggs,
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

func handleFofaHostInfo(client *src.FofaClient, args map[string]interface{}) (CallToolResult, error) {
	host, ok := args["host"].(string)
	if !ok || host == "" {
		return CallToolResult{}, fmt.Errorf("host参数是必需的")
	}

	result, err := client.GetHostInfo(host)
	if err != nil {
		return CallToolResult{}, err
	}

	// 直接返回所有字段，不写死任何字段
	response := map[string]interface{}{
		"success": true,
	}
	// 将 API 返回的所有字段都包含进来
	if result != nil {
		for k, v := range *result {
			// 跳过 error 字段（已经处理过）
			if k != "error" && k != "errmsg" {
				response[k] = v
			}
		}
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
