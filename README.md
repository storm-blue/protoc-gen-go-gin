# protoc-gen-go-gin

一个用于从 Protocol Buffers 定义文件自动生成 Gin HTTP 服务代码的工具，支持自动生成 Swagger 注释。

## 功能特性

- 从 `.proto` 文件自动生成 Gin HTTP 服务代码
- 支持多种 HTTP 方法映射（GET、POST、PUT、DELETE、PATCH）
- 支持 Google API 的 HTTP 注解
- **自动生成 Swagger 注释**，支持 API 文档生成
- 智能路由生成和参数绑定
- 统一的错误处理和响应格式

## 安装

安装依赖:

- [go 1.22](https://golang.org/dl/)
- [protoc](https://github.com/protocolbuffers/protobuf)
- [protoc-gen-go](https://github.com/protocolbuffers/protobuf-go)

```bash
go install github.com/storm-blue/protoc-gen-go-gin@latest
```

## 使用方法

### 1. 定义 Proto 文件

在 `.proto` 文件中定义服务和方法，并添加注释：

```protobuf
syntax = "proto3";

package example;

import "google/api/annotations.proto";

option go_package = "github.com/example/proto";

// UserService 用户服务
service UserService {
  // GetUser 获取用户信息
  // 根据用户ID获取用户的详细信息
  // @tag:user
  rpc GetUser(GetUserRequest) returns (GetUserResponse) {
    option (google.api.http) = {
      get: "/users/{id}"
    };
  }

  // CreateUser 创建用户
  // 创建新用户并返回用户信息
  // @tag:user
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse) {
    option (google.api.http) = {
      post: "/users"
      body: "*"
    };
  }
}
```

### 2. 生成代码

使用 protoc 编译器生成代码：

```bash
protoc --go_out=. --go-gin_out=. example.proto
```

### 3. 生成的代码

生成的 `*_gin.pb.go` 文件将包含带有 Swagger 注释的 HTTP 处理函数：

```go
// @Summary GetUser
// @Description 根据用户ID获取用户的详细信息
// @Tags user
// @Accept json
// @Produce json
// @Param request body GetUserRequest true "Request body"
// @Success 200 {object} GetUserResponse
// @Failure 400 {object} map[string]interface{} "Invalid parameter"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /users/{id} [GET]
func (s *UserService) GetUser_0(ctx *gin.Context) {
    // 处理逻辑...
}
```

## 注释格式

### 方法注释

在 `.proto` 文件中，你可以为每个 RPC 方法添加注释：

```protobuf
// 方法摘要
// 详细描述信息
// @tag:标签名
rpc MethodName(Request) returns (Response) {
    option (google.api.http) = {
        get: "/path"
    };
}
```

### 注释字段说明

- **第一行**: 用作 Swagger 的 `@Summary`
- **后续行**: 用作 Swagger 的 `@Description`
- **@tag:标签**: 用作 Swagger 的 `@Tags`

### 特殊标记

- **@deprecated**: 如果方法被标记为 deprecated，生成的代码会自动添加 `@Deprecated` 注释

## 集成 Swagger

生成的代码可以直接与 `swaggo/swag` 配合使用：

1. 安装 swag：
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

2. 在项目根目录运行：
```bash
swag init
```

3. 在 Gin 应用中集成 Swagger UI：
```go
import (
    "github.com/gin-gonic/gin"
    "github.com/swaggo/gin-swagger"
    "github.com/swaggo/files"
    _ "your-project/docs"
)

func main() {
    r := gin.Default()
    
    // Swagger UI
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    
    // 你的路由...
    r.Run(":8080")
}
```

## 示例

查看 `example.proto` 文件了解完整的使用示例。

## 许可证

MIT License
