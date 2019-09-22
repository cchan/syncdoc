package syncdoc

import (
  "github.com/json-iterator/go"
  "sync"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
  "log"
  "io"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Syncdoc struct {
  Name string
  Connections []io.ReadWriter
  ConnectionsMutex sync.Mutex
  CurrentState *docState
}

func (doc *Syncdoc) AddAndListen(c io.ReadWriter) {
  stringified, err := json.Marshal(doc.CurrentState.GetInitializingChange())
  if err != nil {
    log.Println("encode initializing:", err)
    return
  }
  if err := wsutil.WriteServerMessage(c, ws.OpText, stringified); err != nil {
    log.Println("send initializing:", err)
    return
  }

  doc.ConnectionsMutex.Lock()
  doc.Connections = append(doc.Connections, c)
  doc.ConnectionsMutex.Unlock()

  for {
    message, op, err := wsutil.ReadClientData(c)
    if err != nil || op != ws.OpText {
      //log.Println("read:", err)
      break
    }

    //log.Printf("[%s] recv: %s\n", doc.Name, string(message))

    doc.Apply(message, c)
  }
}

func (doc *Syncdoc) RemoveConnection(c io.ReadWriter) {
  doc.ConnectionsMutex.Lock()
  for i := range doc.Connections {
    if doc.Connections[i] == c {
      doc.Connections = append(doc.Connections[:i], doc.Connections[i+1:]...)
      break
    }
  }
  doc.ConnectionsMutex.Unlock()
}

func (doc *Syncdoc) Apply(message []byte, currentconn io.ReadWriter){
  var changeObj Change
  if err := json.Unmarshal(message, &changeObj); err != nil {
    log.Println("decode:", err)
  } else {
    doc.CurrentState.Apply(changeObj)

    // What happens when RemoveConnection while doing this?
    // Can I not have a lock here?
    doc.ConnectionsMutex.Lock()
    for _, otherconn := range doc.Connections {
      if otherconn != currentconn {
        if err := wsutil.WriteServerMessage(otherconn, ws.OpText, message); err != nil {
          log.Println(err)
        }
      }
    }
    doc.ConnectionsMutex.Unlock()
  }
}

func NewDocument(name string) *Syncdoc {
  d := new(Syncdoc)
  d.Name = name
  d.CurrentState = newDocState(name)
  return d
}
