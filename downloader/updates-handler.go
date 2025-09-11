package main

import (
	"github.com/gorilla/websocket"
)

type UpdatesHandler struct {
	maxConnections int
	updatesChan    chan DownloadItemUpdate
	conns          map[*websocket.Conn]struct{}
}

func (uh *UpdatesHandler) Handle() {
	go func() {
		for update := range uh.updatesChan {
			deleteBuf := make([]*websocket.Conn, 0)
			for conn := range uh.conns {
				err := conn.WriteJSON(&update)
				if err != nil {
					deleteBuf = append(deleteBuf, conn)
				}
			}
			for _, conn := range deleteBuf {
				delete(uh.conns, conn)
			}
		}
	}()
}
