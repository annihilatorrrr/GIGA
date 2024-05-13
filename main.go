package main

import (
	"flag"
	"os"
	"time"

	"github.com/anonyindian/logger"
	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/dispatcher/handlers/filters"
	"github.com/celestix/gotgproto/generic"
	"github.com/gigauserbot/giga/bot/helpmaker"
	"github.com/gigauserbot/giga/config"
	"github.com/gigauserbot/giga/db"
	"github.com/gigauserbot/giga/modules"
	"github.com/gigauserbot/giga/utils"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"
)

var (
	delay          = flag.Int("delay", 0, "")
	restartChatId  = flag.Int("chat", 0, "")
	restartMsgId   = flag.Int("msg_id", 0, "")
	restartMsgText = flag.String("msg_text", "", "")
)

func main() {
	flag.Parse()
	l := logger.New(os.Stderr, &logger.LoggerOpts{
		ProjectName: "GIGA-USERBOT",
	})
	if *restartMsgId != 0 {
		// Clean Console
		os.Stderr.Write([]byte("\n"))
	}

	if *delay != 0 {
		l.Println("Delaying start for", *delay, "seconds")
		time.Sleep(time.Second * time.Duration(*delay))
	}
	if config.DEBUG {
		l.ChangeMinimumLevel(logger.LevelDebug)
	}
	utils.InitUpdate(l)
	config.Load(l)
	handlers.DefaultPrefix = []rune{'.', '$'}
	db.Load(l)
	runClient(l)
}

func runClient(l *logger.Logger) {
	log := l.Create("CLIENT")

	client, err := gotgproto.NewClient(
		// Get AppID from https://my.telegram.org/apps
		config.ValueOf.AppId,
		// Get ApiHash from https://my.telegram.org/apps
		config.ValueOf.ApiHash,
		// ClientType, as we defined above
		gotgproto.ClientTypePhone(""),
		// Optional parameters of client
		&gotgproto.ClientOpts{
			Session:          config.GetSession(),
			DisableCopyright: true,
			DCList: func() (dct dcs.List) {
				if config.ValueOf.TestServer {
					dct = dcs.Test()
				}
				return
			}(),
		},
	)
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}
	log.ChangeLevel(logger.LevelInfo).Println("STARTED")
	utils.TelegramClient = client.Client
	config.Self = client.Self
	dispatcher := client.Dispatcher
	dispatcher.AddHandlerToGroup(handlers.NewMessage(filters.Message.Text, utils.GetBotToken(l)), 2)
	ctx := client.CreateContext()
	if *restartMsgId == 0 && *restartMsgText == "" {
		utils.StartupAutomations(l, ctx, client)
	} else {
		generic.EditMessage(ctx, *restartChatId, &tg.MessagesEditMessageRequest{
			ID:      *restartMsgId,
			Message: *restartMsgText,
		})
	}
	// Modules shall not be loaded unless the setup is complete
	modules.Load(l, dispatcher)
	helpmaker.MakeHelp()
	if config.ValueOf.TestServer {
		l.ChangeLevel(logger.LevelMain).Println("RUNNING ON TEST SERVER")
	}
	l.ChangeLevel(logger.LevelMain).Println("GIGA HAS BEEN STARTED")
	client.Idle()
}
