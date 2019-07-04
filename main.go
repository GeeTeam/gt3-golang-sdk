package main

import (
	"encoding/json"
	"github.com/gorilla/sessions"
	"gt3-golang-sdk/geetest"
	"net/http"
	"time"
)

const (
	captchaID  = "48a6ebac4ebc6642d68c217fca33eb4d"
	privateKey = "4f1c085290bec5afdc54df73535fc361"
)

var store = sessions.NewCookieStore([]byte("geetestdemo"))

func registerGeetest(w http.ResponseWriter, r *http.Request) {

	geetest := geetest.NewGeetestLib(captchaID, privateKey, 2 * time.Second)
	status, response := geetest.PreProcess("", "")
	session, _ := store.Get(r, "geetest")
	session.Values["geetest_status"] = status
	session.Save(r, w)
	w.Write(response)
}

func validateGeetest(w http.ResponseWriter, r *http.Request) {
	var geetestRes bool
	r.ParseForm()
	geetest := geetest.NewGeetestLib(captchaID, privateKey, 2 * time.Second)
	res := make(map[string]interface{})
	session, _ := store.Get(r, "geetest")
	challenge := r.Form.Get("geetest_challenge")
	validate := r.Form.Get("geetest_validate")
	seccode := r.Form.Get("geetest_seccode")
	val := session.Values["geetest_status"]
	status := val.(int8)
	if status == 1 {
		geetestRes = geetest.SuccessValidate(challenge, validate, seccode, "", "")
	} else {
		geetestRes = geetest.FailbackValidate(challenge, validate, seccode)
	}
	if geetestRes {
		res["code"] = 0
		res["msg"] = "Success"
	} else {
		res["code"] = -100
		res["msg"] = "Failed"
	}
	response, _ := json.Marshal(res)
	w.Write(response)
}

func main() {
	fsh := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fsh))
	http.HandleFunc("/gt/preprocess", registerGeetest)
	http.HandleFunc("/gt/validate", validateGeetest)
	if err := http.ListenAndServe(":8888", nil); err != nil {
		panic(err)
	}
}
