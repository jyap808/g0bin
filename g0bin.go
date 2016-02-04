package main

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Paste struct {
	Expiration string
	Content    []byte
	UUID       string
}

var uuidValidator = regexp.MustCompile("^[a-zA-Z0-9-_]+$")

// CONFIG
type Config struct {
	Version               string
	CompressedStaticFiles bool
	Host                  string
	Port                  int
	MaxSize               int
}

var (
	config     *Config
	configLock = new(sync.RWMutex)
)

func init() {
	loadConfig(true)
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGHUP)
	go func() {
		for {
			<-s
			loadConfig(false)
			log.Print("Reloaded configuration")
		}
	}()
}

func GetConfig() *Config {
	configLock.RLock()
	defer configLock.RUnlock()
	return config
}

func loadConfig(fail bool) {
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Print("Open config error: ", err)
		if fail {
			os.Exit(1)
		}
	}

	temp := new(Config)
	if err = json.Unmarshal(file, temp); err != nil {
		log.Print("Parse config error: ", err)
		if fail {
			os.Exit(1)
		}
	}
	configLock.Lock()
	config = temp
	configLock.Unlock()
}

func newHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Settings *Config
	}{
		config,
	}
	t := template.Must(template.ParseFiles("templates/base.html", "templates/new.html"))
	t.Execute(w, data)
}

func abs(x int64) int64 {
	switch {
	case x < 0:
		return -x
	case x == 0:
		return 0 // return correctly abs(-0)
	}
	return x
}

func humanizedExpiration(since int64, t time.Time) string {
	abs_since := abs(since)
	if abs_since < 60 {
		return fmt.Sprintf("in %d s", abs_since)
	}

	if abs_since < 60*60 {
		mins_since := abs_since / 60
		return fmt.Sprintf("in %d m", mins_since)
	}

	if abs_since < 60*60*24 {
		hours_since := abs_since / (60 * 60)
		return fmt.Sprintf("in %d h", hours_since)
	}

	if abs_since < 60*60*24*10 {
		days_since := abs_since / (60 * 60 * 24)
		return fmt.Sprintf("in %d day(s)", days_since)
	}

	return fmt.Sprintf("on %s %d, %d", t.Month(), t.Day(), t.Year())
}

func pasteHandler(w http.ResponseWriter, r *http.Request) {
	paste_id := strings.TrimPrefix(r.URL.Path, "/paste/")
	keep_alive := false
	burn_after_reading := false
	humanized_expiration := ""

	if !uuidValidator.MatchString(paste_id) {
		log.Print("Invalue paste ID error")
		w.WriteHeader(http.StatusNotFound)
		data := struct {
			Settings *Config
		}{
			config,
		}
		t := template.Must(template.ParseFiles("templates/base.html", "templates/404.html"))
		t.Execute(w, data)
		return
	}

	paste := &Paste{UUID: paste_id}
	err := paste.load()
	if err != nil {
		log.Print("Load file error")
		w.WriteHeader(http.StatusNotFound)
		data := struct {
			Settings *Config
		}{
			config,
		}
		t := template.Must(template.ParseFiles("templates/base.html", "templates/404.html"))
		t.Execute(w, data)
		return
	}

	if strings.HasPrefix(paste.Expiration, "burn_after_reading") {
		// burn_after_reading contains the paste creation date
		// if this read appends 10 seconds after the creation date
		// we don't delete the paste because it means it's the redirection
		// to the paste that happens during the paste creation
		burn_after_reading = true
		keep_alive = true
		keep_alive_timestamp := strings.TrimPrefix(paste.Expiration, "burn_after_reading#")
		t, _ := time.Parse(time.RFC3339Nano, keep_alive_timestamp)
		since := int64(time.Since(t) / time.Second)
		if since > 10 {
			keep_alive = false
			paste.delete()
		}
	} else {
		keep_alive = true
		t, _ := time.Parse(time.RFC3339Nano, paste.Expiration)
		since := int64(time.Since(t) / time.Second)
		if since > 0 {
			keep_alive = false
			paste.delete()
			log.Print("Expiration error")
			w.WriteHeader(http.StatusNotFound)
			data := struct {
				Settings *Config
			}{
				config,
			}
			t := template.Must(template.ParseFiles("templates/base.html", "templates/404.html"))
			t.Execute(w, data)
			return
		} else {
			humanized_expiration = humanizedExpiration(since, t)
		}
	}

	data := struct {
		Settings            *Config
		Paste               *Paste
		KeepAlive           bool
		BurnAfterReading    bool
		HumanizedExpiration string
	}{
		config,
		paste,
		keep_alive,
		burn_after_reading,
		humanized_expiration,
	}
	t := template.Must(template.ParseFiles("templates/base.html", "templates/paste.html"))
	t.Execute(w, data)

}

func createHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		log.Print("Not a post. Redirecting")
		http.Redirect(w, r, "/new/", 302)
		return
	}
	expiration := r.FormValue("expiration")
	content := r.FormValue("content")

	// Create UUID from content
	h := sha1.New()
	h.Write([]byte(content))
	uuid := base64.URLEncoding.EncodeToString(h.Sum(nil))
	uuid = strings.TrimSuffix(uuid, "=")
	uuid = strings.Replace(uuid, "/", "-", -1)

	// Check paste content size
	if len(content) / 2 > config.MaxSize {
		w.Header().Set("Content-Type", "application/json")
		mapD := map[string]string{"status": "error", "message": "Content too big"}
		mapB, _ := json.Marshal(mapD)
		fmt.Fprintf(w, string(mapB))
	} else {
		// Save paste
		p := &Paste{Expiration: expiration, Content: []byte(content), UUID: uuid}
		p.save()

		// Response
		w.Header().Set("Content-Type", "application/json")
		mapD := map[string]string{"status": "ok", "paste": uuid}
		mapB, _ := json.Marshal(mapD)
		fmt.Fprintf(w, string(mapB))
	}
}

func (p *Paste) save() error {
	filename := p.UUID + ".txt"
	expiration := ""
	t := time.Now()
	if p.Expiration == "burn_after_reading" {
		expiration = fmt.Sprintf("burn_after_reading#%s", t.Format(time.RFC3339Nano))
	} else {
		if p.Expiration == "1_day" {
			expiration = time.Now().Add(24 * time.Hour).Format(time.RFC3339Nano)
		} else if p.Expiration == "1_month" {
			expiration = time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339Nano)
		} else if p.Expiration == "never" {
			expiration = time.Now().Add(100 * 365 * 24 * time.Hour).Format(time.RFC3339Nano)
		}
	}
	filecontent := fmt.Sprintf("%s\n%s\n", expiration, p.Content)
	return ioutil.WriteFile("./pastes/"+filename, []byte(filecontent), 0600)
}

func (p *Paste) load() error {
	// TODO
	// Handle malformed file
	// Handle cannot open file
	filename := p.UUID + ".txt"
	content, err := ioutil.ReadFile("./pastes/" + filename)
	if err != nil {
		return fmt.Errorf("Load file error")
	}
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if i == 0 {
			p.Expiration = line
		} else if i == 1 {
			p.Content = []byte(line)
		}
	}
	return nil
}

func (p *Paste) delete() error {
	filename := p.UUID + ".txt"
	return os.Remove("./pastes/" + filename)
}

func indexRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/new/", 302)
}

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func main() {
	http.HandleFunc("/", indexRedirect)
	http.HandleFunc("/new/", newHandler)
	http.HandleFunc("/paste/create", createHandler)
	http.HandleFunc("/paste/", pasteHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	log.Printf("Serving from http://%s:%d\n", config.Host, config.Port)
	err := http.ListenAndServe(config.Host+":"+strconv.Itoa(config.Port), Log(http.DefaultServeMux))
	if err != nil {
		log.Printf("ListenAndServe: ", err)
	}

}
