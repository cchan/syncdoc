package syncdoc


// TODO: use protobuf instead, and map out the architecture a little better before jumping in


type cursorpos struct {
  Line    int
  Ch      int
}

type Change struct {
  From    cursorpos
  To      cursorpos
  Added   []string
}

type IncomingEditEvent struct {
  LastSeqNum int
  Chg        Change
}

type OutgoingEditEvent struct {
  SeqNum int
  Chg    Change
}

type OutgoingWatermarkEvent struct {
  SeqNum int
  Lines  []string
}
