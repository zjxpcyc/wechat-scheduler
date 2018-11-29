package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/zjxpcyc/wechat-scheduler/database"
	"github.com/zjxpcyc/wechat-scheduler/jobs"
	"github.com/zjxpcyc/wechat-scheduler/lib"
)

// Version 当前版本
const Version = "v0.1.0"

var version = flag.Bool("v", false, "Show version of the system")
var port = flag.Int("p", 9001, "Define http port, default is 9001")
var logger = lib.GetLogger()

func newHandler() http.Handler {
	database.Init()
	jobs.Init()

	app := new(App)
	mux := http.NewServeMux()
	mux.Handle("/", app)
	return mux
}

func main() {
	flag.Parse()

	if *version {
		fmt.Println(Version)
		os.Exit(0)
	}

	serv := http.Server{Addr: ":" + strconv.Itoa(*port), Handler: newHandler()}
	log.Fatalln(serv.ListenAndServe())
}
