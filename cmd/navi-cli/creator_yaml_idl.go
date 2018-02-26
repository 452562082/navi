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
  http_port: 8081
  max_conn_num: 10
  grpc_service_name: {{.ServiceName}}
  grpc_service_host: 127.0.0.1
  grpc_service_port: 50051
  thrift_service_name: {{.ServiceName}}
  thrift_service_host: 127.0.0.1
  thrift_service_port: 50052
  service_version_mode: prod
  service_version: 1.0
  ip_filter_path: navi/ipfilter
  zookeeper_servers_addr: 127.0.0.1:2181
  zookeeper_url_service_path: /navi/service
  zookeeper_http_service_path: /navi/httpservice
  zookeeper_rpc_service_path: /navi/rpcservice

urlmapping:
  - POST /hello SayHello
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

# 这个结构体定义了服务调用者的请求信息
/*struct Request {
    # 传递的参数信息，使用格式进行表示
    1:required binary paramJSON;
    # 服务调用者请求的服务名，使用serviceName属性进行传递
    2:required string serviceName
}*/

# 这个结构体，定义了服务提供者的返回信息
struct Response {
    # RESCODE 是处理状态代码，是一个枚举类型。例如RESCODE._200表示处理成功
    1:required RESCODE responseCode;
    # 返回的处理结果，同样使用JSON格式进行描述
    2:required string responseJSON;
}

# 异常描述定义，当服务提供者处理过程出现异常时，向服务调用者返回
exception ServiceException {
    # EXCCODE 是异常代码，也是一个枚举类型。
    # 例如EXCCODE.PARAMNOTFOUND表示需要的请求参数没有找到
    1:required EXCCODE exceptionCode;
    # 异常的描述信息，使用字符串进行描述
    2:required string exceptionMess;
}

# 这个枚举结构，描述各种服务提供者的响应代码
enum RESCODE {
    SUCCESS = 200;
	FORBIDDEN = 403;
	NOTFOUND = 404;
    BADGATEWAY = 502;
}

# 这个枚举结构，描述各种服务提供者的异常种类
enum EXCCODE {
    PARAMNOTFOUND = 2001;
    SERVICENOTFOUND = 2002;
}

# 这是经过泛化后的Apache Thrift接口
service {{.ServiceName}} {
        string Ping(),

        string ServiceName(),

        string ServiceMode(),

        Response SayHello(1:string yourName) throws (1:required ServiceException e)
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

message SayHelloRequest {
    string yourName = 1;
}

message SayHelloResponse {
    string message = 1;
}

service {{.ServiceName}} {
    rpc sayHello (SayHelloRequest) returns (SayHelloResponse) {}
}
`,
	)
}
