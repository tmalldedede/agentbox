# Manus API 参考文档存档

> 存档时间: 2025-01-19
> 来源: https://open.manus.im/docs/api-reference

## 概述

**Base URL:** `https://api.manus.ai`
**认证方式:** API Key (通过 `API_KEY` Header 传递)

---

## 1. Projects (项目管理)

### 1.1 创建项目

```
POST /v1/projects
```

**请求体:**
```json
{
  "name": "string (必填)",
  "instruction": "string (可选，项目默认指令)"
}
```

**响应 (200):**
```json
{
  "id": "string",
  "name": "string",
  "instruction": "string",
  "created_at": 1234567890
}
```

### 1.2 列出项目

```
GET /v1/projects
```

**查询参数:**
| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| limit | integer | 100 | 最大返回数量 (1-1000) |

**响应 (200):**
```json
{
  "data": [
    {
      "id": "string",
      "name": "string",
      "instruction": "string",
      "created_at": 1234567890
    }
  ]
}
```

---

## 2. Tasks (任务管理)

### 2.1 创建任务

```
POST /v1/tasks
```

**请求体:**
```json
{
  "prompt": "string (必填，任务提示/指令)",
  "agentProfile": "manus-1.6 | manus-1.6-lite | manus-1.6-max",
  "attachments": ["文件ID/URL/Base64"],
  "taskMode": "chat | adaptive | agent",
  "connectors": ["connector_id"],
  "hideInTaskList": false,
  "createShareableLink": false,
  "taskId": "string (多轮对话时传入)",
  "locale": "string",
  "projectId": "string",
  "interactiveMode": false
}
```

**响应 (200):**
```json
{
  "task_id": "string",
  "task_title": "string",
  "task_url": "string",
  "share_url": "string (仅当 createShareableLink=true)"
}
```

### 2.2 列出任务

```
GET /v1/tasks
```

**查询参数:**
| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| after | string | - | 分页游标 (上一页最后一个 task_id) |
| limit | integer | 100 | 每页数量 (1-1000) |
| order | string | desc | 排序方向 (asc/desc) |
| orderBy | string | created_at | 排序字段 (created_at/updated_at) |
| query | string | - | 搜索关键词 |
| status | string | - | 状态过滤 (pending/running/completed/failed) |
| createdAfter | integer | - | 最小创建时间 (Unix 时间戳) |
| createdBefore | integer | - | 最大创建时间 (Unix 时间戳) |
| project_id | string | - | 项目ID过滤 |

**响应 (200):**
```json
{
  "object": "list",
  "data": [...],
  "first_id": "string",
  "last_id": "string",
  "has_more": true
}
```

### 2.3 获取任务详情

```
GET /v1/tasks/{task_id}
```

**路径参数:**
- `task_id` (必填): 任务ID

**查询参数:**
| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| convert | boolean | false | 转换输出格式 (如 pptx) |

**响应 (200):**
```json
{
  "id": "string",
  "object": "task",
  "created_at": 1234567890,
  "updated_at": 1234567890,
  "status": "pending | running | completed | failed",
  "error": "string (失败时)",
  "incomplete_details": "string",
  "instructions": "string",
  "max_output_tokens": 1000,
  "model": "string",
  "task_title": "string",
  "task_url": "string",
  "metadata": {},
  "output": [
    {
      "role": "user | assistant",
      "content": [...]
    }
  ],
  "locale": "string",
  "credit_usage": 100
}
```

### 2.4 更新任务

```
PUT /v1/tasks/{task_id}
```

**请求体:**
```json
{
  "title": "string",
  "enableShared": true,
  "enableVisibleInTaskList": true
}
```

**响应 (200):**
```json
{
  "task_id": "string",
  "task_title": "string",
  "task_url": "string",
  "share_url": "string"
}
```

### 2.5 删除任务

```
DELETE /v1/tasks/{task_id}
```

**响应 (200):**
```json
{
  "id": "string",
  "object": "task.deleted",
  "deleted": true
}
```

