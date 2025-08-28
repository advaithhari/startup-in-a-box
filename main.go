// webview.go
package main

import (
	"log"
	"os"
	"path/filepath"

	webview "github.com/webview/webview_go"
)

func main() {
	// Find index.html in current directory
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	htmlPath := filepath.Join(wd, "index.html")

	w := webview.New(true) // true = debug mode
	defer w.Destroy()
	w.SetTitle("Go Webview UI")
	w.SetSize(800, 600, webview.HintNone)
	w.Navigate("file://" + htmlPath)
	w.Run()
}
