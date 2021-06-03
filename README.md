# YoMo CLI
Command-line tools for YoMo

## Prerequisites

[Installing Go](https://golang.org/doc/install)

## Installing
```sh
go install github.com/yomorun/cli/yomo@latest
```

## Getting Started

### 1.Source

#### Write a source app

See [example/source/main.go](https://github.com/yomorun/cli/blob/main/example/source/main.go)

#### Run

```sh
go run main.go
```

### 2.Flow

#### Init

Create a serverless function 

```sh
yomo init [Name]
```

#### Run

```sh
yomo run --name [Name] app.go
```


### 3.Sink
#### Write a sink app

See [example/sink/main.go](https://github.com/yomorun/cli/blob/main/example/sink/main.go)

#### Run

```sh
go run main.go
```

### 4.Zipper
#### Configure zipper `workflow.yaml`

```yaml
name: Service
host: localhost
port: 9000
flows:
  - name: Noise
sinks:
  - name: MockDB
```

#### Run

```sh
yomo serve --config workflow.yaml
```

## Example

### Prerequisites
[Installing task](https://taskfile.dev/#/installation)

### Simple Example

#### Run

```sh
task example
```

### Edge-Mesh

#### Run US Node

```sh
task example-mesh-us
```

#### Run EU Node

```sh
task example-mesh-eu
```
