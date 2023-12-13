# 使用官方 Golang 镜像作为基础镜像
FROM golang:latest

# 设置工作目录
WORKDIR /app/go

# 复制所需文件到工作目录
COPY go/main.go .
COPY go/go.mod .

# 构建 Go 应用
RUN go build -o main

# 暴露端口
EXPOSE 82

# 启动应用
CMD ["./main"]

