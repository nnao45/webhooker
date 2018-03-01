package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const (
	ERRLOG       = "./err.log"
	DATE         = "2006-01-02 15:04:05.000"
)

func errLogging(err error) {
	now := time.Now().Format(DATE)
	if err != nil {
		s := "[" + now + "] " + fmt.Sprintln(err)
		addog(s, ERRLOG)
	}
}

func addog(text string, filename string) {
	var writer *bufio.Writer
	text_data := []byte(text)

	write_file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0755)
	errLogging(err)
	writer = bufio.NewWriter(write_file)
	writer.Write(text_data)
	writer.Flush()
	defer write_file.Close()
}

func curl_post(msg string) {

  //chatwork API hook examples
	apiUrl := "https://api.chatwork.com/"
	//resource := "/v2/rooms/XXXX/messages"
	resource := "/v2/rooms/XXXX/messages"

	u, err := url.ParseRequestURI(apiUrl)
	errLogging(err)
	u.Path = resource
	urlStr := fmt.Sprintf("%v", u)

	data := url.Values{}
	data.Set("body", msg)

	client := &http.Client{}
	r, err := http.NewRequest("POST", urlStr, bytes.NewBufferString(data.Encode()))
	errLogging(err)
	r.Header.Add("X-ChatWorkToken", "XXXX")
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(r)
	errLogging(err)
	defer resp.Body.Close()
}

func handler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//To allocate slice for request body
	length, err := strconv.Atoi(r.Header.Get("Content-Length"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//Read body data to parse json
	body := make([]byte, length)
	length, err = r.Body.Read(body)
	if err != nil && err != io.EOF {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var jsonBody map[string]interface{}
	err = json.Unmarshal(body[:length], &jsonBody)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

  //Grafana -> Chatwork webhook SAMPLE
	str := fmt.Sprintln("Grafana Alert") + "[info]" + fmt.Sprint("Title: ") + fmt.Sprintln(jsonBody["title"]) + fmt.Sprint("State: ") + fmt.Sprintln(jsonBody["state"]) + fmt.Sprint("Message: ") + fmt.Sprintln(jsonBody["message"]) + "[/info]" + fmt.Sprintln(`http://localhost:3000/dashboard/db/test`)
	curl_post(str)
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":9000", nil)
}
