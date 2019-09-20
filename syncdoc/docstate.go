package syncdoc

import (
  "sync"
  "io/ioutil"
  "strings"
  "log"
  "os"
  "path/filepath"
  "math"
)

type docState struct {
  Pathname string
  Lines []string
  LinesMutex sync.Mutex
  History []Change
  HistoryMutex sync.Mutex // TODO: docState should include History
}

func newDocState (pathname string) *docState {
  ds := new(docState)
  ds.Pathname = pathname
  plaintext, err := ioutil.ReadFile("data/" + pathname + ".file") //TODO: SANITATION, CONFIGURABLE PATH
  //log.Printf("read: plaintext for %s: %s", pathname, string(plaintext))
  if err == nil {
    ds.Lines = strings.Split(string(plaintext), "\n")
  } else {
    ds.Lines = []string{""}
  }
  return ds
}

func (ds *docState) GetInitializingChange () Change {
  var chg Change
  chg.Added = ds.Lines
  chg.From.Line = 0
  chg.From.Ch = 0
  chg.To.Line = math.MaxInt32
  chg.To.Ch = math.MaxInt32
  return chg
}

func (ds *docState) Apply (chg Change) {
  ds.HistoryMutex.Lock()
  ds.LinesMutex.Lock()

  if chg.To.Line > len(ds.Lines) { chg.To.Line = len(ds.Lines) }
  if chg.From.Line < 0 { chg.From.Line = 0 }
  if chg.To.Line < chg.From.Line { chg.From.Line = chg.To.Line }

  if chg.From.Ch > len(ds.Lines[chg.From.Line]) { chg.From.Ch = len(ds.Lines[chg.From.Line]) }
  if chg.From.Ch < 0 { chg.From.Ch = 0 }
  if chg.To.Ch > len(ds.Lines[chg.To.Line]) { chg.To.Ch = len(ds.Lines[chg.To.Line]) }
  if chg.To.Ch < 0 { chg.To.Ch = 0 }
  if chg.From.Line == chg.To.Line {
    if chg.To.Ch < chg.From.Ch { chg.To.Ch = chg.From.Ch }
  }

  ds.History = append(ds.History, chg)

  //log.Printf("%d:%d, %d:%d", chg.From.Line, chg.From.Ch, chg.To.Line, chg.To.Ch)
  insertprefix := ds.Lines[chg.From.Line][:chg.From.Ch]
  insertpostfix := ds.Lines[chg.To.Line][chg.To.Ch:]

  newlines := []string{}
  newlines = append(newlines, ds.Lines[:chg.From.Line]...)
  newlines = append(newlines, chg.Added...)
  newlines = append(newlines, ds.Lines[chg.To.Line+1:]...)
  newlines[chg.From.Line] = insertprefix + newlines[chg.From.Line]
  newlines[chg.From.Line + len(chg.Added) - 1] = newlines[chg.From.Line + len(chg.Added) - 1] + insertpostfix
  ds.Lines = newlines

  targetFile := "data/" + ds.Pathname + ".file"
  err := os.MkdirAll(filepath.Dir(targetFile), 0700)
  if err != nil { log.Printf("%v", err) }
  err = ioutil.WriteFile(targetFile, []byte(strings.Join(ds.Lines, "\n")), 0644)
  if err != nil { log.Printf("%v", err) }
  //log.Printf("write: plaintext for %s: %s", ds.Pathname, strings.Join(ds.Lines, "\n"))

  ds.HistoryMutex.Unlock()
  ds.LinesMutex.Unlock()
}
