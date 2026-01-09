package src

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// FOFA API 客户端
type FofaClient struct {
	Email   string
	Key     string
	BaseURL string
	Client  *http.Client
}

// 查询参数结构
type QueryParams struct {
	Query    string `json:"query"`     // 查询语句
	Page     int    `json:"page"`      // 页码，从1开始
	Size     int    `json:"size"`      // 每页数量，最大10000
	Fields   string `json:"fields"`    // 返回字段，逗号分隔，如：host,ip,port,protocol
	Full     bool   `json:"full"`      // 是否全量数据
	IsDomain bool   `json:"is_domain"` // 是否为域名查询
}

// 搜索结果响应
type SearchResponse struct {
	Error   bool       `json:"error"`
	ErrMsg  string     `json:"errmsg,omitempty"`
	Size    int        `json:"size"`
	Page    int        `json:"page"`
	Mode    string     `json:"mode"`
	Query   string     `json:"query"`
	Results [][]string `json:"results"`
}

// 统计响应
type StatsResponse struct {
	Error    bool                   `json:"error"`
	ErrMsg   string                 `json:"errmsg,omitempty"`
	Distinct map[string]int         `json:"distinct"`
	Aggs     map[string]interface{} `json:"aggs"`
}

// 主机信息响应 - 使用 map 动态处理所有返回字段，不写死
type HostInfoResponse map[string]interface{}

// 创建新的FOFA客户端
func NewFofaClient(email, key string) *FofaClient {
	return &FofaClient{
		Email:   email,
		Key:     key,
		BaseURL: "https://fofa.info",
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// 对查询语句进行Base64编码
func (c *FofaClient) encodeQuery(query string) string {
	return base64.StdEncoding.EncodeToString([]byte(query))
}

// 执行搜索查询
func (c *FofaClient) Search(params QueryParams) (*SearchResponse, error) {
	// 参数验证和默认值
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Size < 1 {
		params.Size = 100
	}

	// 检查是否包含cert或banner字段，如果包含则限制size最大为2000
	hasCertOrBanner := false
	if params.Fields != "" {
		fieldsLower := strings.ToLower(params.Fields)
		hasCertOrBanner = strings.Contains(fieldsLower, "cert") || strings.Contains(fieldsLower, "banner")
	}

	if hasCertOrBanner {
		if params.Size > 2000 {
			params.Size = 2000
		}
	} else {
		if params.Size > 10000 {
			params.Size = 10000
		}
	}

	if params.Fields == "" {
		params.Fields = "host,ip,port,protocol"
	}

	// 构建URL
	apiURL := fmt.Sprintf("%s/api/v1/search/all", c.BaseURL)

	// 对查询语句进行Base64编码
	queryBase64 := c.encodeQuery(params.Query)

	// 构建查询参数
	queryValues := url.Values{}
	queryValues.Set("email", c.Email)
	queryValues.Set("key", c.Key)
	queryValues.Set("qbase64", queryBase64)
	queryValues.Set("page", strconv.Itoa(params.Page))
	queryValues.Set("size", strconv.Itoa(params.Size))
	queryValues.Set("fields", params.Fields)
	if params.Full {
		queryValues.Set("full", "true")
	}
	if params.IsDomain {
		queryValues.Set("is_domain", "true")
	}

	fullURL := fmt.Sprintf("%s?%s", apiURL, queryValues.Encode())

	// 发送请求
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", "fofa-mcp/1.0")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var searchResp SearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if searchResp.Error {
		return nil, fmt.Errorf("FOFA API错误: %s", searchResp.ErrMsg)
	}

	return &searchResp, nil
}

// 获取统计信息
func (c *FofaClient) Stats(query string, fields string) (*StatsResponse, error) {
	apiURL := fmt.Sprintf("%s/api/v1/search/stats", c.BaseURL)

	queryBase64 := c.encodeQuery(query)

	queryValues := url.Values{}
	queryValues.Set("email", c.Email)
	queryValues.Set("key", c.Key)
	queryValues.Set("qbase64", queryBase64)
	if fields != "" {
		queryValues.Set("fields", fields)
	}

	fullURL := fmt.Sprintf("%s?%s", apiURL, queryValues.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", "fofa-mcp/1.0")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var statsResp StatsResponse
	if err := json.Unmarshal(body, &statsResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if statsResp.Error {
		return nil, fmt.Errorf("FOFA API错误: %s", statsResp.ErrMsg)
	}

	return &statsResp, nil
}

// 获取主机信息
func (c *FofaClient) GetHostInfo(host string) (*HostInfoResponse, error) {
	apiURL := fmt.Sprintf("%s/api/v1/host/%s", c.BaseURL, url.QueryEscape(host))

	queryValues := url.Values{}
	queryValues.Set("email", c.Email)
	queryValues.Set("key", c.Key)

	fullURL := fmt.Sprintf("%s?%s", apiURL, queryValues.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", "fofa-mcp/1.0")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var hostResp HostInfoResponse
	if err := json.Unmarshal(body, &hostResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查是否有错误
	if errorVal, ok := hostResp["error"].(bool); ok && errorVal {
		errMsg := ""
		if msg, ok := hostResp["errmsg"].(string); ok {
			errMsg = msg
		}
		return nil, fmt.Errorf("FOFA API错误: %s", errMsg)
	}

	return &hostResp, nil
}
