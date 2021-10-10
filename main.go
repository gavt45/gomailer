package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/flosch/pongo2"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

func checkErr(error error) {
	if error != nil {
		log.Fatal("Error: ", error)
	}
}

func genUid() string {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	checkErr(err)
	return fmt.Sprintf("%x", key)
}

var unRestrictedPaths = map[string]bool{
	"/status": false, // == restricted path
	"/send":   false, // == restricted path
	"/":       true,  // == un restricted path
}

type TaskResult struct {
	//uid		string
	Code    int    `json:"code"`
	Message string `json:"message"`
}

var Tasks = struct {
	sync.RWMutex
	m map[string]TaskResult
}{m: make(map[string]TaskResult)}

var templateGen = TemplateEng{
	m: make(map[string]*pongo2.Template),
}

type response struct {
	Code    int    `json:"code"`
	Error   string `json:"error"`
	Message string `json:"message"`
	Uid     string `json:"uid"`
}

type badAuth struct {
	Secret string
}

func (b *badAuth) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	username := r.URL.Query().Get("secret")
	if username != b.Secret && !unRestrictedPaths[r.URL.Path] {
		resp := response{
			Code:    1,
			Error:   "Unauthorized",
			Message: "Unauthorized",
			Uid:     "",
		}
		respBytes, err := json.Marshal(resp)
		checkErr(err)
		http.Error(w, string(respBytes), 401)
		return
	}
	ctx := context.WithValue(r.Context(), "secret", username)
	r = r.WithContext(ctx)
	next(w, r)
}

func sendMail(w http.ResponseWriter, r *http.Request) {
	var mailRe = regexp.MustCompile(`(?:[a-z0-9!#$%&'*+/=?^_{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])`)
	var template_regexp = regexp.MustCompile(`^[a-z0-9]+$`)
	to := r.URL.Query().Get("to")
	subject := r.URL.Query().Get("subject")
	template_name := r.URL.Query().Get("template_name")
	field_mames := r.URL.Query().Get("field_names")

	resp := response{
		Code:    0,
		Error:   "",
		Message: "OK",
		Uid:     "",
	}

	if to == "" || subject == "" || template_name == "" || field_mames == "" {
		resp.Code = 1
		resp.Error = "You should pass to, subject and body parameters!"
		resp.Message = resp.Error
		respBytes, err := json.Marshal(resp)
		checkErr(err)
		fmt.Fprintf(w, string(respBytes))
		return
	}

	if !mailRe.MatchString(to) || !template_regexp.MatchString(template_name) {
		resp.Code = 1
		resp.Error = "email not matches RFC regex or template name is strange!"
		resp.Message = resp.Error
		respBytes, err := json.Marshal(resp)
		checkErr(err)
		fmt.Fprintf(w, string(respBytes))
		return
	}

	var args_map = make(map[string]string)

	for _, field_name := range strings.Split(field_mames, ",") {
		tmp := r.URL.Query().Get(field_name)
		if tmp == "" {
			resp.Code = 1
			resp.Error = "Field is in fields, but not in query!"
			resp.Message = resp.Error
			respBytes, err := json.Marshal(resp)
			checkErr(err)
			fmt.Fprintf(w, string(respBytes))
			return
		}
		args_map[field_name] = tmp
	}

	bodyBytes, err := templateGen.execTemplate(template_name, args_map)

	if err != nil {
		resp.Code = 1
		resp.Error = "Body is not a valid base64!"
		resp.Message = resp.Error
		respBytes, err := json.Marshal(resp)
		checkErr(err)
		fmt.Fprintf(w, string(respBytes))
		return
	}

	var uid = genUid()
	resp.Uid = uid

	go Send(to, subject, bodyBytes, uid, config)

	//if err != nil {
	//	resp.Code = 1
	//	resp.Error = "smtp error: "+err.Error()
	//	resp.Message = resp.Error
	//}

	respBytes, err := json.Marshal(resp)
	checkErr(err)
	fmt.Fprintf(w, string(respBytes))
}

func status(w http.ResponseWriter, r *http.Request) {
	//username := r.Context().Value("username").(string)
	uid := r.URL.Query().Get("uid")
	resp := response{
		Code:    0,
		Error:   "",
		Message: "OK",
		Uid:     uid,
	}
	if uid == "" || len(uid) != 64 {
		resp.Code = 1
		resp.Error = "Invalid uid"
		resp.Message = resp.Error
		respBytes, err := json.Marshal(resp)
		checkErr(err)
		fmt.Fprintf(w, string(respBytes))
		return
	}

	Tasks.RLock()
	taskRes := Tasks.m[uid]
	Tasks.RUnlock()

	Tasks.Lock()
	delete(Tasks.m, uid)
	Tasks.Unlock()

	if taskRes.Message == "" {
		resp.Code = 1
		resp.Error = "Mail not yet sent or no such task"
		resp.Message = resp.Error
		respBytes, err := json.Marshal(resp)
		checkErr(err)
		fmt.Fprintf(w, string(respBytes))
		return
	}

	resp.Code = taskRes.Code
	resp.Message = taskRes.Message
	if taskRes.Message != "OK" {
		resp.Error = taskRes.Message
	}

	respBytes, err := json.Marshal(resp)
	checkErr(err)
	fmt.Fprintf(w, string(respBytes))
}

func root(w http.ResponseWriter, r *http.Request) {
	//username := r.Context().Value("username").(string)
	resp := response{
		Code:    0,
		Error:   "",
		Message: "OK",
	}
	respBytes, err := json.Marshal(resp)
	checkErr(err)
	fmt.Fprintf(w, string(respBytes))
}

func setupEndpoints(router *mux.Router) {
	router.HandleFunc("/status", status).Methods("GET")
	router.HandleFunc("/send", sendMail).Methods("GET")
	router.HandleFunc("/", root).Methods("GET")
}

var config = new(Config)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: ", os.Args[0], " <command start/cfg>")
	}

	config.SmtpAddr = os.Getenv("SMTP_ADDR")
	config.Password = os.Getenv("SMTP_PASSWORD")
	config.Email = os.Getenv("EMAIL")
	config.Port, _ = strconv.Atoi(os.Getenv("PORT"))
	config.Secret = os.Getenv("SECRET")
	config.ServerCert = ""
	config.ServerKey = ""
	config.TemplatePath = os.Getenv("TEMPLATE_PATH")

	if config.Port == 0 || config.SmtpAddr == "" || config.Password == "" || config.Email == "" || config.Secret == "" || config.TemplatePath == "" {
		panic("config is invalid!")
	}

	switch os.Args[1] { // command
	//case "cfg":
	//	confStr, err := json.Marshal(config)
	//	checkErr(err)
	//	err = ioutil.WriteFile(os.Args[2], confStr, 0644)
	//	checkErr(err)
	case "start":
		//configFileContent, err := ioutil.ReadFile(os.Args[2])
		//checkErr(err)
		//err = json.Unmarshal(configFileContent, config)
		//checkErr(err)

		templateGen.loadTemplates(config.TemplatePath)

		r := mux.NewRouter()
		setupEndpoints(r)
		n := negroni.Classic()
		n.Use(&badAuth{
			Secret: config.Secret,
		})
		n.UseHandler(r)
		log.Println("Starting server on port ", config.Port, "...")
		err := http.ListenAndServe(":"+strconv.Itoa(config.Port), n)
		//err = http.ListenAndServeTLS(":"+strconv.Itoa(config.Port), config.ServerCert, config.ServerKey, n)
		checkErr(err)
	default:
		log.Fatal("No such command: ", os.Args[1])
	}
}
