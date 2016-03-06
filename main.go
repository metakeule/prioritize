package main

import (
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
	cfg     = config.MustNew("prioritize", "0.1", "a webbased tool to help you prioritize based on dependencies")
	argPort = cfg.NewInt32("port", "port on which the webserver runs", config.Default(int32(8080)), config.Shortflag('p'))
	argHost = cfg.NewString("host", "hostname or ip address of the webserver", config.Default("localhost"), config.Shortflag('h'))
	argFile = cfg.NewString("file", "file that acts as data store (json)", config.Default("prioritize.json"), config.Shortflag('f'))
)

func main() {

	var (
		err         error
		selfbinname string
		zfs         zgok.FileSystem
		file        *os.File
		isNew       bool
		store       *lib.JSONStore
		port        int
		host        string
		wd          string
	)

steps:
	for jump := 1; err == nil; jump++ {
		switch jump - 1 {
		default:
			break steps
		case 0:
			err = cfg.Run()
		case 1:
			port = int(argPort.Get())
			host = argHost.Get()
			wd, err = os.Getwd()
		case 2:
			selfbinname = os.Args[0]
			_, err = os.Stat(selfbinname)
			if err != nil && os.IsNotExist(err) {
				selfbinname, err = which(selfbinname)
			}
		case 3:
			zfs, err = zgok.RestoreFileSystem(selfbinname)
		case 4:
			fpath := filepath.Join(wd, argFile.Get())
			file, err = os.OpenFile(fpath, os.O_RDWR, 0644)
			if err != nil && os.IsNotExist(err) {
				isNew = true
				file, err = os.Create(fpath)
			}
		case 5:
			defer file.Close()
			store = lib.NewJSONStore()
			store.Reader = file
			store.Writer = file
			if !isNew {
				err = store.Load()
			} else {
				err = store.Save()
			}
			defer store.Save()
		case 6:
			serve(getAppname(wd), store, zfs, port, host)
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
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

func serve(appName string, store *lib.JSONStore, zfs zgok.FileSystem, port int, host string) {
	server := webserver.NewStoreServer(appName, store)

	// assetServer := zfs.FileServer("static")
	http.Handle("/static/", http.StripPrefix("/static/", zfs.FileServer("static")))

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

	hoststr := fmt.Sprintf("%s:%d", host, port)
	fmt.Fprintf(os.Stdout, "listening on http://%s\n", hoststr)
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
