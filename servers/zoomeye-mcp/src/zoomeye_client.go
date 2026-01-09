package src

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ZoomEye API 客户端
type ZoomEyeClient struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

// 用户信息响应
type UserInfoResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Username     string `json:"username"`
		Email        string `json:"email"`
		Phone        string `json:"phone"`
		CreatedAt    string `json:"created_at"`
		Subscription struct {
			Plan          string `json:"plan"`
			EndDate       string `json:"end_date"`
			Points        string `json:"points"`
			ZoomEyePoints string `json:"zoomeye_points"`
		} `json:"subscription"`
	} `json:"data"`
}

// 搜索请求参数
type SearchParams struct {
	QBase64     string `json:"qbase64"`      // Base64 编码的查询语句
	Fields      string `json:"fields"`       // 返回字段，逗号分隔
	SubType     string `json:"sub_type"`     // 数据类型：v4, v6, web
	Page        int    `json:"page"`         // 页码，从1开始
	PageSize    int    `json:"pagesize"`     // 每页数量，最大10000
	Facets      string `json:"facets"`       // 统计项，逗号分隔
	IgnoreCache bool   `json:"ignore_cache"` // 是否忽略缓存
}

// 搜索结果响应
type SearchResponse struct {
	Code    int                      `json:"code"`
	Message string                   `json:"message"`
	Total   int                      `json:"total"`
	Query   string                   `json:"query"`
	Data    []map[string]interface{} `json:"data"`
}

// 创建新的 ZoomEye 客户端
func NewZoomEyeClient(apiKey string) *ZoomEyeClient {
	return &ZoomEyeClient{
		APIKey:  apiKey,
		BaseURL: "https://api.zoomeye.org",
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// 对查询语句进行 Base64 编码
func (c *ZoomEyeClient) EncodeQuery(query string) string {
	return base64.StdEncoding.EncodeToString([]byte(query))
}

// 获取用户信息
func (c *ZoomEyeClient) GetUserInfo() (*UserInfoResponse, error) {
	apiURL := fmt.Sprintf("%s/v2/userinfo", c.BaseURL)

	req, err := http.NewRequest("POST", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("API-KEY", c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "zoomeye-mcp/1.0")

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

	var userInfoResp UserInfoResponse
	if err := json.Unmarshal(body, &userInfoResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if userInfoResp.Code != 60000 {
		return nil, fmt.Errorf("ZoomEye API错误: %s (code: %d)", userInfoResp.Message, userInfoResp.Code)
	}

	return &userInfoResp, nil
}

// 执行资产搜索
func (c *ZoomEyeClient) Search(params SearchParams) (*SearchResponse, error) {
	// 参数验证和默认值
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 10
	}
	if params.PageSize > 10000 {
		params.PageSize = 10000
	}
	if params.SubType == "" {
		params.SubType = "v4"
	}
	if params.Fields == "" {
		params.Fields = "ip,port,domain,update_time"
	}

	apiURL := fmt.Sprintf("%s/v2/search", c.BaseURL)

	// 构建请求体
	requestBody := map[string]interface{}{
		"qbase64": params.QBase64,
		"page":    params.Page,
	}

	if params.Fields != "" {
		requestBody["fields"] = params.Fields
	}
	if params.SubType != "" {
		requestBody["sub_type"] = params.SubType
	}
	if params.PageSize > 0 {
		requestBody["pagesize"] = params.PageSize
	}
	if params.Facets != "" {
		requestBody["facets"] = params.Facets
	}
	if params.IgnoreCache {
		requestBody["ignore_cache"] = true
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("构建请求体失败: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("API-KEY", c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "zoomeye-mcp/1.0")

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

	var searchResp SearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if searchResp.Code != 60000 {
		return nil, fmt.Errorf("ZoomEye API错误: %s (code: %d)", searchResp.Message, searchResp.Code)
	}

	return &searchResp, nil
}
