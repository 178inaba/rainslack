package main

import (
	"flag"
	"github.com/178inaba/rainimg"
	"github.com/BurntSushi/toml"
	"github.com/golang/glog"
	"github.com/nlopes/slack"
	"regexp"
	"time"
)

const (
	settingToml = "setting.toml"
)

var (
	s Setting
)

type Setting struct {
	Token `toml:"token"`
}

type Token struct {
	User string `toml:"user"`
	Bot  string `toml:"bot"`
}

func init() {
	glog.Info("init()")

	flag.Parse()
}

func main() {
	glog.Info("main()")

	loadSetting()

	postRainImg()
}

func loadSetting() {
	_, err := toml.DecodeFile(settingToml, &s)
	if err != nil {
		glog.Error("load error: ", err)
	}
}

func postRainImg() {
	glog.Info("postRainImg()")

	botApi := slack.New(s.Token.Bot)
	sendCh := make(chan slack.OutgoingMessage)
	eventCh := make(chan slack.SlackEvent)

	ws, err := botApi.StartRTM("", "https://slack.com/")
	if err != nil {
		glog.Error(err)
		return
	}

	go ws.HandleIncomingEvents(eventCh)
	go ws.Keepalive(20 * time.Second)
	go func(ws *slack.SlackWS, sendCh <-chan slack.OutgoingMessage) {
		for {
			om := <-sendCh
			ws.SendMessage(&om)
		}
	}(ws, sendCh)

	for {
		event := <-eventCh
		switch event.Data.(type) {
		case *slack.MessageEvent:
			msg := event.Data.(*slack.MessageEvent)
			glog.Info("channel id: ", msg.ChannelId)
			glog.Info("text: ", msg.Text)

			match, _ := regexp.MatchString("雨", msg.Text)
			if match {
				f, _ := rainImgUpload()
				sendCh <- *ws.NewOutgoingMessage(f.URL, msg.ChannelId)
			}
		case slack.LatencyReport:
			latency := event.Data.(slack.LatencyReport)
			glog.Info("ping latency: ", latency.Value)
		}
	}
}

func rainImgUpload() (*slack.File, error) {
	glog.Info("rainImgUpload()")

	var fup slack.FileUploadParameters
	fup.File = rainimg.GetImgPath()

	return slack.New(s.Token.User).UploadFile(fup)
}
