// Plaintext OT.

// Try this later: https://medium.freecodecamp.org/million-websockets-and-go-cc58418460bb

package main

import (
  "io/ioutil"
  "net/http"
  "log"
  "strconv"
  "os"
  "github.com/gorilla/websocket"
  "encoding/json"
  "sync"
  "strings"
  //"errors"
)

type cursorpos struct {
  Line    int32
  Ch      int32
}

type change struct {
  From    cursorpos
  To      cursorpos
  Added   []string
}

type document struct {
  Name string
  History []change
  HistoryMutex *sync.Mutex
  Connections []*websocket.Conn
  ConnectionsMutex *sync.Mutex
}

func NewDocument(name string) *document {
  d := new(document)
  d.Name = name
  d.History = make([]change, 0)
  d.HistoryMutex = &sync.Mutex{}
  d.Connections = make([]*websocket.Conn, 0)
  d.ConnectionsMutex =  &sync.Mutex{}
  return d
}

func sendEdit(c *websocket.Conn, changeObj change) error {
  stringified, err := json.Marshal(changeObj)
  if err != nil {
    return err
  }
  if err := c.WriteMessage(websocket.TextMessage, stringified); err != nil {
    return err
  }
  return nil
}

func recvEdits(doc *document, c *websocket.Conn) {
  for {
    _, message, err := c.ReadMessage()
    if err != nil {
      log.Println("read:", err)
      break
    }

    log.Printf("[%s] recv: %s\n", doc.Name, string(message))

    var changeObj change
    if err := json.Unmarshal(message, &changeObj); err != nil {
      log.Println("decode:", err)
      continue
    }

    doc.HistoryMutex.Lock()
    doc.History = append(doc.History, changeObj)
    doc.HistoryMutex.Unlock()

    for _, otherconn := range doc.Connections {
      if otherconn != c {
        sendEdit(otherconn, changeObj)
      }
    }
  }
}

var Documents = make(map[string]*document)


var upgrader = websocket.Upgrader{}

func edit(w http.ResponseWriter, r *http.Request) {
  c, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    log.Print("upgrade:", err)
    return
  }
  defer c.Close()

  docname := strings.TrimPrefix(string(r.URL.Path), "/ws/")
  log.Println(docname)

  if Documents[docname] == nil {
    Documents[docname] = NewDocument(docname)
  }
  doc := Documents[docname]

  doc.ConnectionsMutex.Lock()
  doc.Connections = append(doc.Connections, c)
  doc.ConnectionsMutex.Unlock()

  doc.HistoryMutex.Lock()
  for i := range doc.History {
    sendEdit(c, doc.History[i])
  }
  doc.HistoryMutex.Unlock()

  log.Println("Connected")

  recvEdits(doc, c)

  log.Println("Disconnected")

  doc.ConnectionsMutex.Lock()
  for i := range doc.Connections {
    if doc.Connections[i] == c {
      doc.Connections = append(doc.Connections[:i], doc.Connections[i+1:]...)
      break
    }
  }
  doc.ConnectionsMutex.Unlock()
}

func main() {
  port := os.Getenv("PORT")
  if _, err := strconv.ParseInt(port, 10, 16); err != nil {
    port = "8080"
  }
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    html, err := ioutil.ReadFile("index.html")
    if err != nil { log.Fatal("Could not open index.html for reading") }
    log.Printf("Requested %v %v %v", r.Method, r.URL, r.Proto)
    w.Header().Set("Content-Type", "text/html")
    w.Write(html)
  })
  http.HandleFunc("/ws/", edit);
  log.Printf("Listening on 127.0.0.1:" + port)
  log.Fatal(http.ListenAndServe(":" + port, nil))
}
