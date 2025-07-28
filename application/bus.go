package application

import (
	"errors"
	"reflect"
)

type Command interface{}
type Query interface{}

type CommandHandler interface {
	Handle(Command) error
}

type QueryHandler interface {
	Handle(Query) (interface{}, error)
}

type CommandBus struct {
	handlers map[reflect.Type]CommandHandler
}

func NewCommandBus() *CommandBus {
	return &CommandBus{handlers: make(map[reflect.Type]CommandHandler)}
}

func (cb *CommandBus) Register(cmd Command, handler CommandHandler) {
	cmdType := reflect.TypeOf(cmd)
	cb.handlers[cmdType] = handler
}

func (cb *CommandBus) Dispatch(cmd Command) error {
	cmdType := reflect.TypeOf(cmd)
	handler, ok := cb.handlers[cmdType]
	if !ok {
		return errors.New("no command handler registered for type: " + cmdType.String())
	}
	return handler.Handle(cmd)
}

type QueryBus struct {
	handlers map[reflect.Type]QueryHandler
}

func NewQueryBus() *QueryBus {
	return &QueryBus{handlers: make(map[reflect.Type]QueryHandler)}
}

func (qb *QueryBus) Register(query Query, handler QueryHandler) {
	qType := reflect.TypeOf(query)
	qb.handlers[qType] = handler
}

func (qb *QueryBus) Dispatch(query Query) (interface{}, error) {
	qType := reflect.TypeOf(query)
	handler, ok := qb.handlers[qType]
	if !ok {
		return nil, errors.New("no query handler registered for type: " + qType.String())
	}
	return handler.Handle(query)
}
