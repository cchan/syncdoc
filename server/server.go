// Plaintext OT.

// Try this later: https://medium.freecodecamp.org/million-websockets-and-go-cc58418460bb

package main

import (
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"

	"github.com/cchan/syncdoc/syncdoc"
	"github.com/gobwas/ws"
	"github.com/valyala/fasttemplate"
)

var (
	documents   = make(map[string]*syncdoc.Syncdoc)
	documentsMu sync.Mutex
)

// Limits so a flood of connections can't exhaust memory on the host.
const (
	maxOpenDocs    = 500
	maxConnsPerDoc = 50
)

// acquireDocument returns the in-memory document, loading it if needed, and counts the caller as a user.
func acquireDocument(name string) *syncdoc.Syncdoc {
	documentsMu.Lock()
	defer documentsMu.Unlock()
	doc := documents[name]
	if doc == nil {
		if len(documents) >= maxOpenDocs {
			return nil
		}
		doc = syncdoc.NewDocument(name)
		documents[name] = doc
	}
	if doc.Users >= maxConnsPerDoc {
		return nil
	}
	doc.Users++
	return doc
}

// releaseDocument drops the document from memory once nobody is using it; its state is already on disk.
func releaseDocument(name string, doc *syncdoc.Syncdoc) {
	documentsMu.Lock()
	defer documentsMu.Unlock()
	doc.Users--
	if doc.Users == 0 && documents[name] == doc {
		delete(documents, name)
	}
}

var validDocName = regexp.MustCompile("^[a-zA-Z0-9\\_\\-]+$")

var appTemplate *fasttemplate.Template

func edit(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path) < 4 || r.URL.Path[:4] != "/ws/" {
		docname := r.URL.Path[1:]
		if !validDocName.Match([]byte(docname)) {
			return
		}
		plaintext, err := ioutil.ReadFile("/home/www/data/" + docname + ".file")
		if err != nil {
			plaintext = []byte{}
		}

		// This was originally in main but was moved here so I don't have to reload the service every time I make a change.
		// It should be moved back for prod.
		appTemplateContent, err := ioutil.ReadFile("/home/www/go/src/github.com/cchan/syncdoc/static/app.html")
		if err != nil {
			return //panic(err)
		}
		appTemplate = fasttemplate.New(string(appTemplateContent), "[[[", "]]]")

		_, err = appTemplate.Execute(w, map[string]interface{}{
			"content": html.EscapeString(string(plaintext)),
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
	if !validDocName.Match([]byte(docname)) {
		return
	}

	doc := acquireDocument(docname)
	if doc == nil {
		return
	}
	defer releaseDocument(docname, doc)
	defer doc.RemoveConnection(c)
	doc.AddAndListen(c)
}

func main() {
	port := os.Getenv("PORT")
	if _, err := strconv.ParseInt(port, 10, 16); err != nil {
		port = "8080"
	}

	log.Printf("Listening on 127.0.0.1:" + port + "\n")
	log.Fatal(http.ListenAndServe("127.0.0.1:"+port, http.HandlerFunc(edit)))
}
