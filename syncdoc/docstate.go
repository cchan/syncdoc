package syncdoc

import (
  "errors"
  "sync"
  "io/ioutil"
  "strings"
  "log"
  "os"
  "path/filepath"
  "math"
)

const dataDir = "/home/www/data/"

// Limits so one client can't exhaust memory or disk on the host.
const (
  MaxDocBytes = 256 << 10
  MaxDocFiles = 2000
)

var errDocTooBig = errors.New("document would exceed MaxDocBytes")

type docState struct {
  Pathname string
  Lines []string
  LinesMutex sync.Mutex
}

func newDocState (pathname string) *docState {
  ds := new(docState)
  ds.Pathname = pathname
  plaintext, err := ioutil.ReadFile(dataDir + pathname + ".file") // pathname is validated by the caller
  if err == nil {
    ds.Lines = strings.Split(string(plaintext), "\n")
  } else {
    ds.Lines = []string{""}
  }
  return ds
}

func (ds *docState) GetInitializingChange () Change {
  ds.LinesMutex.Lock()
  defer ds.LinesMutex.Unlock()
  var chg Change
  chg.Added = append([]string(nil), ds.Lines...)
  chg.From.Line = 0
  chg.From.Ch = 0
  chg.To.Line = math.MaxInt32
  chg.To.Ch = math.MaxInt32
  return chg
}

func clamp(v, lo, hi int) int {
  if v < lo { return lo }
  if v > hi { return hi }
  return v
}

func (ds *docState) Apply (chg Change) error {
  ds.LinesMutex.Lock()
  defer ds.LinesMutex.Unlock()

  last := len(ds.Lines) - 1
  chg.From.Line = clamp(chg.From.Line, 0, last)
  chg.To.Line = clamp(chg.To.Line, 0, last)
  if chg.To.Line < chg.From.Line { chg.From.Line = chg.To.Line }
  chg.From.Ch = clamp(chg.From.Ch, 0, len(ds.Lines[chg.From.Line]))
  chg.To.Ch = clamp(chg.To.Ch, 0, len(ds.Lines[chg.To.Line]))
  if chg.From.Line == chg.To.Line && chg.To.Ch < chg.From.Ch { chg.To.Ch = chg.From.Ch }
  if len(chg.Added) == 0 { chg.Added = []string{""} }

  insertprefix := ds.Lines[chg.From.Line][:chg.From.Ch]
  insertpostfix := ds.Lines[chg.To.Line][chg.To.Ch:]

  newlines := []string{}
  newlines = append(newlines, ds.Lines[:chg.From.Line]...)
  newlines = append(newlines, chg.Added...)
  newlines = append(newlines, ds.Lines[chg.To.Line+1:]...)
  newlines[chg.From.Line] = insertprefix + newlines[chg.From.Line]
  newlines[chg.From.Line + len(chg.Added) - 1] = newlines[chg.From.Line + len(chg.Added) - 1] + insertpostfix

  size := len(newlines)
  for _, l := range newlines { size += len(l) }
  if size > MaxDocBytes { return errDocTooBig }
  ds.Lines = newlines

  targetFile := dataDir + ds.Pathname + ".file"
  if _, err := os.Stat(targetFile); os.IsNotExist(err) {
    if entries, err := ioutil.ReadDir(dataDir); err == nil && len(entries) >= MaxDocFiles {
      log.Printf("not saving new doc %s: %d files already", ds.Pathname, len(entries))
      return nil
    }
  }
  err := os.MkdirAll(filepath.Dir(targetFile), 0700)
  if err != nil { log.Printf("%v", err) }
  err = ioutil.WriteFile(targetFile, []byte(strings.Join(ds.Lines, "\n")), 0644)
  if err != nil { log.Printf("%v", err) }
  return nil
}
