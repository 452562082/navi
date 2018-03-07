package navicli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"kuaishangtong/common/utils/log"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type switcher func(s Servable, methodName string, resp http.ResponseWriter, req *http.Request) (interface{}, error)

var switcherFunc switcher

func router(s Servable) *mux.Router {
	r := mux.NewRouter()

	for _, v := range s.ServerField().Config.mappings[urlServiceMaps] {
		httpMethods := strings.Split(v[0], ",")
		path := strFirstToUpper(v[1])
		methodName := strFirstToUpper(v[2])
		r.HandleFunc(path, handler(s, methodName)).Methods(httpMethods...)
	}
	return r
}

type key int

var componentsKey key = 0

// copyComponentsPtr copies the ptr to Server.Components to context, to ensure that,
// we are using the same Components through out one request lifecycle.
// Server.Components may change on reloading config.
func copyComponentsPtr(s Servable, req *http.Request) {
	ctx := context.WithValue(req.Context(), componentsKey, s.ServerField().Components)
	*req = *req.WithContext(ctx)
}

func components(req *http.Request) *Components {
	return req.Context().Value(componentsKey).(*Components)
}

func handler(s Servable, methodName string) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		copyComponentsPtr(s, req)
		parseRequestForm(req)
		interceptors := getInterceptors(s, req)
		req, err := doBefore(&interceptors, resp, req)
		if err == nil {
			doRequest(s, methodName, resp, req)
		} else {
			components(req).errorHandlerFunc()(resp, req, err)
		}
		doAfter(interceptors, resp, req)
	}
}

func getInterceptors(s Servable, req *http.Request) []Interceptor {
	interceptors := components(req).Interceptors(req)
	if len(interceptors) == 0 {
		interceptors = components(req).CommonInterceptors()
	}
	return interceptors
}

func doBefore(interceptors *[]Interceptor, resp http.ResponseWriter, req *http.Request) (request *http.Request, err error) {
	for index, i := range *interceptors {
		err = i.Before(resp, req)
		if err != nil {
			log.Error("error in Before(): ", err.Error())
			*interceptors = (*interceptors)[0:index]
			return req, err
		}
	}
	return req, nil
}

func doRequest(s Servable, methodName string, resp http.ResponseWriter, req *http.Request) {
	if hijack := components(req).Hijacker(req); hijack != nil {
		hijack(resp, req)
		return
	}
	err := doPreprocessor(s, resp, req)
	if err != nil {
		components(req).errorHandlerFunc()(resp, req, err)
		return
	}
	serviceResp, err := switcherFunc(s, methodName, resp, req)
	if err != nil {
		log.Infof("request RemoteAddr: %s, url: %s, error: %v", req.Header.Get("RemoteAddr"), req.URL.Path, err)
		components(req).errorHandlerFunc()(resp, req, err)
		return
	}
	log.Infof("request RemoteAddr: %s, url: %s, success", req.Header.Get("RemoteAddr"), req.URL.Path)
	doPostprocessor(s, resp, req, serviceResp, err)
}

type headerKey struct{}
type trailerKey struct{}
type peerKey struct{}

// CallOptions returns grpc CallOptions,
// you can overwrite this func to build CallOptions for your needs.
var CallOptions = func(methodName string, req *http.Request) ([]grpc.CallOption, *metadata.MD, *metadata.MD, *peer.Peer) {
	header := new(metadata.MD)
	trailer := new(metadata.MD)
	peer := &peer.Peer{}
	callOptions := []grpc.CallOption{
		grpc.Header(header),
		grpc.Trailer(trailer),
		grpc.Peer(peer),
	}
	return callOptions, header, trailer, peer
}

// WithCallOptions read header, trailer and peer into context
func WithCallOptions(req *http.Request, header *metadata.MD, trailer *metadata.MD, peer *peer.Peer) {
	ctx := context.WithValue(req.Context(), headerKey{}, header)
	ctx = context.WithValue(ctx, trailerKey{}, trailer)
	ctx = context.WithValue(ctx, peerKey{}, peer)
	*req = *req.WithContext(ctx)
}

// GrpcMetadataHeader returns the header in metadata
func GrpcMetadataHeader(ctx context.Context) *metadata.MD {
	return ctx.Value(headerKey{}).(*metadata.MD)
}

// GrpcMetadataTrailer returns the trailer in metadata
func GrpcMetadataTrailer(ctx context.Context) *metadata.MD {
	return ctx.Value(trailerKey{}).(*metadata.MD)
}

