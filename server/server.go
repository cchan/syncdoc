// Plaintext OT.

// Try this later: https://medium.freecodecamp.org/million-websockets-and-go-cc58418460bb

package main

import (
  "net/http"
  "log"
  "strconv"
  "os"
  "strings"
  //"errors"
  "github.com/cchan/syncdoc/syncdoc"
  "github.com/gobwas/ws"
  "regexp"
  "html"
  "github.com/valyala/fasttemplate"
)

var documents = make(map[string]*syncdoc.Syncdoc)
// TODO: ADD RWMUTEX HERE - map is not concurrent write safe! but is read safe.

var validDocName = regexp.MustCompile("^[a-zA-Z0-9\\_\\-]+$")

var appTemplateContent, err = ioutil.ReadFile("/home/www/go/src/github.com/cchan/syncdoc/static/app.html")
if err != nil { panic(err) }
var appTemplate = fasttemplate.New(appTemplateContent, "{", "}")

func edit(w http.ResponseWriter, r *http.Request) {
  if len(r.URL.Path) < 4 || r.URL.Path[:4] != "/ws/" {
    docname := r.URL.Path[1:]
    if ! validDocName.Match([]byte(docname)) { return }
    plaintext, err := ioutil.ReadFile("/home/www/data/" + pathname + ".file")
    if err != nil { plaintext = "" }
    appTemplate.Execute(w, html.EscapeString(plaintext), map[string]interface{}{
      "content": plaintext
    })
    return
  }

  c, _, _, err := ws.UpgradeHTTP(r, w)
  if err != nil {
    log.Println("upgrade:", err)
    return
  }
  defer c.Close()

  docname := r.URL.Path[4:]
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

  log.Printf("Listening on 127.0.0.1:" + port + "\n")
  log.Fatal(http.ListenAndServe("127.0.0.1:" + port, http.HandlerFunc(edit)))
}
