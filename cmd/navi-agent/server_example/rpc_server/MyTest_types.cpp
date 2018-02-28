/**
 * Autogenerated by Thrift Compiler (0.10.0)
 *
 * DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING
 *  @generated
 */
#include "MyTest_types.h"

#include <algorithm>
#include <ostream>

#include <thrift/TToString.h>




Response::~Response() throw() {
}


void Response::__set_responseCode(const int32_t val) {
  this->responseCode = val;
}

void Response::__set_responseJSON(const std::string& val) {
  this->responseJSON = val;
}

uint32_t Response::read(::apache::thrift::protocol::TProtocol* iprot) {

  apache::thrift::protocol::TInputRecursionTracker tracker(*iprot);
  uint32_t xfer = 0;
  std::string fname;
  ::apache::thrift::protocol::TType ftype;
  int16_t fid;

  xfer += iprot->readStructBegin(fname);

  using ::apache::thrift::protocol::TProtocolException;

  bool isset_responseCode = false;
  bool isset_responseJSON = false;

  while (true)
  {
    xfer += iprot->readFieldBegin(fname, ftype, fid);
    if (ftype == ::apache::thrift::protocol::T_STOP) {
      break;
    }
    switch (fid)
    {
      case 1:
        if (ftype == ::apache::thrift::protocol::T_I32) {
          xfer += iprot->readI32(this->responseCode);
          isset_responseCode = true;
        } else {
          xfer += iprot->skip(ftype);
        }
        break;
      case 2:
        if (ftype == ::apache::thrift::protocol::T_STRING) {
          xfer += iprot->readString(this->responseJSON);
          isset_responseJSON = true;
        } else {
          xfer += iprot->skip(ftype);
        }
        break;
      default:
        xfer += iprot->skip(ftype);
        break;
    }
    xfer += iprot->readFieldEnd();
  }

  xfer += iprot->readStructEnd();

  if (!isset_responseCode)
    throw TProtocolException(TProtocolException::INVALID_DATA);
  if (!isset_responseJSON)
    throw TProtocolException(TProtocolException::INVALID_DATA);
  return xfer;
}

uint32_t Response::write(::apache::thrift::protocol::TProtocol* oprot) const {
  uint32_t xfer = 0;
  apache::thrift::protocol::TOutputRecursionTracker tracker(*oprot);
  xfer += oprot->writeStructBegin("Response");

  xfer += oprot->writeFieldBegin("responseCode", ::apache::thrift::protocol::T_I32, 1);
  xfer += oprot->writeI32(this->responseCode);
  xfer += oprot->writeFieldEnd();

  xfer += oprot->writeFieldBegin("responseJSON", ::apache::thrift::protocol::T_STRING, 2);
  xfer += oprot->writeString(this->responseJSON);
  xfer += oprot->writeFieldEnd();

  xfer += oprot->writeFieldStop();
  xfer += oprot->writeStructEnd();
  return xfer;
}

void swap(Response &a, Response &b) {
  using ::std::swap;
  swap(a.responseCode, b.responseCode);
  swap(a.responseJSON, b.responseJSON);
}

Response::Response(const Response& other0) {
  responseCode = other0.responseCode;
  responseJSON = other0.responseJSON;
}
Response& Response::operator=(const Response& other1) {
  responseCode = other1.responseCode;
  responseJSON = other1.responseJSON;
  return *this;
}
void Response::printTo(std::ostream& out) const {
  using ::apache::thrift::to_string;
  out << "Response(";
  out << "responseCode=" << to_string(responseCode);
  out << ", " << "responseJSON=" << to_string(responseJSON);
  out << ")";
}


ServiceException::~ServiceException() throw() {
}


void ServiceException::__set_exceptionCode(const int32_t val) {
  this->exceptionCode = val;
}

void ServiceException::__set_exceptionMeg(const std::string& val) {
  this->exceptionMeg = val;
}

uint32_t ServiceException::read(::apache::thrift::protocol::TProtocol* iprot) {

  apache::thrift::protocol::TInputRecursionTracker tracker(*iprot);
  uint32_t xfer = 0;
  std::string fname;
  ::apache::thrift::protocol::TType ftype;
  int16_t fid;

  xfer += iprot->readStructBegin(fname);

  using ::apache::thrift::protocol::TProtocolException;

  bool isset_exceptionCode = false;
  bool isset_exceptionMeg = false;

  while (true)
  {
    xfer += iprot->readFieldBegin(fname, ftype, fid);
    if (ftype == ::apache::thrift::protocol::T_STOP) {
      break;
    }
    switch (fid)
    {
      case 1:
        if (ftype == ::apache::thrift::protocol::T_I32) {
          xfer += iprot->readI32(this->exceptionCode);
          isset_exceptionCode = true;
        } else {
          xfer += iprot->skip(ftype);
        }
        break;
      case 2:
        if (ftype == ::apache::thrift::protocol::T_STRING) {
          xfer += iprot->readString(this->exceptionMeg);
          isset_exceptionMeg = true;
        } else {
          xfer += iprot->skip(ftype);
        }
        break;
      default:
        xfer += iprot->skip(ftype);
        break;
    }
    xfer += iprot->readFieldEnd();
  }

  xfer += iprot->readStructEnd();

  if (!isset_exceptionCode)
    throw TProtocolException(TProtocolException::INVALID_DATA);
  if (!isset_exceptionMeg)
    throw TProtocolException(TProtocolException::INVALID_DATA);
  return xfer;
}

uint32_t ServiceException::write(::apache::thrift::protocol::TProtocol* oprot) const {
  uint32_t xfer = 0;
  apache::thrift::protocol::TOutputRecursionTracker tracker(*oprot);
  xfer += oprot->writeStructBegin("ServiceException");

  xfer += oprot->writeFieldBegin("exceptionCode", ::apache::thrift::protocol::T_I32, 1);
  xfer += oprot->writeI32(this->exceptionCode);
  xfer += oprot->writeFieldEnd();

  xfer += oprot->writeFieldBegin("exceptionMeg", ::apache::thrift::protocol::T_STRING, 2);
  xfer += oprot->writeString(this->exceptionMeg);
  xfer += oprot->writeFieldEnd();

  xfer += oprot->writeFieldStop();
  xfer += oprot->writeStructEnd();
  return xfer;
}

void swap(ServiceException &a, ServiceException &b) {
  using ::std::swap;
  swap(a.exceptionCode, b.exceptionCode);
  swap(a.exceptionMeg, b.exceptionMeg);
}

ServiceException::ServiceException(const ServiceException& other2) : TException() {
  exceptionCode = other2.exceptionCode;
  exceptionMeg = other2.exceptionMeg;
}
ServiceException& ServiceException::operator=(const ServiceException& other3) {
  exceptionCode = other3.exceptionCode;
  exceptionMeg = other3.exceptionMeg;
  return *this;
}
void ServiceException::printTo(std::ostream& out) const {
  using ::apache::thrift::to_string;
  out << "ServiceException(";
  out << "exceptionCode=" << to_string(exceptionCode);
  out << ", " << "exceptionMeg=" << to_string(exceptionMeg);
  out << ")";
}

const char* ServiceException::what() const throw() {
  try {
    std::stringstream ss;
    ss << "TException - service has thrown: " << *this;
    this->thriftTExceptionMessageHolder_ = ss.str();
    return this->thriftTExceptionMessageHolder_.c_str();
  } catch (const std::exception&) {
    return "TException - service has thrown: ServiceException";
  }
}


