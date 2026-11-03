package syncdoc

import (
  "github.com/json-iterator/go"
  "sync"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
  "log"
  "io"
  "net"
  "time"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
  MaxMessageBytes = 64 << 10
  writeTimeout = 5 * time.Second
)

type Syncdoc struct {
  Name string
  Connections []net.Conn
  ConnectionsMutex sync.Mutex
  CurrentState *docState
  Users int // guarded by the server's documents mutex
}

func (doc *Syncdoc) AddAndListen(c net.Conn) {
  stringified, err := json.Marshal(doc.CurrentState.GetInitializingChange())
  if err != nil {
    log.Println("encode initializing:", err)
    return
  }
  c.SetWriteDeadline(time.Now().Add(writeTimeout))
  if err := wsutil.WriteServerMessage(c, ws.OpText, stringified); err != nil {
    log.Println("send initializing:", err)
    return
  }

  doc.ConnectionsMutex.Lock()
  doc.Connections = append(doc.Connections, c)
  doc.ConnectionsMutex.Unlock()

  // Read frame by frame so oversized messages are rejected before being buffered.
  rd := wsutil.NewReader(c, ws.StateServerSide)
  controlHandler := wsutil.ControlFrameHandler(c, ws.StateServerSide)
  rd.OnIntermediate = controlHandler
  for {
    hdr, err := rd.NextFrame()
    if err != nil {
      break
    }
    if hdr.OpCode.IsControl() {
      if err := controlHandler(hdr, rd); err != nil {
        break
      }
      continue
    }
    if hdr.OpCode != ws.OpText || !hdr.Fin || hdr.Length > MaxMessageBytes {
      break
    }
    message := make([]byte, hdr.Length)
    if _, err := io.ReadFull(rd, message); err != nil {
      break
    }
    doc.Apply(message, c)
  }
}

func (doc *Syncdoc) RemoveConnection(c net.Conn) {
  doc.ConnectionsMutex.Lock()
  for i := range doc.Connections {
    if doc.Connections[i] == c {
      doc.Connections = append(doc.Connections[:i], doc.Connections[i+1:]...)
      break
    }
  }
  doc.ConnectionsMutex.Unlock()
}

func (doc *Syncdoc) Apply(message []byte, currentconn net.Conn){
  var changeObj Change
  if err := json.Unmarshal(message, &changeObj); err != nil {
    log.Println("decode:", err)
    return
  }
  if err := doc.CurrentState.Apply(changeObj); err != nil {
    log.Printf("[%s] rejected change: %v", doc.Name, err)
    return
  }

  // A deadline per write so one stalled client can't block the others; a failed write
  // closes that connection, which ends its read loop and removes it.
  doc.ConnectionsMutex.Lock()
  for _, otherconn := range doc.Connections {
    if otherconn != currentconn {
      otherconn.SetWriteDeadline(time.Now().Add(writeTimeout))
      if err := wsutil.WriteServerMessage(otherconn, ws.OpText, message); err != nil {
        log.Println(err)
        otherconn.Close()
      }
    }
  }
  doc.ConnectionsMutex.Unlock()
}

func NewDocument(name string) *Syncdoc {
  d := new(Syncdoc)
  d.Name = name
  d.CurrentState = newDocState(name)
  return d
}
