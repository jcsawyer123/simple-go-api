package profiling

import (
	"log"
	"net/http"
	_ "net/http/pprof"
)

func Start(port string) {
	go func() {
		log.Printf("Starting pprof server on port %s", port)
		log.Println(http.ListenAndServe("localhost:"+port, nil))
	}()
}
