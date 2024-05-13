package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"github.com/anonyindian/logger"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/gotd/td/tg"
)

const DEBUG = false

var ValueOf = &config{}

type config struct {
	AppId             int    `json:"app_id"`
	ApiHash           string `json:"api_hash"`
	RedisUri          string `json:"redis_uri"`
	RedisPass         string `json:"redis_pass"`
	TestSessionString string `json:"test_session_string"`
	SessionString     string `json:"session_string"`
	SessionType       string `json:"session_type,omitempty"`
	HerokuApiKey      string `json:"heroku_api_key,omitempty"`
	HerokuAppName     string `json:"heroku_app_name,omitempty"`
	TestServer        bool   `json:"test_mode,omitempty"`
	BotToken          string `json:"bot_token,omitempty"`
}

var Self *tg.User

func Load(l *logger.Logger) {
	l = l.Create("CONFIG")
	defer l.ChangeLevel(logger.LevelMain).Println("LOADED")
	initPlatform()
	b, err := ioutil.ReadFile("config.json")
	if err != nil {
		if err := ValueOf.setupEnvVars(l); err != nil {
			l.ChangeLevel(logger.LevelError).Println(err.Error())
			os.Exit(1)
		}
		return
	}
	err = json.Unmarshal(b, ValueOf)
	if err != nil {
		l.ChangeLevel(logger.LevelCritical).Println("failed to load config:", err.Error())
		os.Exit(1)
	}
}

func getSessionString() string {
	if ValueOf.TestServer {
		return ValueOf.TestSessionString
	}
	return ValueOf.SessionString
}

func GetSession() sessionMaker.SessionConstructor {
	sessionString := getSessionString()
	const sessionName = "giga"
	switch strings.ToLower(ValueOf.SessionType) {
	case "pyrogram", "pyro":
		return sessionMaker.PyrogramSession(sessionString).Name(sessionName)
	case "gotgproto", "native":
		return sessionMaker.StringSession(sessionString).Name(sessionName)
	default:
		return sessionMaker.TelethonSession(sessionString).Name(sessionName)
	}
}
