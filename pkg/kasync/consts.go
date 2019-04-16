package kasync

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
	MsgTypeUnk      MsgHdr  = "UNK"
)
