# httpMock 服务

## 简介
httpMock 是一个用于性能测试的模拟 HTTP 服务器，它提供了可配置的响应延迟、成功率控制等功能，适用于各种性能测试场景。

## 主要特性

- **高并发处理**：支持大量并发连接
- **响应延迟控制**：可配置固定或随机响应延迟
- **成功率控制**：可设置请求的成功/失败比率
- **长连接支持**：支持 HTTP Keep-Alive
- **自定义响应**：支持配置响应内容和状态码

## 使用方法

### 启动服务

在 Windows 系统下：
```bash
.\httpMock.exe [options]
```

在 Linux 系统下：
```bash
./httpMock [options]
```

### 配置参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| -port | 服务监听端口 | 8080 |
| -delay | 响应延迟（毫秒） | 0 |
| -success-rate | 成功率（0-100） | 100 |
| -response-size | 响应体大小（字节） | 1024 |

### 示例

1. 启动基本服务：
Windows:
```bash
.\httpMock.exe
```
Linux:
```bash
./httpMock
```

2. 指定端口和延迟：
Windows:
```bash
.\httpMock.exe -port 9090 -delay 100
```
Linux:
```bash
./httpMock -port 9090 -delay 100
```

3. 设置80%成功率：
Windows:
```bash
.\httpMock.exe -success-rate 80
```
Linux:
```bash
./httpMock -success-rate 80
```

## API 接口

### 健康检查
- 路径：`/health`
- 方法：GET
- 响应：返回服务状态信息

### 模拟接口
- 路径：`/mock`
- 方法：GET/POST
- 响应：返回配置的响应内容

## 性能指标

- 支持并发连接数：10000+
- 最小响应延迟：<1ms
- 内存占用：<50MB

## 注意事项

1. 建议在测试环境中使用
2. 大量并发连接时注意系统资源限制
3. 响应延迟设置过大可能影响测试效率

## 构建

在 Windows 系统下构建：
```bash
go build -o httpMock.exe httpMock.go
```

在 Linux 系统下构建：
```bash
go build -o httpMock httpMock.go
```

## 贡献

欢迎提交 Issue 和 Pull Request 来帮助改进项目。

## 许可证

本项目采用 MIT 许可证。