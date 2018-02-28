/**
 * Autogenerated by Thrift Compiler (0.10.0)
 *
 * DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING
 *  @generated
 */
#ifndef MyTest_H
#define MyTest_H

#include <thrift/TDispatchProcessor.h>
#include <thrift/async/TConcurrentClientSyncInfo.h>
#include "MyTest_types.h"



#ifdef _WIN32
  #pragma warning( push )
  #pragma warning (disable : 4250 ) //inheriting methods via dominance 
#endif

class MyTestIf {
 public:
  virtual ~MyTestIf() {}
  virtual void Ping(std::string& _return) = 0;
  virtual void ServiceName(std::string& _return) = 0;
  virtual void ServiceMode(std::string& _return) = 0;
  virtual void SayHello(Response& _return, const std::string& yourName) = 0;
  virtual void SaveWave(Response& _return, const std::string& fileName, const std::string& wavFormat, const std::string& data) = 0;
};

class MyTestIfFactory {
 public:
  typedef MyTestIf Handler;

  virtual ~MyTestIfFactory() {}

  virtual MyTestIf* getHandler(const ::apache::thrift::TConnectionInfo& connInfo) = 0;
  virtual void releaseHandler(MyTestIf* /* handler */) = 0;
};

class MyTestIfSingletonFactory : virtual public MyTestIfFactory {
 public:
  MyTestIfSingletonFactory(const boost::shared_ptr<MyTestIf>& iface) : iface_(iface) {}
  virtual ~MyTestIfSingletonFactory() {}

  virtual MyTestIf* getHandler(const ::apache::thrift::TConnectionInfo&) {
    return iface_.get();
  }
  virtual void releaseHandler(MyTestIf* /* handler */) {}

 protected:
  boost::shared_ptr<MyTestIf> iface_;
};

class MyTestNull : virtual public MyTestIf {
 public:
  virtual ~MyTestNull() {}
  void Ping(std::string& /* _return */) {
    return;
  }
  void ServiceName(std::string& /* _return */) {
    return;
  }
  void ServiceMode(std::string& /* _return */) {
    return;
  }
  void SayHello(Response& /* _return */, const std::string& /* yourName */) {
    return;
  }
  void SaveWave(Response& /* _return */, const std::string& /* fileName */, const std::string& /* wavFormat */, const std::string& /* data */) {
    return;
  }
};


class MyTest_Ping_args {
 public:

  MyTest_Ping_args(const MyTest_Ping_args&);
  MyTest_Ping_args& operator=(const MyTest_Ping_args&);
  MyTest_Ping_args() {
  }

  virtual ~MyTest_Ping_args() throw();

  bool operator == (const MyTest_Ping_args & /* rhs */) const
  {
    return true;
  }
  bool operator != (const MyTest_Ping_args &rhs) const {
    return !(*this == rhs);
  }

  bool operator < (const MyTest_Ping_args & ) const;

  uint32_t read(::apache::thrift::protocol::TProtocol* iprot);
  uint32_t write(::apache::thrift::protocol::TProtocol* oprot) const;

};


class MyTest_Ping_pargs {
 public:


  virtual ~MyTest_Ping_pargs() throw();

  uint32_t write(::apache::thrift::protocol::TProtocol* oprot) const;

};

typedef struct _MyTest_Ping_result__isset {
  _MyTest_Ping_result__isset() : success(false) {}
  bool success :1;
} _MyTest_Ping_result__isset;

class MyTest_Ping_result {
 public:

  MyTest_Ping_result(const MyTest_Ping_result&);
  MyTest_Ping_result& operator=(const MyTest_Ping_result&);
  MyTest_Ping_result() : success() {
  }

  virtual ~MyTest_Ping_result() throw();
  std::string success;

  _MyTest_Ping_result__isset __isset;

  void __set_success(const std::string& val);

  bool operator == (const MyTest_Ping_result & rhs) const
  {
    if (!(success == rhs.success))
      return false;
    return true;
  }
  bool operator != (const MyTest_Ping_result &rhs) const {
    return !(*this == rhs);
  }

  bool operator < (const MyTest_Ping_result & ) const;

  uint32_t read(::apache::thrift::protocol::TProtocol* iprot);
  uint32_t write(::apache::thrift::protocol::TProtocol* oprot) const;

};

