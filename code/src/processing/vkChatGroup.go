package processing

import (
	"encoding/json"
	"log"
	"time"
	"vkapi"
)

type VKChatGroup struct {
	groupID           int64
	api               *vkapi.VKBotAPI
	chatProc          ChatProc
	outcomingMessages chan<- vkapi.SendMessageRequest
}

func NewVKChatGroup(accessToken string, groupID int64) *VKChatGroup {
	api := vkapi.NewVKBotAPI(accessToken, groupID)
	grp := &VKChatGroup{
		groupID: groupID,
		api:     api,
	}
	outcomingMessages := make(chan vkapi.SendMessageRequest, 30)
	grp.outcomingMessages = outcomingMessages
	grp.chatProc.outcomingMsg = outcomingMessages
	distributionTicker := time.Tick(time.Second * 5)
	go grp.chatProc.DistributeByChatsWorker(distributionTicker)
	go grp.processOutcomingMessages(outcomingMessages)
	return grp
}

func (grp *VKChatGroup) processOutcomingMessages(outcomingMessages <-chan vkapi.SendMessageRequest) {
	for msg := range outcomingMessages {
		msg.Run(grp.api)
	}

}

func (grp *VKChatGroup) SendMessageTo(userID int64) error {
	apiReq := vkapi.SendMessageRequest{
		UserID:  userID,
		Content: "Hello from struct!",
	}
	response, err := apiReq.Run(grp.api)
	log.Println("response in group:  ", response)
	return err
}

func (grp *VKChatGroup) processNewMessages(messages <-chan vkapi.MessageObject) {
	for msg := range messages {
		senderID := msg.SenderID
		if len(msg.Payload) > 0 {
			var payload ButtonPayload
			if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
				log.Println("Произошла ошибка при попытке преобразовать JSON payload'а сообщения:", err)
			}
			grp.processCommand(&msg, payload.CommandName)
		} else {
			err := grp.chatProc.ProcessSimpleMessage(&msg)
			if err != nil {
				if err == ErrUserOutOfChat {
					log.Printf("Пользователь с id:%d не находится ни в одном чате... Отправляем справку...\n", senderID)
					var notifyMsg string
					var keyboard *vkapi.Keyboard
					if grp.chatProc.checkUserInQueue(senderID) {
						notifyMsg = "Вы находитесь в очереди ожидания! Подождите немного времени и мы найдем вам собеседника :-)"
					} else {
						notifyMsg = "Вы не находитесь в режиме общения. Нажмите на кнопку 'Искать собеседника', для того чтобы найти себе партнера для общения!"
						keyboard = outOfChatKeyboard()
					}
					grp.outcomingMessages <- vkapi.SendMessageRequest{
						UserID:   senderID,
						Content:  notifyMsg,
						Keyboard: keyboard,
					}
				} else {
					log.Println("Произошла ошибка при обработке простого сообщения от пользователя с id:", senderID, ". Причина:", err)
				}
			}
		}
	}
}

func (grp *VKChatGroup) Start() error {
	messagesChan := make(chan vkapi.MessageObject, 100)
	go grp.processNewMessages(messagesChan)
	err := grp.api.OnLongPoolMessage(messagesChan)
	return err
}
