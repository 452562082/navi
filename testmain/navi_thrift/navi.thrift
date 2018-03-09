/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

# Thrift Tutorial
# Mark Slee (mcslee@facebook.com)
#
# This file aims to teach you how to use Thrift, in a .thrift file. Neato. The
# first thing to notice is that .thrift files support standard shell comments.
# This lets you make your thrift file executable and include your Thrift build
# step on the top line. And you can place comments like this anywhere you like.
#
# Before running this file, you will need to have installed the thrift compiler
# into /usr/local/bin.

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
namespace go navi

# 这个结构体定义了服务调用者的请求信息
struct Request {
    # 传递的参数信息，使用格式进行表示
    1:required binary paramJSON;
    # 服务调用者请求的服务名，使用serviceName属性进行传递
    2:required string serviceName
}

# 这个结构体，定义了服务提供者的返回信息
struct Reponse {
    # RESCODE 是处理状态代码，是一个枚举类型。例如RESCODE._200表示处理成功
    1:required  RESCODE responeCode;
    # 返回的处理结果，同样使用JSON格式进行描述
    2:required  binary responseJSON;
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
    _200=200;
    _500=500;
    _400=400;
}

# 这个枚举结构，描述各种服务提供者的异常种类
enum EXCCODE {
    PARAMNOTFOUND = 2001;
    SERVICENOTFOUND = 2002;
}

# 这是经过泛化后的Apache Thrift接口
service naviService {
        string Ping(),

        string ServiceName(),

        string ServiceType(),

        Reponse send(1:required Request request) throws (1:required ServiceException e)
}