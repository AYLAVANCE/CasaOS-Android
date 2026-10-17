//go:generate bash -c "mkdir -p codegen && go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.12.4 -generate types,server,spec -package codegen api/casaos/openapi.yaml > codegen/casaos_api.go"
//go:generate bash -c "mkdir -p codegen/message_bus && go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.12.4 -generate types,client -package message_bus https://raw.githubusercontent.com/IceWhaleTech/CasaOS-MessageBus/main/api/message_bus/openapi.yaml > codegen/message_bus/api.go"
package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/IceWhaleTech/CasaOS-Common/model"
	"github.com/IceWhaleTech/CasaOS-Common/utils/command"
	"github.com/IceWhaleTech/CasaOS-Common/utils/constants"
	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"

	util_http "github.com/IceWhaleTech/CasaOS-Common/utils/http"

	"github.com/IceWhaleTech/CasaOS/common"
	"github.com/IceWhaleTech/CasaOS/pkg/cache"
	"github.com/IceWhaleTech/CasaOS/pkg/config"
	"github.com/IceWhaleTech/CasaOS/pkg/platform"
	"github.com/IceWhaleTech/CasaOS/pkg/sqlite"
	"github.com/IceWhaleTech/CasaOS/pkg/utils/file"
	"github.com/IceWhaleTech/CasaOS/route"
	"github.com/IceWhaleTech/CasaOS/service"
	"go.uber.org/zap"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

const LOCALHOST = "127.0.0.1"

var sqliteDB *gorm.DB

var (
	commit = "private build"
	date   = "private build"

	//go:embed api/index.html
	_docHTML string

	//go:embed api/casaos/openapi.yaml
	_docYAML string

	//go:embed build/sysroot/etc/casaos/casaos.conf.sample
	_confSample string

	configFlag  = flag.String("c", "", "config address")
	dbFlag      = flag.String("db", "", "db path")
	versionFlag = flag.Bool("v", false, "version")
)

func init() {
	flag.Parse()
	if *versionFlag {
		fmt.Println("v" + common.VERSION)
		return
	}

	println("git commit:", commit)
	println("build date:", date)

	platform.Init()

	config.InitSetup(*configFlag, _confSample)

	logger.LogInit(config.AppInfo.LogPath, config.AppInfo.LogSaveName, config.AppInfo.LogFileExt)
	if len(*dbFlag) == 0 {
		*dbFlag = filepath.Join(config.AppInfo.DBPath, "db")
	}

	if platform.IsAndroid() && (strings.HasPrefix(*dbFlag, "/var/") || strings.HasPrefix(*dbFlag, "/usr/")) {
		home := os.Getenv("HOME")
		if home == "" {
			home = "/data/data/com.termux/files/home"
		}
		*dbFlag = filepath.Join(home, ".casaos", "db")
	}

	_ = os.MkdirAll(*dbFlag, 0755)

	sqliteDB = sqlite.GetDb(*dbFlag)
	// gredis.GetRedisConn(config.RedisInfo),

	service.MyService = service.NewService(sqliteDB, config.CommonInfo.RuntimePath)

	service.Cache = cache.Init()

	service.GetCPUThermalZone()

	route.InitFunction()

	//service.MyService.System().GenreateSystemEntry()
	///
	//service.MountLists = make(map[string]*mountlib.MountPoint)
	//configfile.Install()
}

