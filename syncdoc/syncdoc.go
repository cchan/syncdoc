package syncdoc

import (
  "encoding/json"
  "sync"
  "github.com/gorilla/websocket"
  "log"
)

type cursorpos struct {
  Line    int32
  Ch      int32
}

type Change struct {
  From    cursorpos
  To      cursorpos
  Added   []string
}

type Syncdoc struct {
  Name string
  History []Change
  HistoryMutex *sync.Mutex
  Connections []*websocket.Conn
  ConnectionsMutex *sync.Mutex
}

func (doc *Syncdoc) AddConnection(c *websocket.Conn) {
  doc.ConnectionsMutex.Lock()
  doc.Connections = append(doc.Connections, c)
  doc.ConnectionsMutex.Unlock()

  doc.HistoryMutex.Lock()
  for i := range doc.History {
    sendEdit(c, doc.History[i])
  }
  doc.HistoryMutex.Unlock()
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
      log.Println("read:", err)
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
  doc.HistoryMutex.Lock()
  doc.History = append(doc.History, changeObj)
  doc.HistoryMutex.Unlock()

  for _, otherconn := range doc.Connections {
    if otherconn != currentconn {
      sendEdit(otherconn, changeObj)
    }
  }
}

func NewDocument(name string) *Syncdoc {
  d := new(Syncdoc)
  d.Name = name
  d.History = make([]Change, 0)
  d.HistoryMutex = &sync.Mutex{}
  d.Connections = make([]*websocket.Conn, 0)
  d.ConnectionsMutex =  &sync.Mutex{}
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
