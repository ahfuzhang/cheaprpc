# 介绍

打算实现一个非常简单的golang的RPC框架，非常简单，前期仅支持HTTP协议的通讯。对于大多数的小型业务是够用的。

# 已经实现的功能

0.环境准备

- 把gogoproto库git clone到本地:

```shell
mkdir -p /Users/ahfuzhang/code/github.com/gogo
cd /Users/ahfuzhang/code/github.com/gogo
git clone https://github.com/gogo/protobuf.git
```

- 把此项目复制到本地

```shell
mkdir -p /Users/ahfuzhang/code/github.com/ahfuzhang
cd /Users/ahfuzhang/code/github.com/ahfuzhang
git clone https://github.com/ahfuzhang/cheaprpc.git
```

1.先定义一个服务的proto文件：

- 服务名
- 服务接口
- 每条接口的请求和响应格式
  see: [examples/my_easy_service.proto](examples/my_easy_service.proto)

2.生成代码

```shell
cd /Users/ahfuzhang/code/github.com/ahfuzhang/cheaprpc/examples
# proto_path= 这个路径是可以找到 github.com/gogo/protobuf 的父目录
make build proto_path="/Users/ahfuzhang/code/"
```

然后就会在当前目录生成服务的基本代码。

3.填充自己的业务代码
see: [myeasyservice.go](examples/github.com/ahfuzhang/my_easy_service/internal/services/myeasyservice/myeasyservice.go)
在这个文件中，为方法添加具体的业务代码

4.运行服务：

```shell
cd /Users/ahfuzhang/code/github.com/ahfuzhang/cheaprpc/examples/github.com/ahfuzhang/my_easy_service/
go build main.go
./main  # 服务启动在 8080端口
```

5.测试服务：

```shell
curl -XPOST -H "Content-Type: application/json" -d '{"aa":"bb"}' \
  "http://127.0.0.1:8080/cheaprpc.ahfuzhang.my_easy_service.MyEasyService.GetEchoInfo" -v
```

# 计划实现的特性

## 协议的设计

* 前期仅支持HTTP协议的请求响应。未来的版本考虑加入二进制协议
* 协议分为两层
    - 第一层包含业务中的公共信息：用户账号，版本，租户，区域，协议命令字等(可以根据自己的业务特点增减字段)
    - 第二层嵌套具体协议的请求包/响应包
* 未来通过插件化的方式，支持更多的协议通讯

### http协议

* 可以使用JSON格式通讯，其HTTP协议头需要加上`Content-Type: application/json`
    - 整个BODY支持gzip压缩(请求和响应都支持)
* 可以使用protocol buffers二进制格式同学，其http协议头需要加上`Content-Type: application/x-protobuf`
    - 第二层协议支持GZIP压缩，ZSTD压缩(业务包越大，压缩比越高)
* http协议的框架，第一版以[gin](https://github.com/gin-gonic/gin)为主
    - 未来支持 `net/http`
    - 未来支持 [fasthttp](https://github.com/valyala/fasthttp)
* ==**只支持POST请求**==
* 状态码约定：
    - 参数问题：400状态码
    - 服务器内部问题：500状态码
    - 成功：200状态码

### 二进制协议

* 协议0~0字节：代表版本号，第一个版本为1
* 协议1~4字节：代表BODY长度，网络序
* 协议5~结束：代表protocol buffers序列化后的内容
* 前期的版本考虑基于`net/tcp`来实现
* 后期考虑支持QUIC/HTTP3
* 后期考虑底层使用[netpoll](https://github.com/cloudwego/netpoll)组件

### 返回格式

* 第一层协议的返回，仅仅只是框架层面的状态码，与业务无关
    - 基本的字段包括：
        * 请求端的IP端口
        * http协议的header
        * http协议的path, URI等
        * 协议版本
        * 压缩选项
        * 加密选项
        * 数据序列化模式：json, pb, yaml, xml等
        * 请求的命令字(如果是HTTP协议，就是请求的PATH)
* 第二层的响应格式里，必须加上`code`和`msg`两个字段，用于表示业务层面的错误信息。

## 类型的设计

*
服务接口/请求格式/响应格式，以及yaml,json,gorm等用到的类型——全部采用[proto3](https://developers.google.com/protocol-buffers/docs/proto3)
的语法来定义
* 数据序列化方式，支持JSON、protocol buffers二进制格式
* 使用gogoproto作为protocol buffers的序列化库
* 前期先使用标准库的`text/json`作为json序列化的库（后期换为[sonic](https://github.com/bytedance/sonic)）

## 统一proto3为IDL，根据其语法生成代码

* 请求格式，响应格式，通过proto3语法来定义
* JSON， yaml， GORM字段名等，也通过proto3来定义
    - 通过gogoproto的extension的语法来生成struct中对应的语法
    - 具体的[扩展tag](https://github.com/gogo/protobuf/blob/master/extensions.md#more-serialization-formats)
      请看：`jsontag` 和 `moretags`
* service及其method，通过proto3语法来定义
    - 主要生成service对应的interface类型。具体的interface的服务器和客户端实现，使用代码生成来生成主要框架，具体函数实现由具体的业务开发者提供
    - 服务的命令字为: `${namespace}.${app}.${service}.${method}`这样的路径
    - 可以使用method的extension来重定义访问路径, 例如：`/namespace/app/service/method`这样的格式来请求一个service的method

## context管理

* 整个处理链中，所有函数的第一个参数都写成`ctx context.Context`
* gin框架使用了*gin.Context，但是传递到业务函数仍然是 context.Context。同样，就算底层换成了别的网络框架，也仍然使用标准库的Context类型
* 业务中要避免对ctx进行type assert
    - 约定几个key，通过 `ctx.Value("key")`来获取内容
    - 协议第一层的公共信息，通过`ctx.Value("key")`这种方式来获取
    - HTTP协议中的HEADER，通过`ctx.Value("key")`来获取

## 插件化体系

* 功能组件
    - 远程配置文件，通过插件化的方法来实现
    - 日志等能力，通过插件化的方式实现
    - 使用[VictoriaMetrics/metrics](https://github.com/VictoriaMetrics/metrics)组件来上报metric信息
        * 支持golang runtime的上报
        * 支持基本的主调、被调上报
        * 支持接口时延上报
        * 支持panic上报
        * 所有的监控项，都可以通过代码生成的方式，生成grafana报表对应的JSON。直接导入到grafana就可以看见监控数据。
    - 使用fastcache来支持本地cache能力

### 拦截器

* 分为4种拦截器：
    * before: 业务回调之前调用
    * after: 业务回调之后调用
    * error: 业务发生错误的时候调用
    * panic: 发生panic的时候调用

## 服务注册/服务发现

* 基于ETCD来实现
* 未来可以考虑zookeeper，腾讯polaris等组件

## 客户端代码

* 自动为对应的服务生成客户端代码
* 通过 `WithXXX()`这样的格式提供请求时候的各种选项
* 客户端的metric上报的约定：
    - 主调上报
    - 延迟分布
