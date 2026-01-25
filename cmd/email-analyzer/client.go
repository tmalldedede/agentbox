package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Client AgentBox API 客户端
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient 创建客户端
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// setAuth 设置认证头
func (c *Client) setAuth(req *http.Request) {
	if c.apiKey == "" {
		return
	}
	// 支持 API Key (ab_xxx) 和 JWT Token
	if strings.HasPrefix(c.apiKey, "ab_") {
		req.Header.Set("X-API-Key", c.apiKey)
	} else {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
}

// Health 健康检查
func (c *Client) Health() (*HealthResponse, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/health", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var apiResp struct {
		Code int             `json:"code"`
		Data HealthResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	return &apiResp.Data, nil
}

// Login 登录获取 Token
func (c *Client) Login(username, password string) (string, error) {
	body, _ := json.Marshal(LoginRequest{
		Username: username,
		Password: password,
	})

	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/auth/login", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("连接失败: %w", err)
	}
	defer resp.Body.Close()

	var apiResp struct {
		Code    int           `json:"code"`
		Message string        `json:"message"`
		Data    LoginResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", err
	}

	if apiResp.Code != 0 {
		return "", fmt.Errorf("%s", apiResp.Message)
	}

	return apiResp.Data.Token, nil
}

// UploadFile 上传文件
func (c *Client) UploadFile(filePath string) (*UploadedFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 创建 multipart 请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, err
	}
	writer.Close()

	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/files", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("上传失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var apiResp struct {
		Code int          `json:"code"`
		Data UploadedFile `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	return &apiResp.Data, nil
}

// CreateTask 创建任务
func (c *Client) CreateTask(req *CreateTaskRequest) (*Task, error) {
	body, _ := json.Marshal(req)

	httpReq, err := http.NewRequest("POST", c.baseURL+"/api/v1/tasks", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	c.setAuth(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("创建任务失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var apiResp struct {
		Code int  `json:"code"`
		Data Task `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	return &apiResp.Data, nil
}

// GetTask 获取任务
func (c *Client) GetTask(id string) (*Task, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/tasks/"+id, nil)
	if err != nil {
		return nil, err
	}
	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var apiResp struct {
		Code int  `json:"code"`
		Data Task `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	return &apiResp.Data, nil
}

// WaitTask 等待任务完成
func (c *Client) WaitTask(id string, timeout time.Duration) (*Task, error) {
	deadline := time.Now().Add(timeout)
	pollInterval := 500 * time.Millisecond

	for time.Now().Before(deadline) {
		task, err := c.GetTask(id)
		if err != nil {
			return nil, err
		}

		// 检查是否完成
		switch task.Status {
		case "completed", "failed", "cancelled":
			return task, nil
		}

		time.Sleep(pollInterval)
	}

	return nil, fmt.Errorf("任务超时: %s", id)
}

// EventHandler SSE 事件处理函数
type EventHandler func(event SSEEvent) bool // 返回 false 停止接收

// StreamTaskEvents 流式接收任务事件
func (c *Client) StreamTaskEvents(id string, handler EventHandler) error {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/tasks/"+id+"/events", nil)
	if err != nil {
		return err
	}
	c.setAuth(req)
	req.Header.Set("Accept", "text/event-stream")

	// 使用不超时的 HTTP 客户端
	client := &http.Client{
		Timeout: 0, // 无超时
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("SSE 连接失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	reader := bufio.NewReader(resp.Body)
	var eventType string
	var data string

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		line = strings.TrimSpace(line)

		if line == "" {
			// 空行表示事件结束
			if eventType != "" && data != "" {
				var eventData map[string]interface{}
				if err := json.Unmarshal([]byte(data), &eventData); err == nil {
					event := SSEEvent{
						Type: eventType,
						Data: eventData,
					}
					if !handler(event) {
						return nil
					}
				}
			}
			eventType = ""
			data = ""
			continue
		}

		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			data = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		}
	}
}

// ListTasks 列出任务
func (c *Client) ListTasks(limit int) ([]Task, error) {
	url := fmt.Sprintf("%s/api/v1/tasks?limit=%d", c.baseURL, limit)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var apiResp struct {
		Code int `json:"code"`
		Data struct {
			Tasks []Task `json:"tasks"`
			Total int    `json:"total"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	return apiResp.Data.Tasks, nil
}
