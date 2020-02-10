// Plaintext OT.

// Try this later: https://medium.freecodecamp.org/million-websockets-and-go-cc58418460bb

package main

import (
  "net/http"
  "log"
  "strconv"
  "os"
  "github.com/gorilla/websocket"
  "strings"
  //"errors"
  "github.com/cchan/syncdoc/syncdoc"
  "github.com/gobwas/ws"
  "regexp"
)

var documents = make(map[string]*syncdoc.Syncdoc)

var upgrader = websocket.Upgrader{}

var validDocName = regexp.MustCompile("^[a-zA-Z0-9\\_\\-]+$")

func edit(w http.ResponseWriter, r *http.Request) {
  c, _, _, err := ws.UpgradeHTTP(r, w)
  if err != nil {
    log.Print("upgrade:", err)
    return
  }
  defer c.Close()

  if len(r.URL.Path) < 2 || r.URL.Path[0] != '/' { return }
  docname := strings.TrimPrefix(r.URL.Path, "/")
  if ! validDocName.Match([]byte(docname)) { return }

  if documents[docname] == nil {
    documents[docname] = syncdoc.NewDocument(docname)
  }
  doc := documents[docname]

  defer doc.RemoveConnection(c)
  doc.AddAndListen(c)
}

func main() {
  port := os.Getenv("PORT")
  if _, err := strconv.ParseInt(port, 10, 16); err != nil {
    port = "8080"
  }

  log.Printf("Listening on 127.0.0.1:" + port)
  log.Fatal(http.ListenAndServe(":" + port, http.HandlerFunc(edit)))
}