// GrpcMetadataPeer returns the peer in metadata
func GrpcMetadataPeer(ctx context.Context) *peer.Peer {
	return ctx.Value(peerKey{}).(*peer.Peer)
}

func doPreprocessor(s Servable, resp http.ResponseWriter, req *http.Request) error {
	if pre := components(req).Preprocessor(req); pre != nil {
		if err := pre(resp, req); err != nil {
			log.Error(err.Error())
			return errors.New(fmt.Sprintf("turbo: encounter error in preprocessor for %s, error: %s", req.URL, err))
		}
	}
	return nil
}

func doPostprocessor(s Servable, resp http.ResponseWriter, req *http.Request, serviceResponse interface{}, err error) {
	// run Postprocessor, if any
	post := components(req).Postprocessor(req)
	if post != nil {
		post(resp, req, serviceResponse, err)
		return
	}

	// return as json
	m := Marshaler{
		FilterProtoJson: s.ServerField().Config.FilterProtoJson(),
		EmitZeroValues:  s.ServerField().Config.FilterProtoJsonEmitZeroValues(),
		Int64AsNumber:   s.ServerField().Config.FilterProtoJsonInt64AsNumber(),
	}
	jsonBytes, err := m.JSON(serviceResponse)
	if err == nil {
		resp.Write(jsonBytes)
	} else {
		log.Error(err.Error())
		resp.Write([]byte(fmt.Sprintf("turbo: encounter error while converting response to json "+
			"in doPostprocessor() for %s, error: %s", req.URL, err)))
	}
}

func doAfter(interceptors []Interceptor, resp http.ResponseWriter, req *http.Request) (err error) {
	l := len(interceptors)
	for i := l - 1; i >= 0; i-- {
		err = interceptors[i].After(resp, req)
		if err != nil {
			log.Errorf("turbo: error in After(): ", err.Error())
		}
	}
	return nil
}

//BuildStruct finds values from request, and set them to struct fields recursively
func BuildStruct(s Servable, theType reflect.Type, theValue reflect.Value, req *http.Request) {
	if theValue.Kind() == reflect.Invalid {
		log.Info("value is invalid, please check grpc-fieldmapping")
	}
	convertor := components(req).Convertor(theValue.Type().Name())
	if convertor != nil {
		theValue.Set(convertor(req).Elem())
		return
	}

	fieldNum := theType.NumField()
	for i := 0; i < fieldNum; i++ {
		fieldName := theType.Field(i).Name
		fieldValue := theValue.FieldByName(fieldName)
		if fieldValue.Kind() == reflect.Ptr && fieldValue.Type().Elem().Kind() == reflect.Struct {
			convertor := components(req).Convertor(fieldValue.Type().Elem().Name())
			if convertor != nil {
				fieldValue.Set(convertor(req))
				continue
			}
			BuildStruct(s, fieldValue.Type().Elem(), fieldValue.Elem(), req)
			continue
		}
		v, ok := findValue(fieldName, req)
		if !ok {
			continue
		}
		err := setValue(theType.Field(i).Type, fieldValue, v)
		logErrorIf(err)
	}
}

// setValue sets v to fieldValue according to fieldValue's Kind
func setValue(fieldType reflect.Type, fieldValue reflect.Value, v string) error {
	var err error
	switch k := fieldValue.Kind(); k {
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		var i int64
		i, err = strconv.ParseInt(v, 10, 64)
		fieldValue.SetInt(i)
	case reflect.String:
		fieldValue.SetString(v)
	case reflect.Bool:
		var b bool
		b, err = strconv.ParseBool(v)
		fieldValue.SetBool(b)
	case reflect.Float32, reflect.Float64:
		var f float64
		f, err = strconv.ParseFloat(v, 64)
		fieldValue.SetFloat(f)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var u uint64
		u, err = strconv.ParseUint(v, 10, 64)
		fieldValue.SetUint(u)
	case reflect.Slice:
		err = setSliceValue(fieldType, fieldValue, v)
	default:
		return errors.New("turbo: not supported kind[" + k.String() + "]")
	}
	return err
}

