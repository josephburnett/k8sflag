package main

import (
	"fmt"
	"net/http"

	"github.com/josephburnett/k8sflag/pkg/k8sflag"
)

var config = k8sflag.NewConfigMap("/etc/config")
var name = config.String("hello.name", "nobody")

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello %v.\n", name.Get())
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
