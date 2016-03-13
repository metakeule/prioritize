package main

import (
	"encoding/json"
	"fmt"
	"github.com/metakeule/config"
	"github.com/srtkkou/zgok"
	"lib"
	"lib/webserver"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	args     = config.MustNew("prioritize", "0.1", "a webbased tool to help you prioritize based on dependencies")
	argPort  = args.NewInt32("port", "port on which the webserver runs", config.Default(int32(8080)), config.Shortflag('p'))
	argHost  = args.NewString("host", "hostname or ip address of the webserver", config.Default("localhost"), config.Shortflag('h'))
	argFile  = args.NewString("file", "file that acts as data store (json)", config.Default("prioritize.json"), config.Shortflag('f'))
	argDebug = args.NewBool("debug", "turn on debugging", config.Default(false))
)

type setup struct {
	Host         string
	Port         int
	Wd           string
	SelfBinName  string
	App          string
	CreatingFile bool
	zfs          zgok.FileSystem
	file         *os.File
	store        *lib.JSONStore
}

func main() {

	var (
		err error
		set setup
	)

steps:
	for jump := 1; err == nil; jump++ {
		switch jump - 1 {
		default:
			break steps
		case 0:
			err = args.Run()
		case 1:
			set.Port = int(argPort.Get())
			set.Host = argHost.Get()
			set.Wd, err = os.Getwd()
		case 2:
			set.SelfBinName = os.Args[0]
			_, err = os.Stat(set.SelfBinName)
			if err != nil && os.IsNotExist(err) {
				set.SelfBinName, err = which(set.SelfBinName)
			}
		case 3:
			set.zfs, err = zgok.RestoreFileSystem(set.SelfBinName)
		case 4:
			fpath := filepath.Join(set.Wd, argFile.Get())
			set.file, err = os.OpenFile(fpath, os.O_RDWR, 0644)
			if err != nil && os.IsNotExist(err) {
				set.CreatingFile = true
				set.file, err = os.Create(fpath)
			}
		case 5:
			defer set.file.Close()
			set.store = lib.NewJSONStore()
			set.store.Reader = set.file
			set.store.Writer = set.file
			if !set.CreatingFile {
				err = set.store.Load()
			} else {
				err = set.store.Save()
			}
			defer set.store.Save()
			set.App = getAppname(set.Wd)
		}
	}

	if argDebug.Get() {
		if b, e := json.MarshalIndent(set, "", "  "); e == nil {
			os.Stdout.Write(b)
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	set.serve()
	os.Exit(0)
}

func getAppname(wd string) string {
	da := strings.Split(filepath.ToSlash(wd), "/")
	l := len(da)
	if l == 0 {
		return ""
	}
	return da[l-1]
}

func which(name string) (string, error) {
	out, err := exec.Command("which", name).CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (set *setup) serve() {
	server := webserver.NewStoreServer(set.App, set.store)

	// assetServer := zfs.FileServer("static")
	http.Handle("/static/", http.StripPrefix("/static/", set.zfs.FileServer("static")))

	// http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir(staticRoot()))))
	http.HandleFunc("/app/name", server.AppName)
	// http.HandleFunc("/item/tree", server.ItemTree)
	// http.HandleFunc("/item/graphviz", server.ItemsGraphviz)
	// http.HandleFunc("/tag/tree", server.TagTree)
	// http.HandleFunc("/item/all", server.AllItems)
	http.HandleFunc("/item/vis", server.ItemsVisDataSet)
	http.HandleFunc("/item/rename", server.RenameItem)
	http.HandleFunc("/item/remove", server.RemoveItem)
	http.HandleFunc("/item/remove-edge", server.RemoveItemEdge)
	// http.HandleFunc("/tag/all", server.AllTags)
	http.HandleFunc("/item/put", server.PutItem)
	http.HandleFunc("/item/put-edge", server.PutItemEdge)
	// http.HandleFunc("/tag/put", server.PutTag)
	http.HandleFunc("/", serveIndex)

	hoststr := fmt.Sprintf("%s:%d", set.Host, set.Port)
	fmt.Fprintf(os.Stdout, "\nlistening on http://%s\n", hoststr)
	http.ListenAndServe(hoststr, nil)
}

func serveIndex(w http.ResponseWriter, req *http.Request) {
	w.Write(indexPage)
}

var indexPage = []byte(`<!doctype html>
<html>
<head>
  <title>Prioritize</title>

  <script src="/static/jquery-2.2.1.min.js"></script>
  <script type="text/javascript" src="/static/vis-v4.15/vis.min.js"></script>
  <link href="/static/vis-v4.15/vis.min.css" rel="stylesheet" type="text/css" />

  <style type="text/css">
    html {
      width: 100%;
      height: 100%;
    }
    body {
      margin: 0;
      width: 100%;
      height: 100%;
    }

    #mynetwork {
      width: 100%;
      height: 100%;
      background-color: gray;
    }
  </style>
</head>
<body id="canvassizer">
  <div id="mynetwork"></div>
  <script type="text/javascript" src="/static/prioritize.js"></script>
</body>
</html>
`)
