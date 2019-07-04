package geetest

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type GeetestLib struct {
	CaptchaID  string
	PrivateKey string
	Client     *http.Client
}

type FailbackRegisterRespnse struct {
	Success    int    `json:"success"`
	GT         string `json:"gt"`
	Challenge  string `json:"challenge"`
	NewCaptcha int    `json:"new_captcha"`
}

const (
	geetestHost = "http://api.geetest.com"
	registerURL = geetestHost + "/register.php"
	validateURL = geetestHost + "/validate.php"
)

func MD5Encode(input string) string {
	md5Instant := md5.New()
	md5Instant.Write([]byte(input))
	return hex.EncodeToString(md5Instant.Sum(nil))
}

func (g *GeetestLib) GetFailBackRegisterResponse(success int, challenge string) []byte {
	if challenge == "" {
		challenge = hex.EncodeToString(md5.New().Sum(nil))
	}

	response := FailbackRegisterRespnse{
		success,
		g.CaptchaID,
		challenge,
		1,
	}
	res, _ := json.Marshal(response)
	return res
}

func (g *GeetestLib) Do(req *http.Request) (body []byte, err error) {
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	var resp *http.Response
	if resp, err = g.Client.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusInternalServerError {
		err = errors.New("http status code 5xx")
		return
	}

	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}
	return
}

func (g *GeetestLib) PreProcess(userID string, userIP string) (int8, []byte) {
	params := url.Values{}
	params.Add("gt", g.CaptchaID)
	params.Add("new_captcha", "1")
	if userID != "" {
		params.Add("user_id", userID)
	}
	if userIP != "" {
		params.Add("ip_adress", userIP)
	}
	req, _ := http.NewRequest("GET", registerURL+"?"+params.Encode(), nil)
	body, err := g.Do(req)
	if err != nil {
		return 0, g.GetFailBackRegisterResponse(0, "")
	}
	challenge := string(body)
	if len(challenge) != 32 {
		return 0, g.GetFailBackRegisterResponse(0, "")
	} else {
		challenge = MD5Encode(challenge + g.PrivateKey)
		return 1, g.GetFailBackRegisterResponse(1, challenge)
	}
}

func (g *GeetestLib) CheckParas(challenge string, validate string, seccode string) bool {
	if challenge == "" || validate == "" || seccode == "" {
		return false
	}
	return true
}

func (g *GeetestLib) checkSuccessRes(challenge string, validate string) bool {
	return MD5Encode(g.PrivateKey+"geetest"+challenge) == validate
}

func (g *GeetestLib) checkFailbackRes(challenge string, validate string) bool {
	return MD5Encode(challenge) == validate
}

func (g *GeetestLib) SuccessValidate(challenge string, validate string, seccode string, userID string) bool {
	if !g.CheckParas(challenge, validate, seccode) {
		return false
	}
	if !g.checkSuccessRes(challenge, validate) {
		return false
	}
	params := url.Values{}
	params.Add("seccode", seccode)
	params.Add("challenge", challenge)
	params.Add("captchaid", g.CaptchaID)
	params.Add("sdk", "golang_v1.0.0")
	hehe := MD5Encode(g.PrivateKey + "geetest" + challenge)
	fmt.Println(hehe)
	if userID != "" {
		params.Add("user_id", userID)
	}
	req, _ := http.NewRequest("POST", validateURL, strings.NewReader(params.Encode()))
	body, err := g.Do(req)
	if err != nil {
		return false
	}
	res := string(body)
	return res == MD5Encode(seccode)
}

func (g *GeetestLib) FailbackValidate(challenge string, validate string, seccode string) bool {
	if !g.CheckParas(challenge, validate, seccode) {
		return false
	}
	if !g.checkFailbackRes(challenge, validate) {
		return false
	}
	return true
}