func setSliceValue(fieldType reflect.Type, fieldValue reflect.Value, v string) error {
	if len(v) == 0 {
		fieldValue.Set(reflect.MakeSlice(fieldType, 0, 0))
		return nil
	}
	vSlice := strings.Split(v, ",")
	length := len(vSlice)
	s := reflect.MakeSlice(fieldType, length, length)
	switch fieldType.Elem().Kind() {
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		for k, v := range vSlice {
			value, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return err
			}
			s.Index(k).SetInt(value)
		}
	case reflect.String:
		for k, v := range vSlice {
			s.Index(k).SetString(v)
		}
	case reflect.Bool:
		for k, v := range vSlice {
			value, err := strconv.ParseBool(v)
			if err != nil {
				return err
			}
			s.Index(k).SetBool(value)
		}
	case reflect.Float32, reflect.Float64:
		for k, v := range vSlice {
			value, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return err
			}
			s.Index(k).SetFloat(value)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		for k, v := range vSlice {
			value, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return err
			}
			s.Index(k).SetUint(value)
		}
	}
	fieldValue.Set(s)
	return nil
}

// BuildArgs returns a list of reflect.Value for thrift request
func BuildArgs(s Servable, argsType reflect.Type, argsValue reflect.Value, req *http.Request, buildStructArg func(s Servable, typeName string, req *http.Request) (v reflect.Value, err error)) ([]reflect.Value, error) {
	fieldNum := argsType.NumField()
	params := make([]reflect.Value, fieldNum)
	for i := 0; i < fieldNum; i++ {
		field := argsType.Field(i)
		fieldName := field.Name
		valueType := argsValue.FieldByName(fieldName).Type()
		if field.Type.Kind() == reflect.Ptr && valueType.Elem().Kind() == reflect.Struct {
			convertor := components(req).Convertor(valueType.Elem().Name())
			if convertor != nil {
				params[i] = convertor(req)
				continue
			}
			structName := valueType.Elem().Name()
			v, err := buildStructArg(s, structName, req)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("turbo: failed to BuildArgs, error:%s", err))
			}
			params[i] = v
			continue
		}
		v, _ := findValue(fieldName, req)
		value, err := reflectValue(field.Type, argsValue.FieldByName(fieldName), v)
		logErrorIf(err)
		params[i] = value
	}
	return params, nil
}

// reflectValue returns a reflect.Value with v according to fieldValue's Kind
func reflectValue(fieldType reflect.Type, fieldValue reflect.Value, v string) (reflect.Value, error) {
	switch k := fieldValue.Kind(); k {
	case reflect.Int16:
		i, err := strconv.ParseInt(v, 10, 16)
		if err != nil {
			return reflect.ValueOf(int16(0)), err
		}
		return reflect.ValueOf(int16(i)), nil
	case reflect.Int32:
		i, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return reflect.ValueOf(int32(0)), err
		}
		return reflect.ValueOf(int32(i)), nil
	case reflect.Int64:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return reflect.ValueOf(int64(0)), err
		}
		return reflect.ValueOf(int64(i)), nil
	case reflect.String:
		return reflect.ValueOf(v), nil
	case reflect.Bool:
		b, err := strconv.ParseBool(v)
		if err != nil {
			return reflect.ValueOf(false), err
		}
		return reflect.ValueOf(bool(b)), nil
	case reflect.Float64:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return reflect.ValueOf(float64(0)), errors.New("error float")
		}
		return reflect.ValueOf(float64(f)), nil
	case reflect.Slice:
		return reflectSliceValue(fieldType, fieldValue, v)
	default:
		return reflect.ValueOf(0), errors.New("turbo: not supported kind[" + k.String() + "]")
	}
}

func reflectSliceValue(fieldType reflect.Type, fieldValue reflect.Value, v string) (reflect.Value, error) {
	if len(v) == 0 {
		return reflect.MakeSlice(fieldType, 0, 0), nil
	}
	vSlice := strings.Split(v, ",")
	length := len(vSlice)
	s := reflect.MakeSlice(fieldType, length, length)
	switch fieldType.Elem().Kind() {
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		for k, v := range vSlice {
			value, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return s.Slice(0, 0), err
			}
			s.Index(k).SetInt(value)
		}
	case reflect.String:
		for k, v := range vSlice {
			s.Index(k).SetString(v)
		}
	case reflect.Bool:
		for k, v := range vSlice {
			value, err := strconv.ParseBool(v)
			if err != nil {
				return s.Slice(0, 0), err
			}
			s.Index(k).SetBool(value)
		}
	case reflect.Float32, reflect.Float64:
		for k, v := range vSlice {
			value, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return s.Slice(0, 0), err
			}
			s.Index(k).SetFloat(value)
		}
	}
	return s, nil
}

