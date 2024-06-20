package main

import (
	"URLShortner/pkg"
	"encoding/json"
	"net/http"
	"strings"
)

func main() {
	router := http.NewServeMux()
	router.HandleFunc("GET /verify/", func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")

		resp := pkg.WebHookResponse{}
		if strings.HasPrefix(ua, "Mozilla") {
			resp.Pass = false
			resp.Reason = "Cant pass through Mozila user-agnet"
		} else {
			resp.Pass = true
		}
		respBytes, err := json.Marshal(&resp)
		if err != nil {
			panic(err)
		}

		w.Write(respBytes)
	})

	http.ListenAndServe("localhost:8001", router)
}
