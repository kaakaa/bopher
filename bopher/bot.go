// Copyright (c) 2016 Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package bopher

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mattermost/platform/model"
)

var client *model.Client
var webSocketClient *model.WebSocketClient

var botUser *model.User
var botTeam *model.Team
var initialLoad *model.InitialLoad
var debuggingChannel *model.Channel
var service *Service

// Documentation for the Go driver can be found
// at https://godoc.org/github.com/mattermost/platform/model#Client
func RunBot(config Config) {
	bin := filepath.FromSlash(os.ExpandEnv(config.Gopher))
	service = &Service{
		Gopher:       bin,
		MaxOfGophers: config.MaxOfGophers,
	}
	SetupGracefulShutdown()

	url := config.Mattermost.HttpURL()
	fmt.Println("Ping: " + url)
	client = model.NewClient(url)

	// Lets test to see if the mattermost server is up and running
	MakeSureServerIsRunning()

	// lets attempt to login to the Mattermost server as the bot user
	// This will set the token required for all future calls
	// You can get this token with client.AuthToken
	LoginAsTheBotUser(config.Mattermost.Bot)

	// If the bot user doesn't have the correct information lets update his profile
	UpdateTheBotUserIfNeeded(config.Mattermost.Bot)

	// Lets load all the stuff we might need
	InitialLoad()

	// Lets find our bot team
	FindBotTeam(config.Mattermost.Team)

	// This is an important step.  Lets make sure we use the botTeam
	// for all future web service requests that require a team.
	client.SetTeamId(botTeam.Id)

	// Lets create a bot channel for logging debug messages into
	CreateBotDebuggingChannelIfNeeded(config.Mattermost.Channel)
	SendMsgToDebuggingChannel("_Bot has **started** running_", "")

	// Lets start listening to some channels via the websocket!
	wsUrl := fmt.Sprintf(config.Mattermost.WsURL())
	webSocketClient, err := model.NewWebSocketClient(wsUrl, client.AuthToken)
	if err != nil {
		println("We failed to connect to the web socket")
		PrintError(err)
	}

	webSocketClient.Listen()

	go func() {
		for {
			select {
			case resp := <-webSocketClient.EventChannel:
				HandleWebSocketResponse(resp)
			}
		}
	}()

	// You can block forever with
	select {}
}

func MakeSureServerIsRunning() {
	if props, err := client.GetPing(); err != nil {
		println("There was a problem pinging the Mattermost server.  Are you sure it's running?")
		PrintError(err)
		os.Exit(1)
	} else {
		println("Server detected and is running version " + props["version"])
	}
}

func LoginAsTheBotUser(bot Bot) {
	if loginResult, err := client.Login(bot.Email, bot.Password); err != nil {
		println("There was a problem logging into the Mattermost server.  Are you sure ran the setup steps from the README.md?")
		PrintError(err)
		os.Exit(1)
	} else {
		botUser = loginResult.Data.(*model.User)
	}
}

func UpdateTheBotUserIfNeeded(bot Bot) {
	if botUser.FirstName != bot.First || botUser.LastName != bot.Last || botUser.Username != bot.Name {
		botUser.FirstName = bot.First
		botUser.LastName = bot.Last
		botUser.Username = bot.Name

		if updateUserResult, err := client.UpdateUser(botUser); err != nil {
			println("We failed to update the Sample Bot user")
			PrintError(err)
			os.Exit(1)
		} else {
			botUser = updateUserResult.Data.(*model.User)
			println("Looks like this might be the first run so we've updated the bots account settings")
		}
	}
}

func InitialLoad() {
	if initialLoadResults, err := client.GetInitialLoad(); err != nil {
		println("We failed to get the initial load")
		PrintError(err)
		os.Exit(1)
	} else {
		initialLoad = initialLoadResults.Data.(*model.InitialLoad)
	}
}

func FindBotTeam(t string) {
	for _, team := range initialLoad.Teams {
		if team.Name == t {
			botTeam = team
			break
		}
	}

	if botTeam == nil {
		println("We do not appear to be a member of the team '" + t + "'")
		os.Exit(1)
	}
}

