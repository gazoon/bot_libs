package gateway

import (
	"time"

	"context"

	"strings"

	"github.com/gazoon/bot_libs/logging"
	"github.com/gazoon/bot_libs/queue/messages"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
)

var (
	gLogger = logging.WithPackage("telegram_polling")
)

const (
	audioEncoding   = "OGG_OPUS"
	audioSampleRate = 16000
)

func transformUser(u *tgbotapi.User) *msgsqueue.User {
	if u == nil {
		return nil
	}
	name := u.FirstName
	if u.LastName != "" {
		name += " " + u.LastName
	}
	return &msgsqueue.User{
		ID:       u.ID,
		Name:     name,
		Username: u.UserName,
	}
}

func transformChat(c *tgbotapi.Chat) *msgsqueue.Chat {
	return &msgsqueue.Chat{
		ID:        int(c.ID),
		IsPrivate: c.IsPrivate(),
		Title:     c.Title,
	}
}

func prepareContext(requestID string) context.Context {
	logger := logging.WithRequestID(requestID)
	ctx := logging.NewContext(context.Background(), logger)
	return ctx
}

type TelegramPoller struct {
	queue       msgsqueue.Producer
	botName     string
	pollTimeout int
	retryDelay  int
	apiToken    string
}

func NewTelegramPoller(queue msgsqueue.Producer, apiToken, botName string, pollTimeout, retryDelay int) *TelegramPoller {
	return &TelegramPoller{queue: queue, botName: botName, apiToken: apiToken, pollTimeout: pollTimeout, retryDelay: retryDelay}
}

func (tp *TelegramPoller) userIsBot(u *tgbotapi.User) bool {
	if u == nil {
		return false
	}
	return u.UserName == tp.botName
}

func (tp *TelegramPoller) updateMessageToModel(updateMessage *tgbotapi.Message) (*msgsqueue.Message, error) {
	if updateMessage.Chat == nil {
		return nil, errors.New("message without chat")
	}
	if updateMessage.From == nil {
		return nil, errors.New("message without from")
	}
	var voice *msgsqueue.Voice
	if updateMessage.Voice != nil {
		var voiceSize *int
		size := updateMessage.Voice.FileSize
		if size != 0 {
			voiceSize = &size
		}
		voice = &msgsqueue.Voice{
			ID:         updateMessage.Voice.FileID,
			Duration:   updateMessage.Voice.Duration,
			Size:       voiceSize,
			Encoding:   audioEncoding,
			SampleRate: audioSampleRate,
		}
	}
	var newChatMember *tgbotapi.User
	if updateMessage.NewChatMembers != nil && len(*updateMessage.NewChatMembers) != 0 {
		newChatMember = &(*updateMessage.NewChatMembers)[0]
	}
	message := &msgsqueue.Message{
		MessageID:         updateMessage.MessageID,
		Text:       updateMessage.Text,
		Voice:      voice,
		Chat:       transformChat(updateMessage.Chat),
		From:       transformUser(updateMessage.From),
		IsBotAdded: tp.userIsBot(newChatMember),
		IsBotLeft:  tp.userIsBot(updateMessage.LeftChatMember),
	}
	if !message.IsBotAdded {
		message.NewChatMember = transformUser(newChatMember)
	}
	if !message.IsBotLeft {
		message.LeftChatMember = transformUser(updateMessage.LeftChatMember)
	}
	return message, nil
}

func callbackQueryToModel(callback *tgbotapi.CallbackQuery) (*msgsqueue.Message, error) {
	if callback.From == nil {
		return nil, errors.New("callback without from")
	}
	if callback.Message == nil {
		return nil, errors.New("callback without message")
	}
	if callback.Message.Chat == nil {
		return nil, errors.New("callback without chat")
	}
	message := &msgsqueue.Message{
		Chat: transformChat(callback.Message.Chat),
		Text: callback.Data,
		From: transformUser(callback.From),
	}
	return message, nil
}

func (tp *TelegramPoller) processUpdate(update *tgbotapi.Update) {
	requestID := logging.NewRequestID()
	ctx := prepareContext(requestID)
	logger := logging.FromContextAndBase(ctx, gLogger)
	if update.Message == nil && update.CallbackQuery == nil {
		logger.Infof("Skip update without the Message or CallbackQuery fields: %+v", update)
		return
	}
	var msg *msgsqueue.Message
	if update.Message != nil {
		var err error
		msg, err = tp.updateMessageToModel(update.Message)
		if err != nil {
			logger.Warnf("Cannot transform telegram update message %+v to queue message: %s", update.Message, err)
			return
		}
	} else {
		var err error
		msg, err = callbackQueryToModel(update.CallbackQuery)
		if err != nil {
			logger.Warnf("Cannot transform telegram callback %+v to queue message: %s", update.CallbackQuery, err)
			return
		}
	}
	msg.RequestID = requestID
	msg.CreatedAt = time.Now()
	msg.Text = strings.TrimSpace(msg.Text)
	logger.WithField("msg", msg).Info("Put a new msg in the incoming queue")
	err := tp.queue.Put(ctx, msg)
	if err != nil {
		logger.Errorf("Cannot put incoming msg: %s", err)
	}
}

func (tp *TelegramPoller) Start() error {
	bot, err := tgbotapi.NewBotAPI(tp.apiToken)
	if err != nil {
		return errors.Wrap(err, "telegram api initialization failed")
	}
	gLogger.Info("Starting polling for updates")
	updatesConf := tgbotapi.UpdateConfig{Timeout: tp.pollTimeout}
	go func() {
		for {
			gLogger.Info("Requesting new updates from API")
			updates, err := bot.GetUpdates(updatesConf)
			if err != nil {
				gLogger.Warnf("Failed to get updates: %s retrying in %d seconds...", err, tp.retryDelay)
				time.Sleep(time.Second * time.Duration(tp.retryDelay))
				continue
			}
			gLogger.Infof("Telegram API returns updates: %+v", updates)
			for i := range updates {
				update := &updates[i]
				if update.UpdateID >= updatesConf.Offset {
					updatesConf.Offset = update.UpdateID + 1
					tp.processUpdate(update)
				}
			}
		}
	}()
	return nil
}
