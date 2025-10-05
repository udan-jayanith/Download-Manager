package main

import (
	"encoding/json"
	"os"

	"net/http"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

var (
	_                = godotenv.Load("./.env")
	downloadWorkPool = NewDownloadWorkPool()
	updatesHandler   = UpdatesHandler{
		maxConnections: 4,
		updatesChan:    downloadWorkPool.Updates,
		conns:          make(map[*websocket.Conn]struct{}, 1),
	}
)

func main() {
	updatesHandler.Handle()

	mux := http.NewServeMux()
	HandleAuth(mux)
	HandleDownloads(mux)

	//Serve pages in pages dir
	fs := http.FileServer(http.Dir("./pages"))
	mux.Handle("/pages/", http.StripPrefix("/pages/", fs))

	http.ListenAndServe(os.Getenv("port"), mux)
}

func AllowCrossOrigin(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func WriteError(w http.ResponseWriter, errorMsg string) {
	w.Header().Add("Content-Type", "application/json")
	res := map[string]string{
		"error": errorMsg,
	}
	json.NewEncoder(w).Encode(&res)
}
