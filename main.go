package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/kozhevnikova/channellogger"
	"github.com/kozhevnikova/go-get-youtube/youtube"
)

var Error *log.Logger
var channel *channellogger.ChannelData

const (
	pathToFiles = "../ustad/files/"
)

func init() {
	Error = log.New(
		os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile|log.LstdFlags)
}

const (
	applicationName = "ustad_bot::"
	welcomeMessage  = "Welcome to Ustad.\n" +
		"It can extract audio from almost every Youtube video.\n" +
		"Send a link in input field."
	instructionsMessage = "For sending feedback use command /feedback" +
		"and your message"
	receivedRequestMessage = "Your request has been received.\n" +
		"It takes several minutes to process. Please, wait."
	incorrectLinkMessage = "The link is incorrect or video is unavailable.\n" +
		"Please, check whether the link is correct and try later."
	unavailableVideoMessage = "This video is unavailable.\n" +
		"Check whether the link is correct and try later.\n" +
		"If error will be the same send feedback using command " +
		"/feedback and your message"
	sendingAudioMessage   = "Sending the audio."
	feedbackMessage       = "Thanks for your feedback."
	incorrectInputMessage = "Could not understand your instructions." +
		" Use /information command"
	downloadingVideoMessage = "Downloading video from youtube."
)

func main() {
	channel = &channellogger.ChannelData{
		BotID:     "",
		ChannelID: "",
	}

	bot, err := botInitialization()
	if err != nil {
		Error.Println(err)
		channellogger.SendLogInfoToChannel(
			channel, applicationName+err.Error())
		return
	}

	fmt.Fprintln(os.Stdout, "Authorized on account", bot.Self.UserName)

	channellogger.SendLogInfoToChannel(
		channel, applicationName+"started")

	conf := tgbotapi.NewUpdate(0)
	conf.Timeout = 60

	updates, err := bot.GetUpdatesChan(conf)
	if err != nil {
		Error.Println(err)
		channellogger.SendLogInfoToChannel(
			channel, applicationName+err.Error())
		return
	}

	for update := range updates {
		err, errInformation := gettingMessageFromBot(bot, update)
		if err != nil {
			channellogger.SendLogInfoToChannel(
				channel, applicationName+err.Error()+errInformation)
			continue
		}
	}
}

func botInitialization() (*tgbotapi.BotAPI, error) {
	var bot *tgbotapi.BotAPI

	err := os.Setenv("", "")
	if err != nil {
		Error.Println(err)
		return bot, err
	}

	var token string
	token = os.Getenv("token")

	if token == "" {
		err := errors.New("no token")
		return bot, err
	}

	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		Error.Println(err)
		return bot, err
	}

	return bot, nil
}

func sendMessageToAUser(
	bot *tgbotapi.BotAPI,
	chatID int64,
	message string,
) {
	_, err := bot.Send(
		tgbotapi.NewMessage(chatID, fmt.Sprintf(message)))
	if err != nil {
		Error.Println(err)
		channellogger.SendLogInfoToChannel(
			channel, applicationName+err.Error())
		return
	}
}

func sendGreeting(bot *tgbotapi.BotAPI, chatID int64) error {
	_, err := bot.Send(
		tgbotapi.NewMessage(chatID, welcomeMessage))
	if err != nil {
		Error.Println(err)
		return err
	}

	return nil
}

func sendInformationAboutBot(bot *tgbotapi.BotAPI, chatID int64) error {
	sendMessageToAUser(
		bot,
		chatID,
		instructionsMessage,
	)

	return nil
}

func processingVideoID(
	bot *tgbotapi.BotAPI,
	chatID int64,
	update tgbotapi.Update,
) (string, error, string) {

	errInformation := " " +
		update.Message.Text

	videoID, err := getVideoID(update, update.Message.Text)
	if err != nil || videoID == "" {
		Error.Println(err)
		Error.Println(videoID)
		sendMessageToAUser(
			bot,
			chatID,
			incorrectLinkMessage,
		)
		return videoID, err, errInformation
	}

	sendMessageToAUser(
		bot,
		chatID,
		receivedRequestMessage,
	)

	return videoID, nil, errInformation
}

func processingVideoStream(
	videoID string,
	bot *tgbotapi.BotAPI,
	chatID int64,
	update tgbotapi.Update,
) (error, string) {

	errInformation := " " +
		update.Message.Text

	video, err := getVideoStream(videoID)
	if err != nil {
		Error.Println(err)
		sendMessageToAUser(
			bot,
			chatID,
			unavailableVideoMessage,
		)
		return err, errInformation

	} else {
		sendMessageToAUser(bot, chatID, downloadingVideoMessage)

		err = downloadVideo(video, videoID)
		if err != nil {
			Error.Println(err)
			sendMessageToAUser(
				bot,
				chatID,
				unavailableVideoMessage,
			)
			return err, errInformation

		} else {
			err, errInformation := processingAudioStream(
				videoID, bot, chatID, update)
			if err != nil {
				Error.Println(err)
				return err, errInformation
			}
		}
	}

	return nil, ""
}