typedef struct _MyTest_Ping_presult__isset {
  _MyTest_Ping_presult__isset() : success(false) {}
  bool success :1;
} _MyTest_Ping_presult__isset;

class MyTest_Ping_presult {
 public:


  virtual ~MyTest_Ping_presult() throw();
  std::string* success;

  _MyTest_Ping_presult__isset __isset;

  uint32_t read(::apache::thrift::protocol::TProtocol* iprot);

};


class MyTest_ServiceName_args {
 public:

  MyTest_ServiceName_args(const MyTest_ServiceName_args&);
  MyTest_ServiceName_args& operator=(const MyTest_ServiceName_args&);
  MyTest_ServiceName_args() {
  }

  virtual ~MyTest_ServiceName_args() throw();

  bool operator == (const MyTest_ServiceName_args & /* rhs */) const
  {
    return true;
  }
  bool operator != (const MyTest_ServiceName_args &rhs) const {
    return !(*this == rhs);
  }

  bool operator < (const MyTest_ServiceName_args & ) const;

  uint32_t read(::apache::thrift::protocol::TProtocol* iprot);
  uint32_t write(::apache::thrift::protocol::TProtocol* oprot) const;

};


class MyTest_ServiceName_pargs {
 public:


  virtual ~MyTest_ServiceName_pargs() throw();

  uint32_t write(::apache::thrift::protocol::TProtocol* oprot) const;

};

typedef struct _MyTest_ServiceName_result__isset {
  _MyTest_ServiceName_result__isset() : success(false) {}
  bool success :1;
} _MyTest_ServiceName_result__isset;

class MyTest_ServiceName_result {
 public:

  MyTest_ServiceName_result(const MyTest_ServiceName_result&);
  MyTest_ServiceName_result& operator=(const MyTest_ServiceName_result&);
  MyTest_ServiceName_result() : success() {
  }

  virtual ~MyTest_ServiceName_result() throw();
  std::string success;

  _MyTest_ServiceName_result__isset __isset;

  void __set_success(const std::string& val);

  bool operator == (const MyTest_ServiceName_result & rhs) const
  {
    if (!(success == rhs.success))
      return false;
    return true;
  }
  bool operator != (const MyTest_ServiceName_result &rhs) const {
    return !(*this == rhs);
  }

  bool operator < (const MyTest_ServiceName_result & ) const;

  uint32_t read(::apache::thrift::protocol::TProtocol* iprot);
  uint32_t write(::apache::thrift::protocol::TProtocol* oprot) const;

};

typedef struct _MyTest_ServiceName_presult__isset {
  _MyTest_ServiceName_presult__isset() : success(false) {}
  bool success :1;
} _MyTest_ServiceName_presult__isset;

class MyTest_ServiceName_presult {
 public:


  virtual ~MyTest_ServiceName_presult() throw();
  std::string* success;

  _MyTest_ServiceName_presult__isset __isset;

  uint32_t read(::apache::thrift::protocol::TProtocol* iprot);

};


class MyTest_ServiceMode_args {
 public:

  MyTest_ServiceMode_args(const MyTest_ServiceMode_args&);
  MyTest_ServiceMode_args& operator=(const MyTest_ServiceMode_args&);
  MyTest_ServiceMode_args() {
  }

  virtual ~MyTest_ServiceMode_args() throw();

  bool operator == (const MyTest_ServiceMode_args & /* rhs */) const
  {
    return true;
  }
  bool operator != (const MyTest_ServiceMode_args &rhs) const {
    return !(*this == rhs);
  }

  bool operator < (const MyTest_ServiceMode_args & ) const;

  uint32_t read(::apache::thrift::protocol::TProtocol* iprot);
  uint32_t write(::apache::thrift::protocol::TProtocol* oprot) const;

};


class MyTest_ServiceMode_pargs {
 public:


  virtual ~MyTest_ServiceMode_pargs() throw();

  uint32_t write(::apache::thrift::protocol::TProtocol* oprot) const;

};

typedef struct _MyTest_ServiceMode_result__isset {
  _MyTest_ServiceMode_result__isset() : success(false) {}
  bool success :1;
} _MyTest_ServiceMode_result__isset;

class MyTest_ServiceMode_result {
 public:

  MyTest_ServiceMode_result(const MyTest_ServiceMode_result&);
  MyTest_ServiceMode_result& operator=(const MyTest_ServiceMode_result&);
  MyTest_ServiceMode_result() : success() {
  }

