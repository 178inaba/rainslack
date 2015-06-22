package main

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/178inaba/rainimg"
	"github.com/BurntSushi/toml"
	"github.com/golang/glog"
	"github.com/nlopes/slack"
)

const (
	settingToml = "setting.toml"
)

var (
	s   setting
	api *slack.Slack
	o   sync.Once

	fileMap = make(map[string]slack.File)
)

type setting struct {
	token `toml:"token"`
}

type token struct {
	user string `toml:"user"`
	bot  string `toml:"bot"`
}

func init() {
	glog.Info("init()")

	flag.Parse()
}

func main() {
	glog.Info("main()")

	loadSetting()
	api = slack.New(s.token.user)

	userID, err := getUserID()
	if err != nil {
		os.Exit(1)
	}

	getFileList(userID)
	postRainImg()
}

func loadSetting() {
	glog.Info("loadSetting()")

	_, err := toml.DecodeFile(settingToml, &s)
	if err != nil {
		glog.Error("load error: ", err)
	}
}

func getUserID() (string, error) {
	glog.Info("getUserID()")

	info, err := api.AuthTest()
	if err != nil {
		glog.Error("AuthTest Error: ", err)
		return "", err
	}

	glog.Info("User: ", info.User)
	glog.Info("UserId: ", info.UserId)

	return info.UserId, nil
}

func getFileList(userID string) {
	glog.Info("getFileList()")

	searchParam := slack.NewGetFilesParameters()
	searchParam.UserId = userID

	files, _, _ := api.GetFiles(searchParam)

	glog.Info("filename list:")
	for _, file := range files {
		fileMap[file.Name] = file
		glog.Info(file.Name)
	}
}

func postRainImg() {
	glog.Info("postRainImg()")

	botAPI := slack.New(s.token.bot)
	sendCh := make(chan slack.OutgoingMessage)
	eventCh := make(chan slack.SlackEvent)

	ws, err := botAPI.StartRTM("", "https://slack.com/")
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
				f := rainImgUpload()
				sendCh <- *ws.NewOutgoingMessage(f.URL, msg.ChannelId)
			}
		case slack.LatencyReport:
			latency := event.Data.(slack.LatencyReport)
			glog.Info("ping latency: ", latency.Value)
		}
	}
}

func rainImgUpload() slack.File {
	glog.Info("rainImgUpload()")

	// create image
	fPath := rainimg.GetImgPath()

	// get filename
	fileName := filepath.Base(fPath)

	// already uploaded check
	file, ok := fileMap[fileName]
	if ok {
		glog.Info("already uploaded: ", file.Name)
		return file
	}

	// file up param
	var fup slack.FileUploadParameters
	fup.File = fPath

	// upload
	upFile, _ := api.UploadFile(fup)
	glog.Info("upload file: ", upFile.Name)

	// add file list
	fileMap[upFile.Name] = *upFile

	return *upFile
}
