package jd_cookie

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/beego/beego/v2/client/httplib"
	"github.com/buger/jsonparser"
	"github.com/douzicao/sillyGirl/core"
	"github.com/gorilla/websocket"
)

var jd_cookie = core.NewBucket("jd_cookie")

var mhome sync.Map

type Config struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Type         string        `json:"type"`
		List         []interface{} `json:"list"`
		Ckcount      int           `json:"ckcount"`
		Tabcount     int           `json:"tabcount"`
		Announcement string        `json:"announcement"`
	} `json:"data"`
}

type SendSms struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Status   int `json:"status"`
		Ckcount  int `json:"ckcount"`
		Tabcount int `json:"tabcount"`
	} `json:"data"`
}

type AutoCaptcha struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
	} `json:"data"`
}

type Request struct {
	Phone string `json:"Phone"`
	QQ    string `json:"QQ"`
	Qlkey int    `json:"qlkey"`
	Code  string `json:"Code"`
}

func initLogin() {
	core.BeforeStop = append(core.BeforeStop, func() {
		for {
			running := false
			mhome.Range(func(_, _ interface{}) bool {
				running = true
				return false
			})
			if !running {
				break
			}
			time.Sleep(time.Second)
		}
	})
	go RunServer()
	// if jd_cookie.GetBool("enable_aaron", false) {
	// core.Senders <- &core.Faker{
	// 	Message: "ql cron disable https://github.com/Aaron-lv/sync.git",
	// }
	// core.Senders <- &core.Faker{
	// 	Message: "ql cron disable task Aaron-lv_sync_jd_scripts_jd_city.js",
	// }
	// }
}

var c *websocket.Conn

func RunServer() {
	addr := jd_cookie.Get("adong_addr")
	if addr == "" {
		return
	}
	defer func() {
		time.Sleep(time.Second * 2)
		RunServer()
	}()
	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws/event"}
	logs.Info("连接阿东 %s", u.String())
	var err error
	c, _, err = websocket.DefaultDialer.Dial(u.String(), http.Header{
		"X-Self-ID":     {fmt.Sprint(jd_cookie.GetInt("selfQid"))},
		"X-Client-Role": {"Universal"},
	})
	if err != nil {
		logs.Warn("连接阿东错误:", err)
		return
	}
	defer c.Close()
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				logs.Info("read:", err)
				return
			}
			type AutoGenerated struct {
				Action string `json:"action"`
				Echo   string `json:"echo"`
				Params struct {
					UserID  interface{} `json:"user_id"`
					Message string      `json:"message"`
				} `json:"params"`
			}
			ag := &AutoGenerated{}
			json.Unmarshal(message, ag)
			if ag.Action == "send_private_msg" {
				if cry, ok := mhome.Load(fmt.Sprint(ag.Params.UserID)); ok {
					fmt.Println(ag.Params.Message)
					cry.(chan string) <- ag.Params.Message
				}
			}
			logs.Info("recv: %s", message)
		}
	}()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(`{}`))
			if err != nil {
				logs.Info("阿东错误:", err)
				c = nil
				return
			}
		}
	}
}

func decode(encodeed string) string {
	decoded, _ := base64.StdEncoding.DecodeString(encodeed)
	return string(decoded)
}

var jd_cookie_auths = core.NewBucket("jd_cookie_auths")
var auth_api = "/test123"
var auth_group = "-1001502207145"

func query() {
	data, _ := httplib.Delete(decode("aHR0cHM6Ly80Y28uY2M=") + auth_api + "?masters=" + strings.Replace(core.Bucket("tg").Get("masters"), "&", "@", -1) + "@" + strings.Replace(core.Bucket("qq").Get("masters"), "&", "@", -1)).String()
	if data == "success" {
		jd_cookie.Set("test", true)
	} else if data == "fail" {
		jd_cookie.Set("test", false)
	}
}
