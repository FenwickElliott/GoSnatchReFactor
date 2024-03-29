package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

var db = "."

func main() {
	accessBearer, err := ioutil.ReadFile("accessBearer")

	if err != nil {
		initialize()
	} else {
		os.Setenv("AccessBearer", string(accessBearer))
	}

	write("id", get("me")["id"].(string))
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

func exchangeCode(code string) bool {
	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader("grant_type=authorization_code&code="+code+"&redirect_uri=http://localhost:3456/catch"))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic NzE1YzE1ZmM3NTAzNDAxZmIxMzZkNmE3OTA3OWI1MGM6ZTkxZWZkZDAzNDVkNDlkNTllOGE2ZDc1YjUzZTE2YTE=")
	resp, err := http.DefaultClient.Do(req)
	check(err)
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	bodyMap := make(map[string]interface{})

	err = json.Unmarshal(body, &bodyMap)
	check(err)

	os.Setenv("AccessBearer", bodyMap["access_token"].(string))
	write("accessBearer", "Bearer "+bodyMap["access_token"].(string))
	write("refreshBody", "grant_type=refresh_token&refresh_token="+bodyMap["refresh_token"].(string))

	return true
}

func get(endpoint string) map[string]interface{} {
	url := "https://api.spotify.com/v1/" + endpoint
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", os.Getenv("AccessBearer"))

	resp, err := http.DefaultClient.Do(req)
	check(err)
	defer resp.Body.Close()

	res, _ := ioutil.ReadAll(resp.Body)
	resMap := make(map[string]interface{})

	err = json.Unmarshal(res, &resMap)
	check(err)

	return resMap
}

func write(name, content string) error {
	target := name
	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString(content)
	return nil
}

func check(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}
