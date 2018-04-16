package syncdoc

import (
  "sync"
  "io/ioutil"
  "strings"
  "log"
  "os"
  "path/filepath"
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
  chg.Added = append([]string(nil), ds.Lines...)
  return chg
}

func (ds *docState) Apply (chg Change) {
  ds.HistoryMutex.Lock()
  ds.History = append(ds.History, chg)
  ds.HistoryMutex.Unlock()

  ds.LinesMutex.Lock()

  newlines := append([]string(nil), chg.Added...)
  newlines[0] = ds.Lines[chg.From.Line][:chg.From.Ch] + newlines[0]
  newlines[len(newlines)-1] = newlines[len(newlines)-1] + ds.Lines[chg.To.Line][chg.To.Ch:]

  ds.Lines = append(append(ds.Lines[:chg.From.Line], newlines...), ds.Lines[chg.To.Line+1:]...)

  
  targetFile := "data/" + ds.Pathname + ".file"
  err := os.MkdirAll(filepath.Dir(targetFile), 0700)
  if err != nil { log.Printf("%v", err) }
  err = ioutil.WriteFile(targetFile, []byte(strings.Join(ds.Lines, "\n")), 0644)
  if err != nil { log.Printf("%v", err) }
  //log.Printf("write: plaintext for %s: %s", ds.Pathname, strings.Join(ds.Lines, "\n"))

  ds.LinesMutex.Unlock()
}
