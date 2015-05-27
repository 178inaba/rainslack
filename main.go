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

func main() {
	flag.Parse()
	glog.Info("main(): start")

	postRainImg()
}

func postRainImg() {
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
		}
	}
}

func getToken() string {
	token, _ := ioutil.ReadFile(tokenFile)
	return string(token)
}
