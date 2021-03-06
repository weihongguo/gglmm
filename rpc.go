package gglmm

import (
	"log"
	"net/rpc"
	"reflect"
	"strings"
)

// RPCAction --
type RPCAction struct {
	name     string
	request  string
	response string
}

// NewRPCAction --
func NewRPCAction(name string, request string, response string) *RPCAction {
	return &RPCAction{
		name:     name,
		request:  request,
		response: response,
	}
}

func (info RPCAction) String() string {
	return info.name + "(" + info.request + ", " + info.response + ")"
}

// RPCActionsResponse --
type RPCActionsResponse struct {
	Actions []*RPCAction
}

// RPCHandler --
type RPCHandler interface {
	Actions(cmd string, actions *RPCActionsResponse) error
}

// RPCHandlerConfig --
type RPCHandlerConfig struct {
	name       string
	rpcHandler RPCHandler
}

var rpcHandlerConfigs []*RPCHandlerConfig = nil

// RegisterRPC 注册RPCHandler
// rpcHandler 处理者
func RegisterRPC(rpcHandler RPCHandler) *RPCHandlerConfig {
	handlerType := reflect.TypeOf(rpcHandler)
	if handlerType.Kind() == reflect.Ptr {
		handlerType = handlerType.Elem()
	}
	name := handlerType.Name()
	return RegisterRPCName(name, rpcHandler)
}

// RegisterRPCName 注册RPCHandler
// name 名称
// rpcHandler 处理者
func RegisterRPCName(name string, rpcHandler RPCHandler) *RPCHandlerConfig {
	if rpcHandlerConfigs == nil {
		rpcHandlerConfigs = make([]*RPCHandlerConfig, 0)
	}
	config := &RPCHandlerConfig{
		name:       name,
		rpcHandler: rpcHandler,
	}
	rpcHandlerConfigs = append(rpcHandlerConfigs, config)
	return config
}

func registerRPC() {
	if rpcHandlerConfigs == nil || len(rpcHandlerConfigs) == 0 {
		return
	}
	for _, config := range rpcHandlerConfigs {
		rpcActionsResponse := RPCActionsResponse{}
		config.rpcHandler.Actions("all", &rpcActionsResponse)
		rpcInfos := []string{}
		for _, action := range rpcActionsResponse.Actions {
			rpcInfos = append(rpcInfos, action.String())
		}
		rpc.RegisterName(config.name, config.rpcHandler)
		log.Printf("[ rpc] %s [%s]\n", config.name, strings.Join(rpcInfos, "; "))
	}
}