func findValue(fieldName string, req *http.Request) (string, bool) {
	lowerCasesName := strings.ToLower(fieldName)
	snakeCaseName := ToSnakeCase(fieldName)

	v, ok := req.Form[lowerCasesName]
	if ok && len(v) > 0 {
		return v[0], true
	}
	v, ok = req.Form[snakeCaseName]
	if ok && len(v) > 0 {
		return v[0], true
	}
	ctxValue := req.Context().Value(fieldName)
	if ctxValue != nil {
		return ctxValue.(string), true
	}
	ctxValue = req.Context().Value(lowerCasesName)
	if ctxValue != nil {
		return ctxValue.(string), true
	}
	ctxValue = req.Context().Value(snakeCaseName)
	if ctxValue != nil {
		return ctxValue.(string), true
	}
	return "", false
}

func BuildRequest(s Servable, v proto.Message, req *http.Request) error {
	var err error
	if contentTypes, ok := req.Header["Content-Type"]; ok && contentTypes[0] == "application/json" {
		buf := new(bytes.Buffer)
		buf.ReadFrom(req.Body)
		bodyStr := buf.String()
		unmarshaler := &jsonpb.Unmarshaler{AllowUnknownFields: true}
		err = unmarshaler.Unmarshal(strings.NewReader(bodyStr), v)
		if err != nil {
			return errors.New(fmt.Sprintf("turbo: failed to BuildRequest for json api, "+
				"request body: %s, error: %s", bodyStr, err))
		}
		setPathParams(reflect.TypeOf(v).Elem(), reflect.ValueOf(v).Elem(), req)
	} else {
		BuildStruct(s, reflect.TypeOf(v).Elem(), reflect.ValueOf(v).Elem(), req)
	}
	return err
}

func BuildThriftRequest(s Servable, args interface{}, req *http.Request, buildStructArg func(s Servable, typeName string, req *http.Request) (v reflect.Value, err error)) ([]reflect.Value, error) {
	var err error
	var params []reflect.Value
	if contentTypes, ok := req.Header["Content-Type"]; ok && strings.Contains(contentTypes[0], "application/json") {
		buf := new(bytes.Buffer)
		buf.ReadFrom(req.Body)
		v := reflect.New(reflect.ValueOf(args).Elem().Type()).Interface()
		err := json.Unmarshal(buf.Bytes(), v)
		if err != nil {
			return params, errors.New(fmt.Sprintf("turbo: failed to BuildThriftRequest for json api, "+
				"request body: %s, error: %s", buf.String(), err))
		}

		setPathParams(reflect.TypeOf(v).Elem(), reflect.ValueOf(v).Elem(), req)
		numFields := reflect.TypeOf(v).Elem().NumField()
		params = make([]reflect.Value, numFields)
		for i := 0; i < numFields; i++ {
			params[i] = reflect.ValueOf(v).Elem().Field(i)
		}

	} else {
		params, err = BuildArgs(s, reflect.TypeOf(args).Elem(), reflect.ValueOf(args).Elem(), req, buildStructArg)
	}
	return params, err
}

func setPathParams(theType reflect.Type, theValue reflect.Value, req *http.Request) {
	fieldNum := theType.NumField()
	pathParams := mux.Vars(req)
	for i := 0; i < fieldNum; i++ {
		fieldName := theType.Field(i).Name
		fieldValue := theValue.FieldByName(fieldName)
		if fieldValue.Kind() == reflect.Ptr && fieldValue.Type().Elem().Kind() == reflect.Struct {
			setPathParams(fieldValue.Type().Elem(), fieldValue.Elem(), req)
			continue
		}
		v, ok := findPathParamValue(fieldName, pathParams)
		if !ok {
			continue
		}
		err := setValue(theType.Field(i).Type, fieldValue, v)
		logErrorIf(err)
	}
}

func findPathParamValue(fieldName string, pathParams map[string]string) (string, bool) {
	lowerCasesName := strings.ToLower(fieldName)
	v, ok := pathParams[lowerCasesName]
	if ok && len(v) > 0 {
		return v, true
	}
	snakeCaseName := ToSnakeCase(fieldName)
	v, ok = pathParams[snakeCaseName]
	if ok && len(v) > 0 {
		return v, true
	}
	return "", false
}