  virtual ~MyTest_ServiceMode_result() throw();
  std::string success;

  _MyTest_ServiceMode_result__isset __isset;

  void __set_success(const std::string& val);

  bool operator == (const MyTest_ServiceMode_result & rhs) const
  {
    if (!(success == rhs.success))
      return false;
    return true;
  }
  bool operator != (const MyTest_ServiceMode_result &rhs) const {
    return !(*this == rhs);
  }

  bool operator < (const MyTest_ServiceMode_result & ) const;

  uint32_t read(::apache::thrift::protocol::TProtocol* iprot);
  uint32_t write(::apache::thrift::protocol::TProtocol* oprot) const;

};

typedef struct _MyTest_ServiceMode_presult__isset {
  _MyTest_ServiceMode_presult__isset() : success(false) {}
  bool success :1;
} _MyTest_ServiceMode_presult__isset;

class MyTest_ServiceMode_presult {
 public:


  virtual ~MyTest_ServiceMode_presult() throw();
  std::string* success;

  _MyTest_ServiceMode_presult__isset __isset;

  uint32_t read(::apache::thrift::protocol::TProtocol* iprot);

};

typedef struct _MyTest_SayHello_args__isset {
  _MyTest_SayHello_args__isset() : yourName(false) {}
  bool yourName :1;
} _MyTest_SayHello_args__isset;

class MyTest_SayHello_args {
 public:

  MyTest_SayHello_args(const MyTest_SayHello_args&);
  MyTest_SayHello_args& operator=(const MyTest_SayHello_args&);
  MyTest_SayHello_args() : yourName() {
  }

  virtual ~MyTest_SayHello_args() throw();
  std::string yourName;

  _MyTest_SayHello_args__isset __isset;

  void __set_yourName(const std::string& val);

  bool operator == (const MyTest_SayHello_args & rhs) const
  {
    if (!(yourName == rhs.yourName))
      return false;
    return true;
  }
  bool operator != (const MyTest_SayHello_args &rhs) const {
    return !(*this == rhs);
  }

  bool operator < (const MyTest_SayHello_args & ) const;

  uint32_t read(::apache::thrift::protocol::TProtocol* iprot);
  uint32_t write(::apache::thrift::protocol::TProtocol* oprot) const;

};


class MyTest_SayHello_pargs {
 public:


  virtual ~MyTest_SayHello_pargs() throw();
  const std::string* yourName;

  uint32_t write(::apache::thrift::protocol::TProtocol* oprot) const;

};

typedef struct _MyTest_SayHello_result__isset {
  _MyTest_SayHello_result__isset() : success(false) {}
  bool success :1;
} _MyTest_SayHello_result__isset;

class MyTest_SayHello_result {
 public:

  MyTest_SayHello_result(const MyTest_SayHello_result&);
  MyTest_SayHello_result& operator=(const MyTest_SayHello_result&);
  MyTest_SayHello_result() {
  }

  virtual ~MyTest_SayHello_result() throw();
  Response success;
  ServiceException e;

  _MyTest_SayHello_result__isset __isset;

  void __set_success(const Response& val);

  void __set_e(const ServiceException& val);

  bool operator == (const MyTest_SayHello_result & rhs) const
  {
    if (!(success == rhs.success))
      return false;
    if (!(e == rhs.e))
      return false;
    return true;
  }
  bool operator != (const MyTest_SayHello_result &rhs) const {
    return !(*this == rhs);
  }

  bool operator < (const MyTest_SayHello_result & ) const;

  uint32_t read(::apache::thrift::protocol::TProtocol* iprot);
  uint32_t write(::apache::thrift::protocol::TProtocol* oprot) const;

};

typedef struct _MyTest_SayHello_presult__isset {
  _MyTest_SayHello_presult__isset() : success(false) {}
  bool success :1;
} _MyTest_SayHello_presult__isset;

class MyTest_SayHello_presult {
 public:


  virtual ~MyTest_SayHello_presult() throw();
  Response* success;
  ServiceException e;

  _MyTest_SayHello_presult__isset __isset;

  uint32_t read(::apache::thrift::protocol::TProtocol* iprot);

};

typedef struct _MyTest_SaveWave_args__isset {
  _MyTest_SaveWave_args__isset() : fileName(false), wavFormat(false), data(false) {}
  bool fileName :1;
  bool wavFormat :1;
  bool data :1;
} _MyTest_SaveWave_args__isset;

