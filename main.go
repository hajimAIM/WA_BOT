package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

var (
	WAClient *whatsmeow.Client

	// List of recipient numbers
	recipientNumbers = []string{
		"601160564476@s.whatsapp.net",

		"60122412027@s.whatsapp.net",

		"60132270058@s.whatsapp.net",
	}
)

func eventHandler(evt interface{}) {
	switch event := evt.(type) {
	case *events.Message:
		if msg := event.Message.GetConversation(); !event.Info.IsFromMe {
			var isInRecipient bool
			for _, n := range recipientNumbers {
				if event.Info.Sender.String() == n {
					isInRecipient = true
					break
				}
			}

			log := waLog.Stdout(fmt.Sprintf("[UpdateHandler][%s][IsRecipient: %t]", event.Info.Sender, isInRecipient), "DEBUG", false)
			log.Infof(`Received new message -> "%s"`, msg)

			replyStr := "Pesan ini automatik, menggunakan GO!. Anda mengirim pesan: " + msg
			replyMsg := &waProto.Message{
				ListMessage: &waProto.ListMessage{
					Title:       proto.String("ListMessage title"),
					Description: proto.String("ListMessage Description"),
					FooterText:  proto.String("ListMessage footer"),
					ButtonText:  proto.String("ListMessage ButtonText"),
					ListType:    waProto.ListMessage_SINGLE_SELECT.Enum(),
					Sections: []*waProto.Section{
						{
							Title: proto.String("Section1 title"),
							Rows: []*waProto.Row{
								{
									RowId:       proto.String("id1"),
									Title:       proto.String("ListMessage section row title"),
									Description: proto.String("ListMessage section row desc"),
								},
								{
									RowId:       proto.String("id2"),
									Title:       proto.String("title 2"),
									Description: proto.String("desc 2"),
								},
							},
						},
						{
							Title: proto.String("Section2 title"),
							Rows: []*waProto.Row{
								{
									RowId:       proto.String("id1"),
									Title:       proto.String("ListMessage section row title"),
									Description: proto.String("ListMessage section row desc"),
								},
								{
									RowId:       proto.String("id2"),
									Title:       proto.String("title 2"),
									Description: proto.String("desc 2"),
								},
							},
						},
					},
				},
			}

			switch msg {
			case "/admin", "/Admin", "Admin", "admin", "ADMIN":
				if isInRecipient == true {
					replyStr = "YA TUAN APA SAYA BOLEH BANTU"
				} else if isInRecipient != true {
					replyStr = "Pesan ini automatik, menggunakan GO!. Anda mengirim pesan: " + msg
				}
			case "Menu", "menu", "/Menu", "/menu", "MENU":
				replyStr = "Sila pilih menu di atas"
				WAClient.SendMessage(event.Info.Sender, "", replyMsg)

			}

			WAClient.SendMessage(event.Info.Sender, "", &waProto.Message{
				Conversation: proto.String(replyStr),
			})

		}
	}
}

func main() {
	dbLog := waLog.Stdout("Database", "DEBUG", false)

	// Make sure you add appropriate DB connector imports, e.g. github.com/mattn/go-sqlite3 for SQLite
	container, err := sqlstore.New("sqlite3", "file:wsap.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}

	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
	}

	WAClient = whatsmeow.NewClient(deviceStore, waLog.Stdout("Client", "DEBUG", false))
	WAClient.AddEventHandler(eventHandler)

	// Handle client auth
	if WAClient.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := WAClient.GetQRChannel(context.Background())
		err = WAClient.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Render the QR code here
				// e.g. qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				// or just manually `echo 2@... | qrencode -t ansiutf8` in a terminal
				fmt.Println("QR code:", evt.Code)
				//qrterminal.Generate(evt.Code, qrterminal.L, os.Stdout)
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		if err = WAClient.Connect(); err != nil {
			panic(err)
		}
	}

	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	WAClient.Disconnect()
}
