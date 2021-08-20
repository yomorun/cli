# YoMo CLI
Command-line tools for YoMo

## Prerequisites

[Installing Go](https://golang.org/doc/install)

## Installing
You can easily install the latest release globally by running:

```sh
go install github.com/yomorun/cli/yomo@latest
```
Or you can install into another directory:

```sh
env GOBIN=/bin go install github.com/yomorun/cli/yomo@latest
```

## Getting Started

### 1. Source

#### Write a source app

See [example/source/main.go](https://github.com/yomorun/cli/blob/main/example/source/main.go)

#### Run

```sh
go run main.go
```

### 2. Stream Function

#### Init

Create a stream function

```sh
yomo init [Name]
```

#### Run

```sh
yomo run --name [Name] app.go
```

### 3. Output Connector

#### Write an output connector

See [example/output-connector/main.go](https://github.com/yomorun/cli/blob/main/example/output-connector/main.go)

#### Run

```sh
go run main.go
```

### 4. YoMo Server

#### Configure yomo server `workflow.yaml`

```yaml
name: Service
host: localhost
port: 9000
functions:
  - name: Noise
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
