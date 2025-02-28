package gearbox

import (
	"context"
	"fmt"
	"gearbox/api"
	"gearbox/global"
	"github.com/gearboxworks/go-osbridge"
	"github.com/gearboxworks/go-status/only"
	"github.com/zserge/lorca"
	"github.com/zserge/webview"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
)

//
// [1] https://hackernoon.com/how-to-create-a-web-server-in-go-a064277287c9
// [2] https://github.com/zserge/webview
// [3] https://github.com/zserge/lorca
//

const (
	UiWindowTitle = "Gearbox"
	UiHeight      = 800
	UiWidth       = 1024
	UiResizable   = true
	UiDebug       = true
)

type UiWindow struct {
	Title        string
	Height       int
	Width        int
	NotResizable bool
	NoDebug      bool
}

type ViewerType string

const (
	DefaultViewer            = LorcaViewer
	WebViewViewer ViewerType = "webview"
	LorcaViewer   ViewerType = "lorca"
)

type AdminUi struct {
	ViewerType  ViewerType
	webListener net.Listener
	OsBridge    osbridge.OsBridger
	Gearbox     Gearboxer
	webServer   *http.Server
	api         api.Apier
	Window      *UiWindow
	ErrorLog    *ErrorLog
}

type UiWindowArgs struct {
	Title  string
	Height int
	Width  int
}

func NewUiWindow(args *UiWindowArgs) *UiWindow {
	if args.Title != "" {
		// If args.Title contains %s it will replace with UiWindowTitle's value
		// If not, it will just use value of args.Title
		args.Title = fmt.Sprintf(args.Title, UiWindowTitle)
	}
	if args.Height == 0 {
		args.Height = UiHeight
	}
	if args.Width == 0 {
		args.Width = UiWidth
	}
	return &UiWindow{
		Title:  args.Title,
		Height: args.Height,
		Width:  args.Width,
	}
}

func NewAdminUi(gearbox Gearboxer, viewer ViewerType) *AdminUi {
	ui := AdminUi{
		Gearbox:    gearbox,
		OsBridge:   gearbox.GetOsBridge(),
		ViewerType: viewer,
		Window: NewUiWindow(&UiWindowArgs{
			Title: "%s - " + fmt.Sprintf("[%s]", viewer),
		}),
	}
	return &ui
}

func (me *AdminUi) Initialize() {
	me.webListener = me.GetWebListener()
	me.webServer = me.GetWebServer()
	me.api = me.Gearbox.GetApi()
	me.WriteApiBaseUrls()
	me.ErrorLog = &ErrorLog{Gearbox: me.Gearbox}
	log.SetOutput(me.ErrorLog)
}

func (me *AdminUi) StartApi() {
	go me.api.Start()
}

func (me *AdminUi) Start() {
	go me.StartApi()
	go me.ServeWeb()
	switch me.ViewerType {
	case WebViewViewer:
		me.StartWebView()
	case LorcaViewer:
		me.StartLorca()
	default:
		log.Printf("invalid viewer type '%s'\n", me.ViewerType)
	}
}

func (me *AdminUi) StartLorca() {
	win := me.Window
	//time.Sleep(time.Second*3)
	ui, err := lorca.New(
		me.GetWebRootFileUrl(),
		string(me.GetWebRootDir()),
		win.Width,
		win.Height,
	)
	if err != nil {
		log.Printf("error loading Lorca to view Gearbox Admin UI: %s\n", err)
	}
	<-ui.Done()
}

func (me *AdminUi) StartWebView() {
	win := me.Window
	wv := webview.New(webview.Settings{
		Title:     win.Title,
		Height:    win.Height,
		Width:     win.Width,
		Resizable: !win.NotResizable,
		Debug:     !win.NoDebug,
		URL:       me.GetWebRootFileUrl(),
	})
	wv.Run()
}

func (me *AdminUi) WriteApiBaseUrls() {
	var err error
	url := me.api.GetBaseUrl()
	file := me.GetApiBaseUrlsFilepath()
	err = ioutil.WriteFile(file, NewApiBaseurls(url, url).Bytes(), os.ModePerm)
	if err != nil {
		log.Printf("error writing API bootrap file '%s': %s\n",
			me.GetApiBaseUrlsFilepath(),
			err,
		)
	}
}

func (me *AdminUi) GetApiBaseUrlsFilepath() string {
	return fmt.Sprintf("%s/api.json", me.GetWebRootDir())
}

func (me *AdminUi) GetWebListener() net.Listener {
	var err error
	for range only.Once {
		if me.webListener != nil {
			break
		}
		me.webListener, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			log.Printf("error initiating a TCP connection for AdminUi on '127.0.0.0:0': %s\n", err)
		}
		fmt.Printf("Starting %s admin console...", global.Brandname)
		if me.Gearbox.IsDebug() {
			fmt.Printf("\nListening on %s", me.GetHostname())
		}
	}
	return me.webListener
}

func (me *AdminUi) GetHostname() string {
	return me.webListener.Addr().String()
}

func (me *AdminUi) GetWebRootUrl() string {
	return fmt.Sprintf("http://%s", me.GetHostname())
}

func (me *AdminUi) GetWebRootFileUrl() string {
	return fmt.Sprintf("%s/index.html", me.GetWebRootUrl())
}

func (me *AdminUi) GetWebRootDir() http.Dir {
	return http.Dir(me.OsBridge.GetAdminRootDir())
}

func (me *AdminUi) GetWebRootFileDir() string {
	return filepath.FromSlash(fmt.Sprintf("%s/index.html", me.GetWebRootDir()))
}

func (me *AdminUi) GetWebHandler() http.Handler {
	return addCorsMiddleware(me.GetWebRootUrl(), http.FileServer(me.GetWebRootDir()))
}

func addCorsMiddleware(rooturl string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", rooturl)
		w.Header().Add("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Add("Access-Control-Allow-Headers", "Accept,Content-Type,Content-Length,Accept-Encoding,X-CSRF-Token,Authorization")
		next.ServeHTTP(w, r)
	})
}

func (me *AdminUi) shutdownServer(srv *http.Server) {
	fmt.Printf("Stopping %s admin console.\n", global.Brandname)
	err := srv.Shutdown(context.TODO())
	if err != nil {
		panic(err) // failure/timeout shutting down the server gracefully
	}
}

func (me *AdminUi) GetWebServer() *http.Server {
	for range only.Once {
		if me.webServer != nil {
			break
		}
		addr := me.GetHostname()
		me.webServer = &http.Server{
			Addr:    addr,
			Handler: me.GetWebHandler(),
		}
	}
	return me.webServer
}

func (me *AdminUi) GetApi() api.Apier {
	for range only.Once {
		if me.api != nil {
			break
		}
		me.api = me.Gearbox.GetApi()
	}
	return me.api
}

func (me *AdminUi) Close() {
	me.shutdownServer(me.webServer)
	//err := me.webListener.Close()
	//if err != nil {
	//	log.Warnf("error attempting to close AdminUi: %s", err)
	//}
	me.api.Stop()
}

func (me *AdminUi) ServeWeb() {
	err := me.webServer.Serve(me.webListener)
	// returns ErrServerClosed on graceful close
	if err != http.ErrServerClosed {
		// NOTE: there is a chance that next line won't have time to run,
		// as main() doesn't wait for this goroutine to stop. don't use
		// code with race conditions like these for production. see post
		// comments below on more discussion on how to handle this.
		log.Fatalf("error closing http.Server in AdminUi.ServeWeb(): %s", err)
	}
}
