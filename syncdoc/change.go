package syncdoc


type cursorpos struct {
  Line    int32
  Ch      int32
}

type Change struct {
  From    cursorpos
  To      cursorpos
  Added   []string
}
