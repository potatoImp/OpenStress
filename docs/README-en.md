[English](./README-en.md) |
[中文](./README-zh.md) |
[日本語](./README-ja.md)

<div align="center">
  <h1>OpenStress</h1>
  <p>A High-Performance Stress Testing Framework in Go</p>
  
  <a href="https://github.com/potatoImp/OpenStress/blob/main/LICENSE-CODE">
    <img alt="Code License" src="https://img.shields.io/badge/Code_License-MIT-f5de53?&color=f5de53"/>
  </a>
  <a href="https://github.com/potatoImp/OpenStress/blob/main/LICENSE-MODEL">
    <img alt="Model License" src="https://img.shields.io/badge/Model_License-Model_Agreement-f5de53?&color=f5de53"/>
  </a>
  <a href="https://golang.org/doc/install">
    <img alt="Go Version" src="https://img.shields.io/badge/Go-%3E%3D%201.16-blue"/>
  </a>
  <a href="https://github.com/potatoImp/OpenStress/releases">
    <img alt="GitHub release" src="https://img.shields.io/github/v/release/potatoImp/OpenStress?color=brightgreen"/>
  </a>
</div>

## Table of Contents

1. [Introduction](#introduction)
2. [Features](#features)
3. [QuickStart](#quick-started)
4. [Installation](#Installation)
5. [License](#license)
6. [Contact](#contact)

## Introduction

OpenStress is an open-source task management and logging framework designed to simplify the development of concurrent applications in Go. It provides an efficient way to manage tasks and log application events, making it an ideal choice for developers looking to enhance their productivity and maintainability.

## Features

- [x] **Concurrent Task Management**: OpenStress allows you to create and manage a pool of worker threads, enabling efficient execution of tasks in parallel. This helps in maximizing resource utilization and improving application performance.

- **Flexible Logging System**: The built-in logging framework supports various log levels (INFO, WARN, ERROR, DEBUG), allowing developers to track application behavior and diagnose issues easily. Logs can be customized and directed to different outputs.

- **Error Handling**: OpenStress includes a robust error handling mechanism, allowing developers to define custom error types and manage error states effectively. This enhances the reliability of applications built with OpenStress.

- **Web-Based Task Management Interface**: OpenStress provides a visual web interface for task management, complete with user authentication and API endpoints for starting and stopping tasks. It also supports Swagger documentation for easy integration and usage.

- **Simplified Development Process**: Developers can focus solely on executing tasks without worrying about creating and managing coroutine pools or handling concurrency pressure. OpenStress takes care of standard test data collection, output, and reporting, streamlining the development workflow.

- **Easy Integration**: Designed with simplicity in mind, OpenStress can be easily integrated into existing Go projects. The modular architecture allows for seamless usage alongside other libraries and frameworks.

- **Open Source**: As an open-source project, OpenStress encourages community contributions and collaboration. Developers can freely use, modify, and distribute the framework, fostering an environment of innovation and improvement.

## Quick Started

To get started with OpenStress, follow these steps:

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/potatoImp/OpenStress.git
   cd OpenStress
2. **Install Dependencies**:

   Ensure you have Go installed on your machine. Use the following command to install necessary dependencies:
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


## Installation

### Prerequisites

- Go 1.16 or higher
- Redis (optional, for distributed scenarios)

### Installation Steps

1. Clone the repository:
```bash
git clone https://github.com/potatoImp/OpenStress.git
cd OpenStress
```

## License
OpenStress is licensed under the MIT License. See the LICENSE file for more information.

## Contact
For any inquiries or support, please reach out via the issues section of the repository or contact the maintainers directly.


## Acknowledgments

We would like to express our gratitude to the authors and contributors of the following libraries that made this project possible:

- **go-echarts**: A powerful charting library for Go.
- **go-redis**: A Redis client for Go.
- **gokrb5**: A Go library for Kerberos authentication.
- **ants**: A high-performance goroutine pool for Go.
- **zap**: A fast, structured, and leveled logging library for Go.
- **lumberjack**: A log rolling package for Go.
- **yaml**: A YAML support library for Go.
- **xxhash**: A fast non-cryptographic hash algorithm.
