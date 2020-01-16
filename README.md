# GT3-GOLang—SDK

## 概述
行为验证 golang sdk提供基于 `net/http` 的DEMO

## 集成


### 安装

直接试用命令行进行安装

`go get github.com/GeeTeam/gt3-golang-sdk`

### 接口示例

#### 验证初始化(API1)

```go
func registerGeetest(w http.ResponseWriter, r *http.Request) {

	geetest := geetest.NewGeetestLib(captchaID, privateKey, 2 * time.Second)
	status, response := geetest.PreProcess("", "")
	session, _ := store.Get(r, "geetest")
	session.Values["geetest_status"] = status
	session.Save(r, w)
	w.Write(response)
}
```
注意：初始化结果标识status（status=1表示初始化成功，status=0表示宕机状态）需要用户保存，在后续二次验证时会取出并进行逻辑判断。本SDK demo中使用`github.com/gorilla/sessions`来存取status。

#### 二次验证(API2)

```go
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
```
注意：
* 当取出status=0时表示极验宕机，此时流程进入failback模式，后续逻辑都是在您的服务器完成，不会再向极验服务器发送网络请求。本SDK demo中，对于failback模式，只对请求参数做了简单的校验，您也可以自行设计，模拟宕机模式：将验证ID替换为随意一段字符，例如123456789。此时，验证码将进入宕机模式。
* 此SDK仅支持验证3.0的failback模式，不支持验证2.0的failback模式

### 运行demo

1. git clone https://github.com/GeeTeam/gt3-golang-sdk.git
2. `cd gt3-golang-sdk/demo`
2. `go run main.go`
3. 在浏览器中访问**http://localhost:8888/static/login.html**





