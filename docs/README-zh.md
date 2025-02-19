[English](./README-en.md) |
[中文](./README-zh.md) |
[日本語](./README-ja.md)

<div align="center">
  <h1>OpenStress</h1>
  <p>一个高性能的 Go 语言压力测试框架</p>
  
  <a href="https://github.com/potatoImp/OpenStress/blob/main/LICENSE-CODE">
    <img alt="代码许可证" src="https://img.shields.io/badge/Code_License-MIT-f5de53?&color=f5de53"/>
  </a>
  <a href="https://github.com/potatoImp/OpenStress/blob/main/LICENSE-MODEL">
    <img alt="模型许可证" src="https://img.shields.io/badge/Model_License-Model_Agreement-f5de53?&color=f5de53"/>
  </a>
  <a href="https://golang.org/doc/install">
    <img alt="Go 版本" src="https://img.shields.io/badge/Go-%3E%3D%201.16-blue"/>
  </a>
  <a href="https://github.com/potatoImp/OpenStress/releases">
    <img alt="GitHub 发布" src="https://img.shields.io/github/v/release/potatoImp/OpenStress?color=brightgreen"/>
  </a>
</div>

## 目录

1. [介绍](#介绍)
2. [特性](#特性)
3. [快速开始](#快速开始)
4. [安装](#安装)
5. [许可证](#许可证)
6. [联系方式](#联系方式)

## 介绍

OpenStress 是一个开源的任务管理和日志记录框架，旨在简化 Go 语言并发应用程序的开发。它提供了一种高效的方式来管理任务和记录应用程序事件，是开发人员提高生产力和可维护性的理想选择。

## 特性

- [x] **并发任务管理**: OpenStress 允许您创建和管理工作线程池，从而实现任务的高效并行执行。这有助于最大化资源利用率并提高应用程序性能。

- **灵活的日志系统**: 内置的日志框架支持各种日志级别（INFO、WARN、ERROR、DEBUG），允许开发人员轻松跟踪应用程序行为和诊断问题。日志可以自定义并定向到不同的输出。

- **错误处理**: OpenStress 包含一个强大的错误处理机制，允许开发人员定义自定义错误类型并有效管理错误状态。这增强了使用 OpenStress 构建的应用程序的可靠性。

- **基于 Web 的任务管理界面**: OpenStress 提供了一个可视化的 Web 界面用于任务管理，配备用户身份验证和用于启动和停止任务的 API 端点。它还支持 Swagger 文档以便于集成和使用。

- **简化的开发过程**: 开发人员可以专注于执行任务，而无需担心创建和管理协程池或处理并发压力。OpenStress 负责标准测试数据的收集、输出和报告，简化了开发工作流程。

- **易于集成**: OpenStress 设计简单，易于集成到现有的 Go 项目中。模块化架构允许与其他库和框架无缝使用。

- **开源**: 作为一个开源项目，OpenStress 鼓励社区贡献和协作。开发人员可以自由使用、修改和分发该框架，营造创新和改进的环境。

## 快速开始

要开始使用 OpenStress，请按照以下步骤操作：

1. **克隆仓库**:
   ```bash
   git clone https://github.com/potatoImp/OpenStress.git
   cd OpenStress
2. **安装依赖**:

   确保您的机器上安装了 Go。使用以下命令安装必要的依赖项:
   ```bash
   go mod tidy
4. **Run the Application**:
  
   You can run the main application using:
   ```bash
   go run main.go


## 4. Evaluation Results
#### Standard Benchmarks

<div align="center">


|  | Benchmark (Metric) | OpenStress | locust | jmeter | - | - |
|---|-------------------|----------|--------|-------------|---------------|---------|
| requestTotal | - | one million | one million | one million | Dense | MoE |
| QPS | - | **34575** | 4687 | 29111 | - | - |
| requestTime | - | **1.5ms** | 153ms | 6ms | - | - |
| MinRequestTime | - | <1ms | 5ms | <1ms | - | - |
| MaxRequestTime | - | 1150ms | 1336ms | 3210ms | - | - |
| SuccessRate | - | 100% | 100% | 100% | - | - |
| CPU | - | 100% | 100% | 100% | - | - |
| Summary Report | - | yes | yes | yes | - | - |
| HtmlReport | - | yes | yes | yes | - | - |
| AllTestData | - | yes | no | yes | - | - |
| Artificial Intelligence Analysis| - | yes | no | no | - | - |








</div>

> [!NOTE]
> 8c 16G
> For more evaluation details, please check our paper. 


## 安装

### 前置条件

- Go 1.16 或更高版本
- Redis（可选，用于分布式场景）

### 安装步骤

1. 克隆仓库:
```bash
git clone https://github.com/potatoImp/OpenStress.git
cd OpenStress
```

## 许可证
OpenStress 根据 MIT 许可证授权。有关更多信息，请参阅 LICENSE 文件。

## 联系方式
如有任何疑问或需要支持，请通过仓库的问题部分与我们联系或直接联系维护人员。


## 致谢

我们要感谢以下库的作者和贡献者，他们使这个项目成为可能：

- **go-echarts**: A powerful charting library for Go.
- **go-redis**: A Redis client for Go.
- **gokrb5**: A Go library for Kerberos authentication.
- **ants**: A high-performance goroutine pool for Go.
- **zap**: A fast, structured, and leveled logging library for Go.
- **lumberjack**: A log rolling package for Go.
- **yaml**: A YAML support library for Go.
- **xxhash**: A fast non-cryptographic hash algorithm.
