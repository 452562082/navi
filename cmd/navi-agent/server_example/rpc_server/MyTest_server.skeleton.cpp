// This autogenerated skeleton file illustrates how to build a server.
// You should copy it to another filename to avoid overwriting it.

#include "MyTest.h"
#include <stdlib.h>
#include <thrift/protocol/TBinaryProtocol.h>
#include <thrift/server/TSimpleServer.h>
#include <thrift/concurrency/ThreadManager.h>
#include <thrift/concurrency/PosixThreadFactory.h>
#include <thrift/server/TThreadPoolServer.h>
#include <thrift/server/TThreadedServer.h>
#include <thrift/server/TNonblockingServer.h>
#include <thrift/transport/TServerSocket.h>
#include <thrift/transport/TBufferTransports.h>
#include "waveIO.h"

using namespace ::apache::thrift;
using namespace ::apache::thrift::protocol;
using namespace ::apache::thrift::transport;
using namespace ::apache::thrift::server;
using namespace ::apache::thrift::concurrency;

using boost::shared_ptr;

class MyTestHandler : virtual public MyTestIf {
 public:
  MyTestHandler() {
    // Your initialization goes here
  }

  void Ping(std::string& _return) {
    // Your implementation goes here
    _return = "pong";
  }

  void ServiceName(std::string& _return) {
    // Your implementation goes here
     _return = "MyTest";
  }

  void ServiceMode(std::string& _return) {
    // Your implementation goes here
     _return = "dev";
  }

  void SayHello(Response& _return, const std::string& yourName) {
    // Your implementation goes here
   _return.__set_responseCode(200);
   _return.__set_responseJSON("{\"name\": \""+yourName+"\"}");
  }

  void SaveWave(Response& _return, const std::string& fileName, const std::string& wavFormat, const std::string& data) {
    // Your implementation goes here
    int bufLen = data.size() / 2;
    short* buf = new short[bufLen];
    for(int i = 0, j = 0;i < data.size() && j < bufLen;) {
        buf[i] = (short)(data[i] & 0xff);
        buf[j] |= (short)((data[i + 1] << 8) & 0xff00);
        i += 2;
        j++;
    }

    bool flag = WaveSave(fileName,buf,bufLen);
    printf("status: %d",flag);

    _return.__set_responseCode(200);
    _return.__set_responseJSON("{\"file_name\": \""+fileName+"\", \"wav_format\": \"" +wavFormat+"\", \"data\": \"" +data+"\"}");
  }

};

int main(int argc, char **argv) {

    int port = 9191;
    if (argc >= 2)   port = atoi(argv[1]);

      shared_ptr<MyTestHandler> handler(new MyTestHandler());
      shared_ptr<TProcessor> processor(new MyTestProcessor(handler));
      shared_ptr<TProtocolFactory> protocolFactory(new TBinaryProtocolFactory());
      shared_ptr<ThreadManager> threadManager = ThreadManager::newSimpleThreadManager(16);
      shared_ptr<PosixThreadFactory> threadFactory = shared_ptr<PosixThreadFactory > (new PosixThreadFactory());
      threadManager->threadFactory(threadFactory);
      threadManager->start();

      TNonblockingServer server(processor, protocolFactory, port, threadManager);
      try{
          server.serve();
      }
      catch(TException e) {
            printf("Server.serve() failed: %s\n", e.what());
            exit(-1);
      }
      return 0;
}