// @title casaOS API
// @version 1.0.0
// @contact.name lauren.pan
// @contact.url https://www.zimaboard.com
// @contact.email lauren.pan@icewhale.org
// @description casaOS v1版本api
// @host 192.168.2.217:8089
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @BasePath /v1
func main() {
	if *versionFlag {
		return
	}
	v1Router := route.InitV1Router()

	v2Router := route.InitV2Router()
	v2DocRouter := route.InitV2DocRouter(_docHTML, _docYAML)
	v3File := route.InitFile()
	mux := &util_http.HandlerMultiplexer{
		HandlerMap: map[string]http.Handler{
			"v1":  v1Router,
			"v2":  v2Router,
			"v3":  v3File,
			"doc": v2DocRouter,
		},
	}

	crontab := cron.New(cron.WithSeconds())
	if _, err := crontab.AddFunc("@every 5s", route.SendAllHardwareStatusBySocket); err != nil {
		logger.Error("add crontab error", zap.Error(err))
	}

	crontab.Start()
	defer crontab.Stop()

	managementFile := filepath.Join(config.CommonInfo.RuntimePath, "management.url")
	var listener net.Listener
	var err error

	port := "8080"
	if config.ServerInfo.HttpPort != "" {
		port = config.ServerInfo.HttpPort
	}

	if _, statErr := os.Stat(managementFile); statErr == nil {
		// External gateway is present; listen on ephemeral local port and register routes
		listener, err = net.Listen("tcp", net.JoinHostPort(LOCALHOST, "0"))
	} else {
		// Standalone mode: listen directly on standard port (0.0.0.0:8080) so UI/API are reachable
		listener, err = net.Listen("tcp", ":"+port)
		if err != nil {
			// Fallback to ephemeral port if preferred port is already bound
			listener, err = net.Listen("tcp", net.JoinHostPort(LOCALHOST, "0"))
		}
	}
	if err != nil {
		panic(err)
	}
	routers := []string{
		"/v1/sys",
		"/v1/port",
		"/v1/file",
		"/v1/folder",
		"/v1/batch",
		"/v1/image",
		"/v1/samba",
		"/v1/notify",
		"/v1/driver",
		"/v1/cloud",
		"/v1/recover",
		"/v1/other",
		"/v1/zt",
		"/v1/test",
		route.V2APIPath,
		route.V2DocPath,
		route.V3FilePath,
	}
	for _, apiPath := range routers {
		err = service.MyService.Gateway().CreateRoute(&model.Route{
			Path:   apiPath,
			Target: "http://" + listener.Addr().String(),
		})
		if err != nil {
			fmt.Printf("warn: failed to register route %s with gateway: %v\n", apiPath, err)
		}
	}

	// register at message bus
	for i := 0; i < 10; i++ {
		response, err := service.MyService.MessageBus().RegisterEventTypesWithResponse(context.Background(), common.EventTypes)
		if err != nil {
			logger.Error("error when trying to register one or more event types - some event type will not be discoverable", zap.Error(err))
		}
		if response != nil && response.StatusCode() != http.StatusOK {
			logger.Error("error when trying to register one or more event types - some event type will not be discoverable", zap.String("status", response.Status()), zap.String("body", string(response.Body)))
		}
		if response.StatusCode() == http.StatusOK {
			break
		}
		time.Sleep(time.Second)
	}

	go func() {
		time.Sleep(time.Second * 2)
		// v0.3.6
		if config.ServerInfo.HttpPort != "" {
			changePort := model.ChangePortRequest{}
			changePort.Port = config.ServerInfo.HttpPort
			err := service.MyService.Gateway().ChangePort(&changePort)
			if err == nil {
				config.Cfg.Section("server").Key("HttpPort").SetValue("")
				config.Cfg.SaveTo(config.SystemConfigInfo.ConfigPath)
			}
		}
	}()

	urlFilePath := filepath.Join(config.CommonInfo.RuntimePath, "casaos.url")
	if err := file.CreateFileAndWriteContent(urlFilePath, "http://"+listener.Addr().String()); err != nil {
		logger.Error("error when creating address file", zap.Error(err),
			zap.Any("address", listener.Addr().String()),
			zap.Any("filepath", urlFilePath),
		)
	}

	// run any script that needs to be executed
	scriptDirectory := filepath.Join(constants.DefaultConfigPath, "start.d")
	command.ExecuteScripts(scriptDirectory)

	if supported, err := platform.NotifyReadiness(); err != nil {
		logger.Error("Failed to notify supervisor that casaos main service is ready", zap.Any("error", err))
	} else if supported {
		logger.Info("Notified supervisor that casaos main service is ready")
	} else {
		logger.Info("This process is not running as a supervised service.")
	}
	// http.HandleFunc("/v1/file/test", func(w http.ResponseWriter, r *http.Request) {

	// 	//http.ServeFile(w, r, r.URL.Path[1:])
	// 	http.ServeFile(w, r, "/DATA/test.img")
	// })
	// go http.ListenAndServe(":8081", nil)

	_ = mime.AddExtensionType(".js", "application/javascript")
	_ = mime.AddExtensionType(".mjs", "application/javascript")
	_ = mime.AddExtensionType(".css", "text/css")
	_ = mime.AddExtensionType(".svg", "image/svg+xml")
	_ = mime.AddExtensionType(".json", "application/json")
	_ = mime.AddExtensionType(".webmanifest", "application/manifest+json")
	_ = mime.AddExtensionType(".woff2", "font/woff2")
	_ = mime.AddExtensionType(".woff", "font/woff")
	_ = mime.AddExtensionType(".ttf", "font/ttf")

	uiCandidates := []string{}
	if platform.Provider != nil && platform.Provider.WWWDir() != "" {
		uiCandidates = append(uiCandidates, platform.Provider.WWWDir())
	}
	if home := os.Getenv("HOME"); home != "" {
		uiCandidates = append(uiCandidates, filepath.Join(home, ".casaos", "www"))
	}
	uiCandidates = append(uiCandidates, "/data/data/com.termux/files/home/.casaos/www")
	if prefix := os.Getenv("PREFIX"); prefix != "" {
		uiCandidates = append(uiCandidates, filepath.Join(prefix, "var", "lib", "casaos", "www"))
	}
	uiCandidates = append(uiCandidates, "/var/lib/casaos/www")

	wwwDir := ""
	for _, cand := range uiCandidates {
		if cand == "" {
			continue
		}
		if fi, err := os.Stat(filepath.Join(cand, "index.html")); err == nil && !fi.IsDir() {
			wwwDir = cand
			break
		}
	}
	if wwwDir == "" {
		if platform.Provider != nil && platform.Provider.WWWDir() != "" {
			wwwDir = platform.Provider.WWWDir()
		} else {
			home := os.Getenv("HOME")
			if home == "" {
				home = "/data/data/com.termux/files/home"
			}
			wwwDir = filepath.Join(home, ".casaos", "www")
		}
	}
	_ = os.MkdirAll(wwwDir, 0755)

	fileServer := http.FileServer(http.Dir(wwwDir))
	indexPath := filepath.Join(wwwDir, "index.html")

	var rootHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		firstSegment := strings.Split(path, "/")[0]

		if firstSegment == "ping" {
			v1Router.ServeHTTP(w, r)
			return
		}

		if _, ok := mux.HandlerMap[firstSegment]; ok {
			mux.ServeHTTP(w, r)
			return
		}

		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>CasaOS Android</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #0f172a; color: #f8fafc; display: flex; align-items: center; justify-content: center; min-height: 100vh; margin: 0; padding: 20px; box-sizing: border-box; }
        .card { background: #1e293b; border-radius: 12px; padding: 32px; max-width: 640px; width: 100%; box-shadow: 0 10px 25px rgba(0,0,0,0.5); border: 1px solid #334155; }
        h1 { color: #38bdf8; margin-top: 0; font-size: 22px; }
        p { color: #94a3b8; font-size: 14px; line-height: 1.6; }
        pre { background: #0b0f19; padding: 14px; border-radius: 8px; overflow-x: auto; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: 12px; color: #4ade80; border: 1px solid #1e293b; white-space: pre-wrap; word-break: break-all; }
        .badge { display: inline-block; background: #0284c7; color: #fff; padding: 4px 10px; border-radius: 9999px; font-size: 11px; font-weight: bold; margin-bottom: 12px; }
    </style>
</head>
<body>
    <div class="card">
        <span class="badge">CasaOS Android Backend Running</span>
        <h1>CasaOS Dashboard Assets Missing</h1>
        <p>The CasaOS server daemon is active and listening, but frontend static assets are missing at <code>` + wwwDir + `</code>.</p>
        <p>Run the following command in Termux to download and install the official UI assets:</p>
        <pre>TMP_UI=$(mktemp -d) &amp;&amp; curl -sSL https://github.com/IceWhaleTech/CasaOS-UI/releases/download/v0.4.25/linux-all-casaos-v0.4.25.tar.gz | tar -xz -C "$TMP_UI" &amp;&amp; mkdir -p ~/.casaos/www &amp;&amp; cp -rf "$TMP_UI"/build/sysroot/var/lib/casaos/www/* ~/.casaos/www/ &amp;&amp; rm -rf "$TMP_UI"</pre>
        <p>Once completed, refresh this page to load your dashboard.</p>
    </div>
</body>
</html>`))
			return
		}

		targetFile := filepath.Join(wwwDir, filepath.Clean("/"+path))
		if fi, err := os.Stat(targetFile); err == nil && !fi.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}

		http.ServeFile(w, r, indexPath)
	})

	s := &http.Server{
		Handler:           rootHandler,
		ReadHeaderTimeout: 5 * time.Second, // fix G112: Potential slowloris attack (see https://github.com/securego/gosec)
	}

	logger.Info("CasaOS main service is listening...", zap.Any("address", listener.Addr().String()))
	// defer service.MyService.Storage().UnmountAllStorage()
	err = s.Serve(listener) // not using http.serve() to fix G114: Use of net/http serve function that has no support for setting timeouts (see https://github.com/securego/gosec)
	if err != nil {
		panic(err)
	}
}
