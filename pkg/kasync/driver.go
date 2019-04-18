package kasync

import (
	"context"
	"io"
)

type MsgHdr = string
type MsgType = string
type CtxKey string

const (
	MsgHdrMsgType MsgHdr = "X-MsgType"
	MsgHdrMsgName MsgHdr = "X-MsgName"
	MsgHdrCntType MsgHdr = "X-CntType"
	MsgHdrValUnk  MsgHdr = "UNK"
)

const (
	CtxKeyMsgID   CtxKey = "k-msgid"
	CtxKeyMsgName CtxKey = "k-msgname"
)

const (
	MsgTypeEvent    MsgType = "EVENT"
	MsgTypeErrEvent MsgType = "ERR_EVENT"
	MsgTypeUnk      MsgType = "UNK"
)

type Router interface {
	io.Closer
	Listen() error
	NewRouteGrp(topic string, defHandler MsgHandler) RouteGroup
	RouteGroup(topic string) (RouteGroup, error)
	RqTopics() []string
}

type RouteGroup interface {
	SetMsgNameResolver(r ResolveMsgName)
	HandleMsg(msgName string, handler MsgHandler)
	MsgHandler(msgName string) (MsgHandler, error)
	ResolveMsgName(msg interface{}) (string, error)
}

type MsgHandler func(ctx context.Context, data []byte) error

type ResolveMsgName func(msg interface{}) (string, error)
