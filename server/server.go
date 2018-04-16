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
  "strings"
  //"errors"
  "github.com/cchan/operational-transformation/syncdoc"
)

// TODO:
// - separate this into neater files
// - periodically dump History results to plaintext file, and use that as the base
//   - keep enough history (1000 entries?) that it can still merge any latecomers
//   - full-file updates periodically (every 10s?)
// - OT is pretty easy; to test it just give a setTimeout delay before js send edit to server
// - collapse history entries (hard to preserve indexes) - should probably happen mainly on js side? for now don't worry about load
// - WSS LetsEncrypt
// - Use a linkedlist for Document connections?

var Documents = make(map[string]*syncdoc.Syncdoc)

var upgrader = websocket.Upgrader{}

func edit(w http.ResponseWriter, r *http.Request) {
  c, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    log.Print("upgrade:", err)
    return
  }
  defer c.Close()

  docname := strings.TrimPrefix(string(r.URL.Path), "/ws/")

  if Documents[docname] == nil {
    Documents[docname] = syncdoc.NewDocument(docname)
  }
  doc := Documents[docname]

  doc.AddConnection(c)
  defer doc.RemoveConnection(c)

  doc.Listen(c)
}

func main() {
  port := os.Getenv("PORT")
  if _, err := strconv.ParseInt(port, 10, 16); err != nil {
    port = "8080"
  }

  html, err := ioutil.ReadFile("static/index.html")
  if err != nil { log.Fatal("Could not open index.html for reading") }

  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    w.Write(html)
  })

  http.HandleFunc("/ws/", edit);

  log.Printf("Listening on 127.0.0.1:" + port)
  log.Fatal(http.ListenAndServe(":" + port, nil))
}
