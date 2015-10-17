package main

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/178inaba/rainimg"
	"github.com/BurntSushi/toml"
	"github.com/nlopes/slack"

	log "github.com/Sirupsen/logrus"
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
	log.Info("init()")

	flag.Parse()
}

func main() {
	log.Info("main()")

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
	log.Info("loadSetting()")

	_, err := toml.DecodeFile(settingToml, &s)
	if err != nil {
		log.Error("load error: ", err)
	}
}

func getUserID() (string, error) {
	log.Info("getUserID()")

	info, err := api.AuthTest()
	if err != nil {
		log.Error("AuthTest Error: ", err)
		return "", err
	}

	log.Info("User: ", info.User)
	log.Info("UserId: ", info.UserID)

	return info.UserID, nil
}

func getFileList(userID string) {
	log.Info("getFileList()")

	searchParam := slack.NewGetFilesParameters()
	searchParam.User = userID

	files, _, _ := api.GetFiles(searchParam)

	log.Info("filename list:")
	for _, file := range files {
		fileMap[file.Name] = file
		log.Info(file.Name)
	}
}

func postRainImg() {
	log.Info("postRainImg()")

	rtm := slack.New(s.token.Bot).NewRTM()
	go rtm.ManageConnection()

	for {
		event := <-rtm.IncomingEvents
		switch event.Data.(type) {
		case *slack.MessageEvent:
			msg := event.Data.(*slack.MessageEvent)
			log.Info("channel: ", msg.Channel)
			log.Info("text: ", msg.Text)

			match, _ := regexp.MatchString("雨", msg.Text)
			if match {
				f := rainImgUpload()
				rtm.SendMessage(rtm.NewOutgoingMessage(f.URL, msg.Channel))
			}
		case slack.LatencyReport:
			latency := event.Data.(slack.LatencyReport)
			log.Info("ping latency: ", latency.Value)
		}
	}
}

func rainImgUpload() slack.File {
	log.Info("rainImgUpload()")

	// create image
	fPath := rainimg.GetImgPath()

	// get filename
	fileName := filepath.Base(fPath)

	// already uploaded check
	file, ok := fileMap[fileName]
	if ok {
		log.Info("already uploaded: ", file.Name)
		return file
	}

	// file up param
	var fup slack.FileUploadParameters
	fup.File = fPath

	// upload
	upFile, _ := api.UploadFile(fup)
	log.Info("upload file: ", upFile.Name)

	// add file list
	fileMap[upFile.Name] = *upFile

	return *upFile
}
