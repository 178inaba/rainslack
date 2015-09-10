package main

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"sync"

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
	api *slack.Client
	o   sync.Once

	fileMap = make(map[string]slack.File)
)

type setting struct {
	token `toml:"token"`
}

type token struct {
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
	api = slack.New(s.token.User)

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
	glog.Info("UserId: ", info.UserID)

	return info.UserID, nil
}

func getFileList(userID string) {
	glog.Info("getFileList()")

	searchParam := slack.NewGetFilesParameters()
	searchParam.User = userID

	files, _, _ := api.GetFiles(searchParam)

	glog.Info("filename list:")
	for _, file := range files {
		fileMap[file.Name] = file
		glog.Info(file.Name)
	}
}

func postRainImg() {
	glog.Info("postRainImg()")

	rtm := slack.New(s.token.Bot).NewRTM()
	go rtm.ManageConnection()

	for {
		event := <-rtm.IncomingEvents
		switch event.Data.(type) {
		case *slack.MessageEvent:
			msg := event.Data.(*slack.MessageEvent)
			glog.Info("channel: ", msg.Channel)
			glog.Info("text: ", msg.Text)

			match, _ := regexp.MatchString("雨", msg.Text)
			if match {
				f := rainImgUpload()
				rtm.SendMessage(rtm.NewOutgoingMessage(f.URL, msg.Channel))
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
