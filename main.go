package main

import (
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"text/template"
)

var (
	publicAddress    = ""
	urlKeyMapping    map[string]string
	mu               sync.Mutex
	homePageTemplate = template.Must(template.ParseFiles("index.html")) // Загружаем шаблон
)

func randKey(url string) (string, error) {
	md5Hash := sha1.New()
	if _, err := md5Hash.Write([]byte(url)); err != nil {
		return "", err
	}

	return hex.EncodeToString(md5Hash.Sum(nil))[:8], nil
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	url := r.FormValue("url")
	if url == "" {
		http.Error(w, "parameter 'url' is required", http.StatusBadRequest)
		return
	}

	key, err := randKey(url)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}

	mu.Lock()
	urlKeyMapping[key] = url
	mu.Unlock()

	shortURL := fmt.Sprintf("%s/%s", publicAddress, key)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	homePageTemplate.Execute(w, shortURL)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	shortKey := r.URL.Path[1:]

	if shortKey == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		homePageTemplate.Execute(w, nil)
		return
	}
	paths := strings.Split(r.URL.Path, "/")
	key := paths[len(paths)-1]

	mu.Lock()
	url, ok := urlKeyMapping[key]
	mu.Unlock()

	if !ok {
		http.Error(w, "short key not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, url, http.StatusMovedPermanently)
}

func main() {
	port := flag.String("port", "8080", "server port")
	publicAddress = *flag.String("host", "http://localhost:8080", "server host")
	flag.Parse()

	host := fmt.Sprintf(":%s", *port)

	urlKeyMapping = make(map[string]string)

	http.HandleFunc("/shorten", shortenHandler)
	http.HandleFunc("/", rootHandler)

	fmt.Println("Сервер запущен на:", host)
	fmt.Println("Переадресация ссылок на:", publicAddress)

	log.Fatal(http.ListenAndServe(host, nil))
}
