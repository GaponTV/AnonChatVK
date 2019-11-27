package main

import (
	"fmt"
	"log"
	"time"
	"processing"
)

func msgProcessor(messages <-chan string) {
	for m := range messages {
		fmt.Println("msgProcessor: " + m)
		time.Sleep(3 * time.Second)
	}
	fmt.Println("msgProcessor() END!")
}

func main() {
	group := processing.NewVKChatGroup("57d2bea39b21468fc86a98b779fe7e4e9c6f19a9850cc05276f621b38130a172f0d4eca4b5be8cd09b8b3", 187139970) // 1 8
	err := group.SendMessageTo(173960451)
	if err != nil {
		log.Println("An error occured: ", err.Error())
	}

	go func() {
		err := group.Start()
		if err != nil {
			fmt.Println("on long polling listener error occured:", err.Error())
		}
	}()

	fmt.Scanln()
}
