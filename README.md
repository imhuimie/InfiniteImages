# InfiniteImages

InfiniteImages 是一个高性能的图片存储和处理服务，支持多种存储后端和图片格式转换。

## 功能特点

- 支持多种存储后端：本地文件系统、S3 兼容存储、FTP
- 自动图片格式转换：支持 WebP、AVIF 等现代图片格式
- 图片处理：调整大小、压缩、添加水印
- 安全特性：API 密钥认证、IP 白名单/黑名单、CORS 配置
- 元数据管理：支持标签、过期时间等
- 自动清理：可配置自动清理过期图片
- 高性能：异步处理、缓存支持

## 快速开始

### 前置条件

- Go 1.18 或更高版本
- 对于 AVIF 格式支持，需要安装 [libavif](https://github.com/AOMediaCodec/libavif)

### 安装

1. 克隆仓库

```bash
git clone https://github.com/emper0r/InfiniteImages.git
cd InfiniteImages
```

2. 安装依赖

```bash
go mod tidy
```

3. 配置

复制 `.env.example` 文件为 `.env` 并根据需要修改配置：

```bash
cp .env.example .env
```

4. 运行

```bash
go run main.go
```

或者构建并运行：

```bash
go build
./InfiniteImages

# 服务将在 http://localhost:6868 上运行
```

### Docker 部署

1. 构建 Docker 镜像

```bash
docker build -t infiniteimages .
```

2. 运行容器

```bash
docker run -p 6868:6868 -v ./data:/app/data -v ./config:/app/config infiniteimages
```

## API 文档

### 认证

所有 API 请求（除了公共 API）都需要通过 API 密钥进行认证。API 密钥可以通过以下方式提供：

- 请求头：`X-API-Key: your_api_key`
- 查询参数：`?api_key=your_api_key`

### 上传图片

```
POST /api/upload
```

请求体：
- `images`：图片文件（可多个）

响应：
```json
{
  "success": true,
  "message": "成功上传 1 张图片",
  "data": ["/static/images/original/landscape/1617123456789"]
}
```

### 获取图片列表

```
GET /api/images
```

查询参数：
- `page`：页码（默认 1）
- `limit`：每页数量（默认 20）
- `tag`：按标签筛选

响应：
```json
{
  "success": true,
  "message": "获取图片列表成功",
  "total": 100,
  "page": 1,
  "limit": 20,
  "data": [
    {
      "id": "1617123456789",
      "filename": "example.jpg",
      "url": "/static/images/original/landscape/1617123456789",
      "thumbnailUrl": "/static/images/webp/landscape/1617123456789.webp",
      "size": 1024000,
      "width": 1920,
      "height": 1080,
      "format": "jpeg",
      "orientation": "landscape",
      "tags": ["nature", "landscape"],
      "createdAt": "2023-04-01T12:34:56Z"
    }
  ]
}
```

### 获取随机图片

```
GET /api/random
```

响应：
```json
{
  "success": true,
  "message": "获取随机图片成功",
  "data": {
    "id": "1617123456789",
    "url": "/static/images/original/landscape/1617123456789",
    "filename": "example.jpg"
  }
}
```

### 删除图片

```
DELETE /api/images/:id
```

响应：
```json
{
  "success": true,
  "message": "删除图片成功",
  "id": "1617123456789"
}
```

### 更新图片标签

```
PUT /api/images/:id/tags
```

请求体：
```json
{
  "tags": ["nature", "landscape", "mountain"]
}
```

响应：
```json
{
  "success": true,
  "message": "更新标签成功",
  "data": {
    "id": "1617123456789",
    "tags": ["nature", "landscape", "mountain"]
  }
}
```

### 获取所有标签

```
GET /api/tags
```

响应：
```json
{
  "success": true,
  "message": "获取标签列表成功",
  "data": [
    {
      "name": "nature",
      "count": 42
    },
    {
      "name": "landscape",
      "count": 27
    }
  ]
}
```

### 手动触发清理过期图片

```
POST /api/trigger-cleanup
```

响应：
```json
{
  "success": true,
  "message": "成功清理 5 张过期图片",
  "data": {
    "count": 5
  }
}
```

## 配置说明

详细的配置选项可以在 `.env.example` 文件中找到，包括：

- 服务器配置
- 存储配置
- 图像处理配置
- 安全配置
- 水印配置
- 缓存配置
- 日志配置

## 存储后端

### 本地存储

默认使用本地文件系统存储图片，路径为 `static/images`。

### S3 兼容存储

支持 Amazon S3 和其他兼容 S3 API 的存储服务，如 MinIO、DigitalOcean Spaces 等。

### FTP 存储

支持通过 FTP 协议存储图片到远程服务器。

## 贡献

欢迎贡献代码、报告问题或提出建议！请提交 Issue 或 Pull Request。

## 许可证

本项目采用 MIT 许可证。