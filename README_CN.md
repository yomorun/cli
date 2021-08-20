# YoMo CLI
YoMo 命令行工具

## 前置条件
❗️确保已安装 Go 编译运行环境，参考 [Installing Go](https://golang.org/doc/install)
## 安装
```sh
go install github.com/yomorun/cli/yomo@latest
```

## 快速指南

### 1. Source 应用程序(数据来源)
#### 编写数据生产应用程序
参见 [example/source/main.go](https://github.com/yomorun/cli/blob/main/example/source/main.go)

#### 运行 Source 应用

```
go run main.go
```

### 2. Stream Function 流处理函数
#### 初始化一个流处理函数 

```sh
yomo init [Name]
```

#### 运行流处理函数

```shell
yomo run --name [Name] app.go
```

### 3. Output Connector (数据输出)
#### 编写数据消费应用程序
参见 [example/output-connector/main.go](https://github.com/yomorun/cli/blob/main/example/output-connector/main.go)

#### 运行 Output Connector 应用

```shell
go run main.go
```

### 4. YoMo Server 应用编排
#### 编写工作流配置文件 `workflow.yaml`

```yaml
name: Service
host: localhost
port: 9000
functions:
  - name: Noise
```

#### 运行 YoMo Server 应用程序

```shell
yomo serve --config workflow.yaml
```

## 示例

### 前置条件
- 安装 [task](https://taskfile.dev/#/installation)

### 运行示例

```shell
task example
```