func CreateBotDebuggingChannelIfNeeded(ch Channel) {
	if channelsResult, err := client.GetChannels(""); err != nil {
		println("We failed to get the channels")
		PrintError(err)
	} else {
		channelList := channelsResult.Data.(*model.ChannelList)
		for _, channel := range channelList.Channels {

			// The logging channel has alredy been created, lets just use it
			if channel.Name == ch.Name {
				debuggingChannel = channel
				return
			}
		}
	}

	// Looks like we need to create the logging channel
	channel := &model.Channel{}
	channel.Name = ch.Name
	channel.DisplayName = ch.DisplayName
	channel.Purpose = ch.Purpose
	channel.Type = model.CHANNEL_OPEN
	if channelResult, err := client.CreateChannel(channel); err != nil {
		println("We failed to create the channel " + ch.Name)
		PrintError(err)
	} else {
		debuggingChannel = channelResult.Data.(*model.Channel)
		println("Looks like this might be the first run so we've created the channel " + ch.Name)
	}
}

func SendMsgToDebuggingChannel(msg string, replyToId string) {
	post := &model.Post{}
	post.ChannelId = debuggingChannel.Id
	post.Message = msg

	post.RootId = replyToId

	if _, err := client.CreatePost(post); err != nil {
		println("We failed to send a message to the logging channel")
		PrintError(err)
	}
}

func HandleWebSocketResponse(event *model.WebSocketEvent) {
	HandleMsgFromDebuggingChannel(event)
}

func HandleMsgFromDebuggingChannel(event *model.WebSocketEvent) {
	// If this isn't the debugging channel then lets ingore it
	if event.ChannelId != debuggingChannel.Id {
		return
	}

	// Lets only reponded to messaged posted events
	if event.Event != model.WEBSOCKET_EVENT_POSTED {
		return
	}

	// Lets ignore if it's my own events just in case
	if event.UserId == botUser.Id {
		return
	}

	println("responding to debugging channel msg")

	post := model.PostFromJson(strings.NewReader(event.Data["post"].(string)))
	if post != nil {

		// if you see any word matching 'alive' then respond
		if matched, _ := regexp.MatchString(`(?:^|\W)alive(?:$|\W)`, post.Message); matched {
			SendMsgToDebuggingChannel("Yes I'm running", post.Id)
			return
		}

		// if you see any word matching 'byeGopher' then respond
		if matched, _ := regexp.MatchString(`(?:^|\W)bye gopher(?:$|\W)`, post.Message); matched {
			service.byeGopher()
			return
		}

		// if you see any word matching 'service' then respond
		if matched, _ := regexp.MatchString(`(?:^|\W)hello gopher(?:$|\W)`, post.Message); matched {
			service.sayGopher("Hello!")
			return
		}

		// if you see any word matching 'service' then respond
		if matched, _ := regexp.MatchString(`(?:^|\W)jump gopher(?:$|\W)`, post.Message); matched {
			service.jumpGopher()
			return
		}

		// if you see any word matching 'service' then respond
		if matched, _ := regexp.MatchString(`(?:^|\W)whats gopher?(?:$|\W)`, post.Message); matched {
			for i := 0; i < 5; i++ {
				service.goGopher()
			}
			time.Sleep(5 * time.Second)
			service.sayGopher("Hi! I'm Gopher!")
			return
		}

		// if you see any word matching 'service' then respond
		if matched, _ := regexp.MatchString(`(?:^|\W)gopher(?:$|\W)`, post.Message); matched {
			service.goGopher()
			return
		}

	}

	SendMsgToDebuggingChannel("I did not understand you!", post.Id)
}

func PrintError(err *model.AppError) {
	println("\tError Details:")
	println("\t\t" + err.Message)
	println("\t\t" + err.Id)
	println("\t\t" + err.DetailedError)
}

func SetupGracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			if webSocketClient != nil {
				webSocketClient.Close()
			}

			SendMsgToDebuggingChannel("_Bot has **stopped** running_", "")
			os.Exit(0)
		}
	}()
}
