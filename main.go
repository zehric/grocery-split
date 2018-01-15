package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"golang.org/x/net/html"
	"html/template"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// TODO: reduce redundancy on cookies

type UploadInfo struct {
	Groceries  map[string]float64
	Html       template.HTML
	Creator    string
	N          uint8
	Total      float64
	SelectMode bool
}

type SubmitInfo struct {
	Unwanted map[string]StringSet
	Ready    StringSet
}

type ViewModel struct {
	Balances map[string]float64
	Creator  string
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
var templates = make(map[string]*template.Template)

var uploadInfo *UploadInfo = &UploadInfo{Groceries: make(map[string]float64)}
var submitInfo *SubmitInfo = &SubmitInfo{Unwanted: make(map[string]StringSet)}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates[tmpl].ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("username")
	if err != nil {
		renderTemplate(w, "login", nil)
	} else {
		http.Redirect(w, r, "/dash/", http.StatusFound)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		_, err := r.Cookie("username")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusFound)
		} else {
			http.Redirect(w, r, "/dash/", http.StatusFound)
		}
	} else {
		username := r.FormValue("username")
		expire := time.Now().AddDate(0, 0, 7)
		cookie := http.Cookie{Name: "username", Value: username, Path: "/", Expires: expire}
		http.SetCookie(w, &cookie)
		http.Redirect(w, r, "/dash/", http.StatusFound)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		cookie, err := r.Cookie("username")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		body := r.FormValue("body")
		doc, err := html.Parse(strings.NewReader(body))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		list, err := FindList(doc)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = AddToGroceries(list)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		b := new(bytes.Buffer)
		html.Render(b, list)
		uploadInfo.Html = template.HTML(b.String())
		n, err := strconv.ParseUint(r.FormValue("num"), 10, 8)
		uploadInfo.N = uint8(n)

		uploadInfo.Creator = cookie.Value
		uploadInfo.SelectMode = true

		makeGob(w, "upload", uploadInfo)
		makeGob(w, "submit", submitInfo)
	}
	http.Redirect(w, r, "/dash/", http.StatusFound)
}

func dashHandler(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("username")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if len(balances) > 0 {
		renderTemplate(w, "view", &ViewModel{balances, uploadInfo.Creator})
	} else if uploadInfo.SelectMode {
		renderTemplate(w, "list", uploadInfo)
	} else {
		renderTemplate(w, "upload", nil)
	}
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		cookie, err := r.Cookie("username")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		decoder := json.NewDecoder(r.Body)
		var body []string
		err = decoder.Decode(&body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
		userUnwanted := MakeSetFromSlice(body)
		user := cookie.Value
		for item, people := range submitInfo.Unwanted {
			if userUnwanted.Contains(item) {
				people.Add(user)
			} else {
				people.Remove(user)
			}
		}

		submitInfo.Ready.Add(cookie.Value)

		makeGob(w, "submit", submitInfo)

		if uint8(submitInfo.Ready.Length()) == uploadInfo.N {
			err = calculate()
			if err != nil {
				balances = make(map[string]float64)
				submitInfo.Ready.Remove(cookie.Value)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			makeGob(w, "balances", balances)
			w.Write([]byte("refresh"))
		}
	}
}

func makeGob(w http.ResponseWriter, name string, o interface{}) {
	file, err := os.Create("data/" + name + ".gob")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	encoder := gob.NewEncoder(file)
	encoder.Encode(o)
	file.Close()
}

func readGob(name string, o interface{}) error {
	file, err := os.Open("data/" + name + ".gob")
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	decoder.Decode(o)
	return nil
}

func readData() {
	_, err := os.Stat("data")
	if os.IsNotExist(err) {
		os.Mkdir("data", 0600)
	}
	if readGob("upload", &uploadInfo) != nil || readGob("submit", &submitInfo) != nil {
		return
	}
	readGob("balances", &balances)
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("username")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
	} else if cookie.Value != uploadInfo.Creator {
		http.Redirect(w, r, "/dash/", http.StatusFound)
	} else {
		os.RemoveAll("data/")
		os.Mkdir("data", 0600)
		uploadInfo = &UploadInfo{Groceries: make(map[string]float64)}
		submitInfo = &SubmitInfo{Unwanted: make(map[string]StringSet)}
		submitInfo.Ready = StringSet{make(map[string]struct{})}
		balances = make(map[string]float64)
		http.Redirect(w, r, "/dash/", http.StatusFound)
	}
}

func main() {
	templates["login"] = template.Must(template.ParseFiles("tmpl/login.html", "tmpl/base.html"))
	templates["upload"] = template.Must(template.ParseFiles("tmpl/upload.html", "tmpl/base.html"))
	templates["list"] = template.Must(template.ParseFiles("tmpl/list.html", "tmpl/base.html"))
	templates["view"] = template.Must(template.ParseFiles("tmpl/view.html", "tmpl/base.html"))

	submitInfo.Ready = StringSet{make(map[string]struct{})}

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/login/", loginHandler)
	http.HandleFunc("/upload/", uploadHandler)
	http.HandleFunc("/dash/", dashHandler)
	http.HandleFunc("/submit/", submitHandler)
	http.HandleFunc("/reset/", resetHandler)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	readData()

	http.ListenAndServe(":80", nil)
}
