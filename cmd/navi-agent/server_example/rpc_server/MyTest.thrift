namespace go gen

# 这个结构体，定义了服务提供者的返回信息
struct Response {
    # RESCODE 是处理状态代码，是一个int32型, 具体状态码参考文档;
    1:required i32 responseCode;
    # 返回的处理结果，同样使用JSON格式进行描述
    2:required string responseJSON;
}

# 异常描述定义，当服务提供者处理过程出现异常时，向服务调用者返回
exception ServiceException {
    # EXCCODE 是异常代码，也是一个int32型。
    1:required i32 exceptionCode;
    # 异常的描述信息，使用字符串进行描述
    2:required string exceptionMeg;
}

# 这是经过泛化后的Apache Thrift接口
service MyTest {

		# rpc server必须实现的接口，返回字符串 "pong" 即可
        string Ping(),

		# rpc server必须实现的接口，返回服务名称，为首字母大写的驼峰格式，例如 "AsvService"
        string ServiceName(),

		# rpc server必须实现的接口，说明该server是以什么模式运行，分为dev和prod；dev为开发版本，prod为生产版本
        string ServiceMode(),

        Response SayHello(1:string yourName) throws (1:required ServiceException e)

        Response SaveWave(1:string fileName, 2:string wavFormat, 3:binary data) throws (1:required ServiceException e)
}