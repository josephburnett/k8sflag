package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/josephburnett/k8sflag/pkg/k8sflag"
)

var config = k8sflag.NewConfigMap("/etc/config")
var name = config.String("hello.name", "nobody")
var birthday = time.Now()

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello %v. My birthday is %v.\n", name.Get(), birthday)
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
