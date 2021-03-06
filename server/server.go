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

	"github.com/cchan/syncdoc/syncdoc"
	"github.com/gobwas/ws"
	"github.com/valyala/fasttemplate"
)

var documents = make(map[string]*syncdoc.Syncdoc)

// TODO: ADD RWMUTEX HERE - map is not concurrent write safe! but is read safe.

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
	log.Fatal(http.ListenAndServe("127.0.0.1:"+port, http.HandlerFunc(edit)))
}