func processingAudioStream(
	videoID string,
	bot *tgbotapi.BotAPI,
	chatID int64,
	update tgbotapi.Update,
) (error, string) {

	errInformation := " " +
		update.Message.Text

	reply, err := getAudio(videoID)
	if err != nil {
		Error.Println(err)
		sendMessageToAUser(
			bot,
			chatID,
			unavailableVideoMessage,
		)
		return err, errInformation

	} else {
		sendMessageToAUser(
			bot,
			chatID,
			sendingAudioMessage,
		)

		_, err = bot.Send(tgbotapi.NewAudioUpload(chatID, reply))
		if err != nil {
			Error.Println(err)
			return err, errInformation
		}
	}

	err = deleteFiles(pathToFiles)
	if err != nil {
		Error.Println(err)
		return err, "could not delete files"
	}

	return nil, ""
}

func cutCommandWord(message string, command string) string {
	return strings.Replace(message, command, "", -1)
}

func gettingFeedback(
	bot *tgbotapi.BotAPI,
	chatID int64,
	message string,
) {
	message = cutCommandWord(message, "/feedback ")
	channellogger.SendLogInfoToChannel(
		channel, "ustad_bot::feedback::"+message,
	)

	sendMessageToAUser(bot, chatID, feedbackMessage)
}

func gettingMessageFromBot(
	bot *tgbotapi.BotAPI,
	update tgbotapi.Update,
) (error, string) {

	chatID := update.Message.Chat.ID
	command := update.Message.Command()

	switch command {
	case "start":
		err := sendGreeting(bot, chatID)
		if err != nil {
			return err, ""
		}

	case "information":
		err := sendInformationAboutBot(bot, chatID)
		if err != nil {
			return err, ""
		}

	case "":
		videoID, err, errInformation := processingVideoID(bot, chatID, update)
		if err != nil {
			return err, errInformation
		}

		if videoID == "" {
			break
		}

		err, errInformation = processingVideoStream(
			videoID, bot, chatID, update)
		if err != nil {
			return err, errInformation
		}

	case "feedback":
		gettingFeedback(bot, chatID, update.Message.Text)

	default:
		sendMessageToAUser(
			bot,
			chatID,
			incorrectInputMessage,
		)
	}

	return nil, ""
}

func deleteFiles(pathToFiles string) error {
	dir, err := os.Open(pathToFiles)
	if err != nil {
		return err
	}

	dirFile, err := dir.Readdir(0)
	if err != nil {
		return err
	}

	for i := range dirFile {
		file := dirFile[i]
		name := file.Name()
		pathToFile := pathToFiles + name
		os.Remove(pathToFile)
	}

	channellogger.SendLogInfoToChannel(
		channel, "Files have been removed")

	return nil
}

func downloadVideo(video youtube.Video, videoID string) error {
	option := &youtube.Option{
		Resume: false,
		Mp3:    true,
	}

	if ok := checkDirIfNotExist(pathToFiles); !ok {
		err := createDir(pathToFiles)
		if err != nil {
			return err
		}
	}

	err := video.Download(0, pathToFiles+videoID+".mp4", option)
	if err != nil {
		return err
	}

	return nil
}

func checkDirIfNotExist(dir string) bool {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return false
	}

	return true
}

func createDir(dir string) error {
	err := os.Mkdir(dir, 0777)
	if err != nil {
		return err
	}

	return nil
}

func getAudio(videoID string) (tgbotapi.FileBytes, error) {
	file := tgbotapi.FileBytes{}

	audio, err := ioutil.ReadFile(pathToFiles + videoID + ".mp3")
	if err != nil {
		return file, err
	}

	file = tgbotapi.FileBytes{
		Name:  videoID + ".mp3",
		Bytes: audio,
	}

	return file, nil
}

func getVideoStream(videoID string) (youtube.Video, error) {
	videoStream, err := youtube.Get(videoID)
	if err != nil {
		return videoStream, err
	}

	return videoStream, nil
}

func getVideoID(update tgbotapi.Update, url string) (string, error) {
	videoID := ""

	if ok := ifWeb(url); ok {
		videoID = strings.TrimPrefix(
			update.Message.Text, "https://www.youtube.com/watch?v=")
		return videoID, nil
	}

	if ok := ifMobile(url); ok {
		videoID = strings.TrimPrefix(
			update.Message.Text, "https://m.youtube.com/watch?v=")
		return videoID, nil
	}

	return videoID, nil
}

func ifWeb(url string) bool {
	web, err := regexp.MatchString("www.youtube", url)
	if err != nil {
		return false
	}

	return web
}

func ifMobile(url string) bool {
	mobile, err := regexp.MatchString("m.youtube", url)
	if err != nil {
		return mobile
	}

	return mobile
}
