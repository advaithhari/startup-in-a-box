package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	webview "github.com/webview/webview_go"
)

func main() {
	// Start websocket server in background goroutine
	go func() {
		http.HandleFunc("/ws", wsHandler)
		log.Println("WebSocket server running on :8080")
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal("ListenAndServe:", err)
		}
	}()

	// Load the UI in webview
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	htmlPath := filepath.Join(wd, "index.html")

	w := webview.New(true) // debug mode
	defer w.Destroy()
	w.SetTitle("Go Webview UI")
	w.SetSize(800, 600, webview.HintNone)
	w.Navigate("file://" + htmlPath)
	w.Run()
}