class MyTest_SaveWave_args {
 public:

  MyTest_SaveWave_args(const MyTest_SaveWave_args&);
  MyTest_SaveWave_args& operator=(const MyTest_SaveWave_args&);
  MyTest_SaveWave_args() : fileName(), wavFormat(), data() {
  }

  virtual ~MyTest_SaveWave_args() throw();
  std::string fileName;
  std::string wavFormat;
  std::string data;

  _MyTest_SaveWave_args__isset __isset;

  void __set_fileName(const std::string& val);

  void __set_wavFormat(const std::string& val);

  void __set_data(const std::string& val);

  bool operator == (const MyTest_SaveWave_args & rhs) const
  {
    if (!(fileName == rhs.fileName))
      return false;
    if (!(wavFormat == rhs.wavFormat))
      return false;
    if (!(data == rhs.data))
      return false;
    return true;
  }
  bool operator != (const MyTest_SaveWave_args &rhs) const {
    return !(*this == rhs);
  }

  bool operator < (const MyTest_SaveWave_args & ) const;

  uint32_t read(::apache::thrift::protocol::TProtocol* iprot);
  uint32_t write(::apache::thrift::protocol::TProtocol* oprot) const;

};


class MyTest_SaveWave_pargs {
 public:


  virtual ~MyTest_SaveWave_pargs() throw();
  const std::string* fileName;
  const std::string* wavFormat;
  const std::string* data;

  uint32_t write(::apache::thrift::protocol::TProtocol* oprot) const;

};

typedef struct _MyTest_SaveWave_result__isset {
  _MyTest_SaveWave_result__isset() : success(false) {}
  bool success :1;
} _MyTest_SaveWave_result__isset;

class MyTest_SaveWave_result {
 public:

  MyTest_SaveWave_result(const MyTest_SaveWave_result&);
  MyTest_SaveWave_result& operator=(const MyTest_SaveWave_result&);
  MyTest_SaveWave_result() {
  }

  virtual ~MyTest_SaveWave_result() throw();
  Response success;
  ServiceException e;

  _MyTest_SaveWave_result__isset __isset;

  void __set_success(const Response& val);

  void __set_e(const ServiceException& val);

  bool operator == (const MyTest_SaveWave_result & rhs) const
  {
    if (!(success == rhs.success))
      return false;
    if (!(e == rhs.e))
      return false;
    return true;
  }
  bool operator != (const MyTest_SaveWave_result &rhs) const {
    return !(*this == rhs);
  }

  bool operator < (const MyTest_SaveWave_result & ) const;

  uint32_t read(::apache::thrift::protocol::TProtocol* iprot);
  uint32_t write(::apache::thrift::protocol::TProtocol* oprot) const;

};

typedef struct _MyTest_SaveWave_presult__isset {
  _MyTest_SaveWave_presult__isset() : success(false) {}
  bool success :1;
} _MyTest_SaveWave_presult__isset;

class MyTest_SaveWave_presult {
 public:


  virtual ~MyTest_SaveWave_presult() throw();
  Response* success;
  ServiceException e;

  _MyTest_SaveWave_presult__isset __isset;

  uint32_t read(::apache::thrift::protocol::TProtocol* iprot);

};

