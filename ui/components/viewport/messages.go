package viewport

type MsgStatus string
type MsgError error
type MsgOpenBuffer string
type MsgWriteBuffer string
type MsgCloseBuffers []string
type MsgCloseBuffersForced []string
