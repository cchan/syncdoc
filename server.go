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
)

var upgrader = websocket.Upgrader{}

type change struct {
    Line    int32
    Ch      int32
    Added   []string
    Removed []string
}

type document struct {
    History []*change
    HistoryMutex *sync.Mutex
    Connections []*websocket.Conn
    ConnectionsMutex *sync.Mutex
}

func NewDocument() *document {
    d := new(document)
    d.History = make([]*change, 0)
    d.HistoryMutex = &sync.Mutex{}
    d.Connections = make([]*websocket.Conn, 0)
    d.ConnectionsMutex =  &sync.Mutex{}
    return d
}

func recvEdits(doc *document, c *websocket.Conn) {
  for {
    messageType, message, err := c.ReadMessage()
    if err != nil {
      log.Println("read:", err)
      break
    }

    log.Println("recv:", string(message))

    var changeObj change
    if err := json.Unmarshal(message, &changeObj); err != nil {
      log.Println("decode:", err)
      continue
    }

    for _, otherconn := range doc.Connections {
      if otherconn != c {
        stringified, err := json.Marshal(changeObj)
        if err != nil {
          log.Println("encode:", err)
          continue
        }
        if err := otherconn.WriteMessage(messageType, stringified); err != nil {
          log.Println("write:", err)
          break
        }
      }
    }
  }
}

var Documents = make(map[string]*document)

func edit(w http.ResponseWriter, r *http.Request) {
  c, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    log.Print("upgrade:", err)
    return
  }
  defer c.Close()

  if Documents[r.URL.RawPath] == nil {
    Documents[r.URL.RawPath] = NewDocument()
  }
  doc := Documents[r.URL.RawPath]

  doc.ConnectionsMutex.Lock()
  doc.Connections = append(doc.Connections, c)
  doc.ConnectionsMutex.Unlock()

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
