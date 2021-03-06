package navicli

import (
	"os"
	"strings"
)

func (c *Creator) createServiceYaml(serviceRootPath, serviceName, configFileName string) {
	type serviceYamlValues struct {
		ServiceRoot string
		ServiceName string
	}
	if _, err := os.Stat(serviceRootPath + "/" + configFileName + ".yaml"); err == nil {
		return
	}
	writeFileWithTemplate(
		serviceRootPath+"/"+configFileName+".yaml",
		serviceYamlValues{ServiceRoot: serviceRootPath, ServiceName: serviceName},
		`config:
  environment: development
  service_root_path: {{.ServiceRoot}}
  turbo_log_path: log
  http_host: 127.0.0.1:8081
  is_docker: true
  http_port: 8081
  max_conn_num: 10
  grpc_service_name: {{.ServiceName}}
  thrift_service_name: {{.ServiceName}}
  service_version_mode: dev
  service_version: 1.0
  ip_filter_path: navi/ipfilter
  zookeeper_servers_addr: 127.0.0.1:2181
  zookeeper_url_service_path: /navi/service
  zookeeper_http_service_path: /navi/httpservice
  zookeeper_rpc_service_path: /navi/rpcservice
  jaeger_addr: 127.0.0.1:6831
  prometheus_target_host: 192.168.1.16
  prometheus_target_port: 9100

urlmapping:
  - GET /ping Ping
  - GET /servicename ServiceName
  - GET /servicemode ServiceMode
  - POST /hello SayHello
  - POST /savewave SaveWave

log:
    # 开启日志文件
   enable: true
    # 日志文件路径
   file: /navi/logs/navi.log
    # 日志等级
    # Fatal:0, Error:1, Alert:2, Warn:3
    # Notice:4, Info:5, Debug:6, Trace:7
   level: 6
    # 异步日志
   async: false
    # 日志等级配色
   coloured: true
    # 日志显示行号
   show_lines: true
    # 日志最大行数
   maxlines: 5000000
    # 日志最大容量
   maxsize: 536870912
    # 日志隔天回滚
   daily: true
    # 保存日志最大天数
   maxdays: 15
`)
}

func (c *Creator) createThrift(serviceName string) {
	type thriftValues struct {
		ServiceName string
	}
	nameLower := strings.ToLower(serviceName)
	writeFileWithTemplate(
		c.c.ServiceRootPathAbsolute()+"/"+nameLower+".thrift",
		thriftValues{ServiceName: serviceName},
		`
/**
 * The first thing to know about are types. The available types in Thrift are:
 *
 *  bool        Boolean, one byte
 *  i8 (byte)   Signed 8-bit integer
 *  i16         Signed 16-bit integer
 *  i32         Signed 32-bit integer
 *  i64         Signed 64-bit integer
 *  double      64-bit floating point value
 *  string      String
 *  binary      Blob (byte array)
 *  map<t1,t2>  Map from one type to another
 *  list<t1>    Ordered list of one type
 *  set<t1>     Set of unique elements of one type
 *
 * Did you also notice that Thrift supports C style comments?
 */

namespace go gen

# TODO(暂时不使用Request结构体，后续用到分布式追踪，再修改)
# 这个结构体定义了服务调用者的请求信息
/*struct Request {
    # 传递的参数信息，使用格式进行表示
    1:required binary paramJSON;
    # 服务调用者请求的服务名，使用serviceName属性进行传递
    2:required string serviceName
}*/

# 这个结构体，定义了服务提供者的返回信息
struct Response {
    # RESCODE 是处理状态代码，是一个int32型, 具体状态码参考文档;
    1:required i32 responseCode;
    # 返回的处理结果，同样使用JSON格式进行描述
    2:required string responseJSON;
}

# 这是经过泛化后的Apache Thrift接口
service {{.ServiceName}} {

		# rpc server必须实现的接口，返回字符串 "pong" 即可
		string Ping(),

		# rpc server必须实现的接口，返回服务名称，为首字母大写的驼峰格式，例如 "AsvService"
		string ServiceName(),

		# rpc server必须实现的接口，说明该server是以什么模式运行，分为dev和prod；dev为开发版本，prod为生产版本
		string ServiceMode(),

		Response SayHello(1:string yourName)

		Response SaveWave(1:string fileName, 2:string wavFormat, 3:binary data)
}
`,
	)
}

func (c *Creator) createProto(serviceName string) {
	type protoValues struct {
		ServiceName string
	}
	nameLower := strings.ToLower(serviceName)
	writeFileWithTemplate(
		c.c.ServiceRootPathAbsolute()+"/"+nameLower+".proto",
		protoValues{ServiceName: serviceName},
		`syntax = "proto3";
package proto;

message PingRequest {}

//message PingResponse {
//    string pong = 1;
//}

message ServiceNameRequest {}

//message ServiceNameResponse {
//    string service_name = 1;
//}

message ServiceModeRequest {}

//message ServiceModeResponse {
//    string service_mode = 1;
//}

message SayHelloRequest {
	string your_name = 1;
}

//message SayHelloResponse {
//    string message = 1;
//}

message SaveWaveRequest {
	string file_name = 1;
	string wav_format = 2;
	int32 data = 3;
}

// 这个结构体，定义了服务提供者的返回信息
message Response {
	// RESCODE 是处理状态代码，是一个int32型, 具体状态码参考文档;
	int32 responseCode = 1;
	// 返回的处理结果，同样使用JSON格式进行描述
	string responseJSON = 2;
}

service {{.ServiceName}} {
	// rpc server必须实现的接口，返回字符串 "pong" 即可
	rpc Ping(PingRequest) returns (Response) {}

	// rpc server必须实现的接口，返回服务名称，为首字母大写的驼峰格式，例如 "AsvService"
	rpc ServiceName(ServiceNameRequest) returns (Response) {}

	// rpc server必须实现的接口，说明该server是以什么模式运行，分为dev和prod；dev为开发版本，prod为生产版本
	rpc ServiceMode(ServiceModeRequest) returns (Response) {}

	rpc SayHello (SayHelloRequest) returns (Response) {}

	rpc SaveWave(SaveWaveRequest) returns (Response) {}
}
`,
	)
}
