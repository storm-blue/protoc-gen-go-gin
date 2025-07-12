.PHONY: build install test clean example

# 构建项目
build:
	go build -o protoc-gen-go-gin .

# 安装到本地
install:
	go install .

# 测试
test:
	go test -v ./...

# 清理
clean:
	rm -f protoc-gen-go-gin
	rm -f *.pb.go

# 生成示例代码
example:
	protoc --go_out=. --go-gin_out=. example.proto

# 运行示例
run-example: example
	go run main.go

# 格式化代码
fmt:
	go fmt ./...

# 检查代码
vet:
	go vet ./...

# 构建并安装
all: fmt vet build install