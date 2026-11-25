# Go gRPC 通信示例

一个本机运行的一元 RPC 示例：客户端发送名称，服务端返回问候语。
协议为 `helloworld.v1.Greeter/SayHello`，定义见
`proto/helloworld/v1/helloworld.proto`；生成代码已包含在仓库中，普通构建无需 protoc。
此示例使用明文 gRPC，默认绑定回环地址，仅供本地学习，不是生产安全配置。

## 环境与构建

需要 Go 1.24.0 或更新版本。实测 Go 1.24.13（Linux arm64）。
首次准备依赖需要能访问 Go 模块源；版本及校验和由 go.mod/go.sum 固定。
在本仓库根目录执行：

```sh
go mod download
go build ./...
go test ./...
mkdir -p bin
go build -o bin/server ./cmd/server
go build -o bin/client ./cmd/client
```

## 两个终端运行

终端一，在仓库根目录启动服务端：

```sh
./bin/server
# 日志包含 listening on 127.0.0.1:50051
```

终端二，在同一仓库根目录调用：

```sh
./bin/client
# Hello, World!
./bin/client -name ' Alice '
# Hello, Alice!
```

可连续调用。服务端按 Ctrl-C 或收到 SIGTERM 时停止，优雅关闭最多等待 3 秒，之后强制停止。
自定义地址时，在终端一停止旧服务并运行：

```sh
./bin/server -addr 127.0.0.1:50052
```

终端二运行：

```sh
./bin/client -addr 127.0.0.1:50052 -name 小明 -timeout 3s
# Hello, 小明!
```

参数与默认值：

| 程序 | 参数 | 默认值 | 说明 |
| --- | --- | --- | --- |
| server | -addr | 127.0.0.1:50051 | TCP 监听地址，启动日志显示实际地址 |
| client | -addr | 127.0.0.1:50051 | 目标地址 |
| client | -name | World | 名称，服务端去除首尾 Unicode 空白 |
| client | -timeout | 3s | 单次 RPC deadline，Go duration，必须大于 0 |
| 两者 | -h | — | 打印帮助并成功退出 |

客户端成功向 stdout 输出一行问候，退出码为 0；失败向 stderr 报错并非零退出。
客户端一次运行仅发起一次请求，连接建立也受 RPC deadline 限制。

## 基础验证与故障排查

`go test ./...` 通过内存传输执行真实 gRPC 调用，覆盖普通名称、中文、首尾空白、内部空格以及空名称和全空白名称的 InvalidArgument 状态。
可在已启动服务端的情况下手动验证：

```sh
./bin/client -name ''       # InvalidArgument，非零退出
./bin/client -name '   '    # InvalidArgument，非零退出
./bin/client -timeout 0s    # timeout must be greater than zero，非零退出
./bin/client -timeout=-1s   # 同上
```

- 服务端出现 `address already in use`：停止占用进程或用 `-addr` 选择其他端口。
- 客户端出现 `Unavailable` / `connection refused`：确认服务端已启动，且两个进程的地址一致。
- 出现 `DeadlineExceeded`：确认地址及网络可达；必要时增大 `-timeout`。离线调用可能在 deadline 前立即返回 Unavailable。
- 出现 `InvalidArgument`：提供非空白名称；显式空名称不会替换为 World。
- 模块下载失败：检查网络或配置可访问的 GOPROXY 后重试 `go mod download`。

停止目标服务后，`./bin/client -timeout 1s` 应有界失败，非零退出。

## 重新生成协议代码（可选）

固定工具版本：protoc **3.21.12**、protoc-gen-go **v1.36.6**、
protoc-gen-go-grpc **v1.5.1**。普通运行无需安装这些工具。

Debian 12 可安装精确版本的 protoc（需要 apt 软件源中存在该版本）：

```sh
sudo apt-get update
sudo apt-get install protobuf-compiler=3.21.12-3
```

其他 Linux/macOS 可从官方 protobuf 发布页下载 v21.12 对应系统的压缩包，解压并将 bin 加入 PATH：
<https://github.com/protocolbuffers/protobuf/releases/tag/v21.12>。
例如 Linux arm64：

```sh
curl -fLO https://github.com/protocolbuffers/protobuf/releases/download/v21.12/protoc-21.12-linux-aarch_64.zip
unzip protoc-21.12-linux-aarch_64.zip -d .protoc
export PATH="$PWD/.protoc/bin:$PATH"
```

在仓库根目录安装固定版本 Go 插件并生成（工具放在仓库外，避免混入源码）：

```sh
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.6
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
export PATH="$(go env GOPATH)/bin:$PATH"
protoc --version                 # libprotoc 3.21.12
protoc-gen-go --version          # protoc-gen-go v1.36.6
protoc-gen-go-grpc --version     # protoc-gen-go-grpc 1.5.1
mkdir -p gen
protoc -I proto --go_out=gen --go_opt=paths=source_relative \
  --go-grpc_out=gen --go-grpc_opt=paths=source_relative \
  proto/helloworld/v1/helloworld.proto
git diff --exit-code -- gen
```

在已提交的干净检出上，最后一条命令应无差异。请保持相同工具版本；不同版本可能改变生成内容。