class MyTestClient : virtual public MyTestIf {
 public:
  MyTestClient(boost::shared_ptr< ::apache::thrift::protocol::TProtocol> prot) {
    setProtocol(prot);
  }
  MyTestClient(boost::shared_ptr< ::apache::thrift::protocol::TProtocol> iprot, boost::shared_ptr< ::apache::thrift::protocol::TProtocol> oprot) {
    setProtocol(iprot,oprot);
  }
 private:
  void setProtocol(boost::shared_ptr< ::apache::thrift::protocol::TProtocol> prot) {
  setProtocol(prot,prot);
  }
  void setProtocol(boost::shared_ptr< ::apache::thrift::protocol::TProtocol> iprot, boost::shared_ptr< ::apache::thrift::protocol::TProtocol> oprot) {
    piprot_=iprot;
    poprot_=oprot;
    iprot_ = iprot.get();
    oprot_ = oprot.get();
  }
 public:
  boost::shared_ptr< ::apache::thrift::protocol::TProtocol> getInputProtocol() {
    return piprot_;
  }
  boost::shared_ptr< ::apache::thrift::protocol::TProtocol> getOutputProtocol() {
    return poprot_;
  }
  void Ping(std::string& _return);
  void send_Ping();
  void recv_Ping(std::string& _return);
  void ServiceName(std::string& _return);
  void send_ServiceName();
  void recv_ServiceName(std::string& _return);
  void ServiceMode(std::string& _return);
  void send_ServiceMode();
  void recv_ServiceMode(std::string& _return);
  void SayHello(Response& _return, const std::string& yourName);
  void send_SayHello(const std::string& yourName);
  void recv_SayHello(Response& _return);
  void SaveWave(Response& _return, const std::string& fileName, const std::string& wavFormat, const std::string& data);
  void send_SaveWave(const std::string& fileName, const std::string& wavFormat, const std::string& data);
  void recv_SaveWave(Response& _return);
 protected:
  boost::shared_ptr< ::apache::thrift::protocol::TProtocol> piprot_;
  boost::shared_ptr< ::apache::thrift::protocol::TProtocol> poprot_;
  ::apache::thrift::protocol::TProtocol* iprot_;
  ::apache::thrift::protocol::TProtocol* oprot_;
};

class MyTestProcessor : public ::apache::thrift::TDispatchProcessor {
 protected:
  boost::shared_ptr<MyTestIf> iface_;
  virtual bool dispatchCall(::apache::thrift::protocol::TProtocol* iprot, ::apache::thrift::protocol::TProtocol* oprot, const std::string& fname, int32_t seqid, void* callContext);
 private:
  typedef  void (MyTestProcessor::*ProcessFunction)(int32_t, ::apache::thrift::protocol::TProtocol*, ::apache::thrift::protocol::TProtocol*, void*);
  typedef std::map<std::string, ProcessFunction> ProcessMap;
  ProcessMap processMap_;
  void process_Ping(int32_t seqid, ::apache::thrift::protocol::TProtocol* iprot, ::apache::thrift::protocol::TProtocol* oprot, void* callContext);
  void process_ServiceName(int32_t seqid, ::apache::thrift::protocol::TProtocol* iprot, ::apache::thrift::protocol::TProtocol* oprot, void* callContext);
  void process_ServiceMode(int32_t seqid, ::apache::thrift::protocol::TProtocol* iprot, ::apache::thrift::protocol::TProtocol* oprot, void* callContext);
  void process_SayHello(int32_t seqid, ::apache::thrift::protocol::TProtocol* iprot, ::apache::thrift::protocol::TProtocol* oprot, void* callContext);
  void process_SaveWave(int32_t seqid, ::apache::thrift::protocol::TProtocol* iprot, ::apache::thrift::protocol::TProtocol* oprot, void* callContext);
 public:
  MyTestProcessor(boost::shared_ptr<MyTestIf> iface) :
    iface_(iface) {
    processMap_["Ping"] = &MyTestProcessor::process_Ping;
    processMap_["ServiceName"] = &MyTestProcessor::process_ServiceName;
    processMap_["ServiceMode"] = &MyTestProcessor::process_ServiceMode;
    processMap_["SayHello"] = &MyTestProcessor::process_SayHello;
    processMap_["SaveWave"] = &MyTestProcessor::process_SaveWave;
  }

  virtual ~MyTestProcessor() {}
};

class MyTestProcessorFactory : public ::apache::thrift::TProcessorFactory {
 public:
  MyTestProcessorFactory(const ::boost::shared_ptr< MyTestIfFactory >& handlerFactory) :
      handlerFactory_(handlerFactory) {}

  ::boost::shared_ptr< ::apache::thrift::TProcessor > getProcessor(const ::apache::thrift::TConnectionInfo& connInfo);

 protected:
  ::boost::shared_ptr< MyTestIfFactory > handlerFactory_;
};

class MyTestMultiface : virtual public MyTestIf {
 public:
  MyTestMultiface(std::vector<boost::shared_ptr<MyTestIf> >& ifaces) : ifaces_(ifaces) {
  }
  virtual ~MyTestMultiface() {}
 protected:
  std::vector<boost::shared_ptr<MyTestIf> > ifaces_;
  MyTestMultiface() {}
  void add(boost::shared_ptr<MyTestIf> iface) {
    ifaces_.push_back(iface);
  }
 public:
  void Ping(std::string& _return) {
    size_t sz = ifaces_.size();
    size_t i = 0;
    for (; i < (sz - 1); ++i) {
      ifaces_[i]->Ping(_return);
    }
    ifaces_[i]->Ping(_return);
    return;
  }

