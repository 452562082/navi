package agent

import (
	"context"
	"reflect"
	"sync"
	"unicode"
	"unicode/utf8"
)

// Precompute the reflect type for error. Can't use error directly
// because Typeof takes an empty interface value. This is annoying.
var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

// Precompute the reflect type for context.
var typeOfContext = reflect.TypeOf((*context.Context)(nil)).Elem()

type methodType struct {
	sync.Mutex // protects counters
	method     reflect.Method
	ArgType    reflect.Type
	ReplyType  reflect.Type
	numCalls   uint
}

type functionType struct {
	sync.Mutex // protects counters
	fn         reflect.Value
	ArgType    reflect.Type
	ReplyType  reflect.Type
	numCalls   uint
}

type service struct {
	name     string                   // name of service
	rcvr     reflect.Value            // receiver of methods for the service
	typ      reflect.Type             // type of the receiver
	method   map[string]*methodType   // registered methods
	function map[string]*functionType // registered functions
}

func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

// RegisterName is like Register but uses the provided name for the type
// instead of the receiver's concrete type.
func (a *Agent) RegisterName(name string, rcvr interface{}, metadata string) error {
	if a.Plugins == nil {
		a.Plugins = &pluginContainer{}
	}

	return a.Plugins.DoRegister(name, rcvr, metadata)
	//return a.register(rcvr, name, true)
}

func (a *Agent) UnRegisterName(name string) error {
	if a.Plugins == nil {
		a.Plugins = &pluginContainer{}
	}

	return a.Plugins.DoUnRegister(name)
	//return a.register(rcvr, name, true)
}