---

## 3. Files (文件管理)

### 3.1 创建文件 (获取上传URL)

```
POST /v1/files
```

**请求体:**
```json
{
  "filename": "string (必填)"
}
```

**响应 (200):**
```json
{
  "id": "string",
  "object": "file",
  "filename": "string",
  "status": "pending",
  "upload_url": "string (S3 预签名URL，使用 PUT 上传)",
  "upload_expires_at": "2023-11-07T05:31:56Z",
  "created_at": "2023-11-07T05:31:56Z"
}
```

**上传流程:**
1. 调用此接口获取 `upload_url`
2. 使用 `PUT` 请求将文件内容上传到 `upload_url`
3. 上传完成后，文件状态变为 `uploaded`
4. 在创建任务时使用 `file_id` 作为附件

### 3.2 列出文件

```
GET /v1/files
```

**说明:** 返回最近上传的 10 个文件

**响应 (200):**
```json
{
  "object": "list",
  "data": [
    {
      "id": "string",
      "object": "file",
      "filename": "string",
      "status": "pending | uploaded | deleted",
      "created_at": "2023-11-07T05:31:56Z"
    }
  ]
}
```

### 3.3 获取文件详情

```
GET /v1/files/{file_id}
```

**响应 (200):**
```json
{
  "id": "string",
  "object": "file",
  "filename": "string",
  "status": "pending | uploaded | deleted",
  "created_at": "2023-11-07T05:31:56Z"
}
```

### 3.4 删除文件

```
DELETE /v1/files/{file_id}
```

**响应 (200):**
```json
{
  "id": "string",
  "object": "file.deleted",
  "deleted": true
}
```

---

## 4. Webhooks

### 4.1 创建 Webhook

```
POST /v1/webhooks
```

**请求体:**
```json
{
  "webhook": {
    "url": "string (Webhook 接收地址)"
  }
}
```

**响应 (200):**
```json
{
  "webhook_id": "string"
}
```

### 4.2 删除 Webhook

```
DELETE /v1/webhooks/{webhook_id}
```

**响应 (204):** 无响应体

---

## 5. 状态枚举

### Task Status
| 值 | 说明 |
|---|------|
| pending | 等待执行 |
| running | 执行中 |
| completed | 已完成 |
| failed | 执行失败 |

### File Status
| 值 | 说明 |
|---|------|
| pending | 等待上传 |
| uploaded | 已上传 |
| deleted | 已删除 |

### Agent Profile
| 值 | 说明 |
|---|------|
| manus-1.6 | 默认模型 |
| manus-1.6-lite | 轻量模型 |
| manus-1.6-max | 增强模型 |

### Task Mode
| 值 | 说明 |
|---|------|
| chat | 对话模式 |
| adaptive | 自适应模式 |
| agent | Agent 模式 |

---

## 6. 错误处理

| HTTP 状态码 | 说明 |
|-------------|------|
| 200 | 成功 |
| 204 | 成功 (无响应体) |
| 400 | 请求参数错误 |
| 401 | 未授权 (API Key 无效) |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

---

## 7. API 端点汇总

| 方法 | 路径 | 功能 |
|------|------|------|
| POST | /v1/projects | 创建项目 |
| GET | /v1/projects | 列出项目 |
| POST | /v1/tasks | 创建任务 |
| GET | /v1/tasks | 列出任务 |
| GET | /v1/tasks/{task_id} | 获取任务详情 |
| PUT | /v1/tasks/{task_id} | 更新任务 |
| DELETE | /v1/tasks/{task_id} | 删除任务 |
| POST | /v1/files | 创建文件 (获取上传URL) |
| GET | /v1/files | 列出文件 |
| GET | /v1/files/{file_id} | 获取文件详情 |
| DELETE | /v1/files/{file_id} | 删除文件 |
| POST | /v1/webhooks | 创建 Webhook |
| DELETE | /v1/webhooks/{webhook_id} | 删除 Webhook |
