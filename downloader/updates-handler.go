package main

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type UpdatesHandler struct {
	maxConnections int
	updatesChan    chan DownloadItemUpdate
	conns          map[*websocket.Conn]struct{}
}

func (uh *UpdatesHandler) AddConn(conn *websocket.Conn) error {
	if len(uh.conns) > uh.maxConnections {
		return fmt.Errorf("Max connections exceeded.")
	}
	uh.conns[conn] = struct{}{}
	return nil
}

func (uh *UpdatesHandler) Handle() {
	go func() {
		for update := range uh.updatesChan {
			updateJSON, err := update.JSON()
			if err != nil {
				log.Println(err)
			}

			deleteBuf := make([]*websocket.Conn, 0)
			for conn := range uh.conns {
				err := conn.WriteMessage(websocket.TextMessage, updateJSON)
				if err != nil {
					deleteBuf = append(deleteBuf, conn)
				}
			}
			for _, conn := range deleteBuf {
				conn.Close()
				delete(uh.conns, conn)
			}
		}
		close(uh.updatesChan)
	}()
}
