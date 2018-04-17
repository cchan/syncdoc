package syncdoc

import (
  "encoding/json"
  "sync"
  "github.com/gorilla/websocket"
  "log"
)

type Syncdoc struct {
  Name string
  Connections []*websocket.Conn
  ConnectionsMutex sync.Mutex
  CurrentState *docState
}

func (doc *Syncdoc) AddConnection(c *websocket.Conn) {
  doc.ConnectionsMutex.Lock()
  doc.Connections = append(doc.Connections, c)
  doc.ConnectionsMutex.Unlock()

  sendEdit(c, doc.CurrentState.GetInitializingChange())
}

func (doc *Syncdoc) RemoveConnection(c *websocket.Conn) {
  doc.ConnectionsMutex.Lock()
  for i := range doc.Connections {
    if doc.Connections[i] == c {
      doc.Connections = append(doc.Connections[:i], doc.Connections[i+1:]...)
      break
    }
  }
  doc.ConnectionsMutex.Unlock()
}

func (doc *Syncdoc) Listen(c *websocket.Conn) {
  for {
    _, message, err := c.ReadMessage()
    if err != nil {
      //log.Println("read:", err)
      break
    }

    //log.Printf("[%s] recv: %s\n", doc.Name, string(message))

    var changeObj Change
    if err := json.Unmarshal(message, &changeObj); err != nil {
      log.Println("decode:", err)
      continue
    }

    doc.Apply(changeObj, c)
  }
}

func (doc *Syncdoc) Apply(changeObj Change, currentconn *websocket.Conn){
  doc.CurrentState.Apply(changeObj)

  for _, otherconn := range doc.Connections {
    if otherconn != currentconn {
      sendEdit(otherconn, changeObj)
    }
  }
}

func NewDocument(name string) *Syncdoc {
  d := new(Syncdoc)
  d.Name = name
  d.CurrentState = newDocState(name)
  return d
}

func sendEdit(c *websocket.Conn, changeObj Change) error {
  stringified, err := json.Marshal(changeObj)
  if err != nil {
    return err
  }
  if err := c.WriteMessage(websocket.TextMessage, stringified); err != nil {
    return err
  }
  return nil
}
