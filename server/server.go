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
  "github.com/cchan/syncdoc/syncdoc"
  "regexp"
)

var Documents = make(map[string]*syncdoc.Syncdoc)

var upgrader = websocket.Upgrader{}

var validDocName = regexp.MustCompile("^[a-zA-Z\\_\\-]+$")
func invalidDocName(w http.ResponseWriter) {
  http.Error(w, "Invalid document URL - can only contain alphanumeric + underscore.", http.StatusBadRequest)
}

func edit(w http.ResponseWriter, r *http.Request) {
  c, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    log.Print("upgrade:", err)
    return
  }
  defer c.Close()

  if r.URL.Path[:4] != "/ws/" { invalidDocName(w); return }
  docname := strings.TrimPrefix(r.URL.Path, "/ws/")

  if ! validDocName.Match([]byte(docname)) { invalidDocName(w); return }

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

  indexhtml, err := ioutil.ReadFile("static/index.html")
  if err != nil { log.Fatal("Could not open index.html for reading") }
  apphtml, err := ioutil.ReadFile("static/app.html")
  if err != nil { log.Fatal("Could not open app.html for reading") }

  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    if r.URL.Path == "/" { w.Write(indexhtml); return }
    if r.URL.Path[:1] != "/" { invalidDocName(w); return }
    docname := strings.TrimPrefix(r.URL.Path, "/")
    if ! validDocName.Match([]byte(docname)) { invalidDocName(w); return }
    w.Write(apphtml)
  })

  http.HandleFunc("/ws/", edit);

  log.Printf("Listening on 127.0.0.1:" + port)
  log.Fatal(http.ListenAndServe(":" + port, nil))
}