  void ServiceName(std::string& _return) {
    size_t sz = ifaces_.size();
    size_t i = 0;
    for (; i < (sz - 1); ++i) {
      ifaces_[i]->ServiceName(_return);
    }
    ifaces_[i]->ServiceName(_return);
    return;
  }

  void ServiceMode(std::string& _return) {
    size_t sz = ifaces_.size();
    size_t i = 0;
    for (; i < (sz - 1); ++i) {
      ifaces_[i]->ServiceMode(_return);
    }
    ifaces_[i]->ServiceMode(_return);
    return;
  }

  void SayHello(Response& _return, const std::string& yourName) {
    size_t sz = ifaces_.size();
    size_t i = 0;
    for (; i < (sz - 1); ++i) {
      ifaces_[i]->SayHello(_return, yourName);
    }
    ifaces_[i]->SayHello(_return, yourName);
    return;
  }

  void SaveWave(Response& _return, const std::string& fileName, const std::string& wavFormat, const std::string& data) {
    size_t sz = ifaces_.size();
    size_t i = 0;
    for (; i < (sz - 1); ++i) {
      ifaces_[i]->SaveWave(_return, fileName, wavFormat, data);
    }
    ifaces_[i]->SaveWave(_return, fileName, wavFormat, data);
    return;
  }

};

// The 'concurrent' client is a thread safe client that correctly handles
// out of order responses.  It is slower than the regular client, so should
// only be used when you need to share a connection among multiple threads
class MyTestConcurrentClient : virtual public MyTestIf {
 public:
  MyTestConcurrentClient(boost::shared_ptr< ::apache::thrift::protocol::TProtocol> prot) {
    setProtocol(prot);
  }
  MyTestConcurrentClient(boost::shared_ptr< ::apache::thrift::protocol::TProtocol> iprot, boost::shared_ptr< ::apache::thrift::protocol::TProtocol> oprot) {
    setProtocol(iprot,oprot);
  }
 private:
  void setProtocol(boost::shared_ptr< ::apache::thrift::protocol::TProtocol> prot) {
  setProtocol(prot,prot);
  }
  void setProtocol(boost::shared_ptr< ::apache::thrift::protocol::TProtocol> iprot, boost::shared_ptr< ::apache::thrift::protocol::TProtocol> oprot) {
    piprot_=iprot;
    poprot_=oprot;
    iprot_ = iprot.get();
    oprot_ = oprot.get();
  }
 public:
  boost::shared_ptr< ::apache::thrift::protocol::TProtocol> getInputProtocol() {
    return piprot_;
  }
  boost::shared_ptr< ::apache::thrift::protocol::TProtocol> getOutputProtocol() {
    return poprot_;
  }
  void Ping(std::string& _return);
  int32_t send_Ping();
  void recv_Ping(std::string& _return, const int32_t seqid);
  void ServiceName(std::string& _return);
  int32_t send_ServiceName();
  void recv_ServiceName(std::string& _return, const int32_t seqid);
  void ServiceMode(std::string& _return);
  int32_t send_ServiceMode();
  void recv_ServiceMode(std::string& _return, const int32_t seqid);
  void SayHello(Response& _return, const std::string& yourName);
  int32_t send_SayHello(const std::string& yourName);
  void recv_SayHello(Response& _return, const int32_t seqid);
  void SaveWave(Response& _return, const std::string& fileName, const std::string& wavFormat, const std::string& data);
  int32_t send_SaveWave(const std::string& fileName, const std::string& wavFormat, const std::string& data);
  void recv_SaveWave(Response& _return, const int32_t seqid);
 protected:
  boost::shared_ptr< ::apache::thrift::protocol::TProtocol> piprot_;
  boost::shared_ptr< ::apache::thrift::protocol::TProtocol> poprot_;
  ::apache::thrift::protocol::TProtocol* iprot_;
  ::apache::thrift::protocol::TProtocol* oprot_;
  ::apache::thrift::async::TConcurrentClientSyncInfo sync_;
};

#ifdef _WIN32
  #pragma warning( pop )
#endif



#endif
