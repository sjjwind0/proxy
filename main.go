package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var imageIndex int = 0

func ReadFileContent(path string) ([]byte, error) {
	fileInfo, err := os.Stat(path)
	if err == nil {
		content := make([]byte, fileInfo.Size())
		file, _ := os.Open(path)
		defer file.Close()
		file.Read(content)
		return content, nil
	}
	return nil, err
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	baseFilePath := "/home/wind/Downloads/we-chat/"
	fileName := r.Form["name"][0]
	realFilePath := baseFilePath + fileName
	content, err := ReadFileContent(realFilePath)
	if err == nil {
		w.Header().Set("Accept", "*/*")
		if strings.Index(fileName, "webp") != -1 {
			w.Header().Set("Content-Type", "image/webp")
		} else {
			w.Header().Set("Content-Type", "image/jpeg")
		}
		w.Write(content)
	} else {
		fmt.Fprintf(w, "Error")
	}
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	var imageList []string
	fileName := r.RemoteAddr
	fileName = strings.Split(fileName, ":")[0]
	filepath.Walk("/home/wind/Downloads/we-chat", func(path string, file os.FileInfo, err error) error {
		if file == nil {
			return err
		}
		if !file.IsDir() {
			baseName := filepath.Base(path)
			if strings.Index(baseName, fileName) != -1 {
				imageList = append(imageList, baseName)
			}
		}
		return nil
	})
	t, err := template.ParseFiles("test.html")
	if err == nil {
		t.Execute(w, imageList)
	} else {
		fmt.Fprintf(w, "Error")
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	r.RequestURI = ""
	client := &http.Client{}
	resp, err := client.Do(r)
	defer resp.Body.Close()
	if err != nil {
		fmt.Println(err)
	}
	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
	for _, c := range resp.Cookies() {
		w.Header().Add("Set-Cookie", c.Raw)
	}
	result, err := ioutil.ReadAll(resp.Body)
	if -1 != strings.Index(r.URL.Path, "/mmsns/") {
		fmt.Println("方式: ", r.Method, "\t请求: ", r.URL)
		fileName := r.RemoteAddr
		fileName = strings.Split(fileName, ":")[0]
		cachePath := "/home/wind/Downloads/we-chat/" + fileName
		cachePath = fmt.Sprintf("%s_%d", cachePath, imageIndex)
		imageIndex++
		contentType := resp.Header.Get("Content-Type")
		if resp.Header.Get("Content-Type") == "image/jpeg" || resp.Header.Get("Content-Type") == "image/webp" {
			if contentType == "image/jpeg" {
				cachePath += ".jpg"
			} else {
				cachePath += ".webp"
			}
			f, err := os.Create(cachePath)
			if err == nil {
				defer f.Close()
				f.Write(result)
			}
		}
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(result)
}

func main() {
	go func() {
		http.HandleFunc("/image", imageHandler)
		http.HandleFunc("/download", downloadHandler)
		http.ListenAndServe(":7777", nil)
	}()
	http.HandleFunc("/", handler)
	log.Println("Start serving on port 8888")
	http.ListenAndServe(":8888", nil)
	os.Exit(0)
}
