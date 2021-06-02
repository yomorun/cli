# YoMo CLI
YoMo 命令行工具

## 安装
```
go install github.com/yomorun/cli/yomo@latest
```

## 使用
❗️确保已安装 Go 编译运行环境，参考 [Installing Go](https://golang.org/doc/install)

### Source 应用程序(数据来源)
#### 编写数据生产应用程序

#### 运行 Source 应用

```
go run main.go
```

### YoMo 流处理函数
#### 编写流处理函数 `app.go`

#### 运行流处理函数

##### 开发环境

```
yomo run --name [Name] app.go
```

##### 生产环境

- 编译函数

```
yomo build --name [Name] app.go
```

- 执行流处理函数
  *nix 环境

    ```
    ./sl.yomo
    ```

	Windows 环境
    ```
    sl.exe
    ```


### Sink 应用程序(数据输出)
#### 编写数据消费应用程序

#### 运行 Sink 应用

```
go run main.go
```

### Zipper 应用编排
#### 编写工作流配置文件 `workflow.yaml`

```yaml
name: Service
host: localhost
port: 9000
flows:
  - name: Noise
sinks:
  - name: MockDB
```

#### 运行 YoMo 应用程序

```
yomo serve --config workflow.yaml
```



## 示例

### 前置条件
- 安装 [task](https://taskfile.dev/#/installation)

### 运行 Zipper 

#### 运行示例 Zipper 服务
```
task example-zipper
```
### 运行示例

#### 基础示例

```
task example
```



## TODO

- serverless 增加builder 子目录用于不同语言构建?
- log 更名 printer?
- serverless options 是否可以和 workflow config共用?

