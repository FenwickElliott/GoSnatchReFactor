package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
)

var db = "."

func main() {
	accessBearer, err := ioutil.ReadFile(db + "accessBearer")

	if err != nil {
		initialize()
	} else {
		os.Setenv("AccessBearer", string(accessBearer))
	}
}

func exchangeCode(code string) bool {
	fmt.Println(code)
	return true
}

func initialize() {
	fmt.Println("Initializing...")
	done := make(chan bool)
	go serve(done)
	exec.Command("open", "https://accounts.spotify.com/authorize/?client_id=715c15fc7503401fb136d6a79079b50c&response_type=code&redirect_uri=http://localhost:3456/catch&scope=user-read-playback-state%20playlist-read-private%20playlist-modify-private").Start()
	finished := <-done
	if finished {
		fmt.Println("Initialization complete")
	}
}

func serve(done chan bool) {
	http.HandleFunc("/catch", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Thank you, GoSnatch can now access your spotify account.\nYou may close this window.\n")
		code := r.URL.Query()["code"][0]
		done <- exchangeCode(code)
	})
	http.ListenAndServe(":3456", nil)
}
