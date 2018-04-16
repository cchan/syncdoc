package syncdoc


type cursorpos struct {
  Line    int
  Ch      int
}

type Change struct {
  From    cursorpos
  To      cursorpos
  Added   []string
}
