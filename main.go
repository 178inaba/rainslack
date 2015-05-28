package main

import (
	"flag"
	"github.com/178inaba/rainimg"
	"github.com/golang/glog"
	"github.com/nlopes/slack"
	"io/ioutil"
	"regexp"
	"time"
)

const (
	tokenFile = "token_file"
)

func init() {
	glog.Info("init()")

	flag.Parse()
}

func main() {
	glog.Info("main()")

	postRainImg()
}

func postRainImg() {
	glog.Info("postRainImg()")

	api := slack.New(getToken())
	eventCh := make(chan slack.SlackEvent)

	ws, err := api.StartRTM("", "https://slack.com/")
	if err != nil {
		glog.Error(err)
	}

	go ws.HandleIncomingEvents(eventCh)
	go ws.Keepalive(20 * time.Second)

	for {
		event := <-eventCh
		switch event.Data.(type) {
		case *slack.MessageEvent:
			msg := event.Data.(*slack.MessageEvent)
			glog.Info("channel id: ", msg.ChannelId)
			glog.Info("text: ", msg.Text)

			match, _ := regexp.MatchString("雨", msg.Text)
			if match {
				// file upload
				var fup slack.FileUploadParameters
				fup.File = rainimg.GetImgPath()
				f, _ := api.UploadFile(fup)

				// post message
				p := slack.NewPostMessageParameters()
				p.Username = "now rain"
				p.IconEmoji = ":rainbow:"
				api.PostMessage("#your_ch", f.URL, p)
			}
		case slack.LatencyReport:
			latency := event.Data.(slack.LatencyReport)
			glog.Info("pong latency: ", latency.Value)
		}
	}
}

func getToken() string {
	token, _ := ioutil.ReadFile(tokenFile)
	return string(token)
}
