/*
The MIT License (MIT)

Copyright (c) 2014-2017 DutchCoders [https://github.com/dutchcoders/]

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package server

import (
	crypto_rand "crypto/rand"
	"encoding/binary"
	"log"
	"math/rand"
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	gorillaHandlers "github.com/gorilla/handlers"

	"github.com/PuerkitoBio/ghost/handlers"
	"github.com/VojtechVitek/ratelimit"
	"github.com/VojtechVitek/ratelimit/memory"
	"github.com/gorilla/mux"

	_ "net/http/pprof"

	"crypto/tls"

	web "github.com/dutchcoders/transfer.sh-web"
	assetfs "github.com/elazarl/go-bindata-assetfs"
)

const SERVER_INFO = "transfer.sh"
const ServerAuthKey = "serverAuthKey"

// parse request with maximum memory of _24Kilobits
const _24K = (1 << 3) * 24

// parse request with maximum memory of _5Megabytes
const _5M = (1 << 20) * 5

type Server struct {
	auths map[string]Authenticator

	logger *log.Logger

	tlsConfig *tls.Config

	profilerEnabled bool

	locks map[string]*sync.Mutex

	rateLimitRequests int

	storage         Storage
	metadataStorage Storage

	forceHTTPs bool

	ipFilterOptions *IPFilterOptions

	VirusTotalKey    string
	ClamAVDaemonHost string

	tempPath string

	webPath      string
	proxyPath    string
	proxyPort    string
	gaKey        string
	userVoiceKey string

	TLSListenerOnly bool

	CorsDomains           string
	ListenerString        string
	TLSListenerString     string
	ProfileListenerString string

	Certificate string

	LetsEncryptCache string
}

type DefaultServAuthenticator struct {
	user     string
	password string
}

func New(options ...OptionFn) (*Server, error) {
	s := &Server{
		auths: make(map[string]Authenticator),
		locks: map[string]*sync.Mutex{},
	}

	for _, optionFn := range options {
		optionFn(s)
	}

	return s, nil
}

func init() {
	var seedBytes [8]byte
	if _, err := crypto_rand.Read(seedBytes[:]); err != nil {
		panic("cannot obtain cryptographically secure seed")
	}
	rand.Seed(int64(binary.LittleEndian.Uint64(seedBytes[:])))
}

func (s *Server) Run() {
	listening := false

	if s.profilerEnabled {
		listening = true

		go func() {
			s.logger.Println("Profiled listening at: :6060")

			http.ListenAndServe(":6060", nil)
		}()
	}

	r := mux.NewRouter()

	var fs http.FileSystem

	if s.webPath != "" {
		s.logger.Println("Using static file path: ", s.webPath)

		fs = http.Dir(s.webPath)

		htmlTemplates, _ = htmlTemplates.ParseGlob(s.webPath + "*.html")
		textTemplates, _ = textTemplates.ParseGlob(s.webPath + "*.txt")
	} else {
		fs = &assetfs.AssetFS{
			Asset:    web.Asset,
			AssetDir: web.AssetDir,
			AssetInfo: func(path string) (os.FileInfo, error) {
				return os.Stat(path)
			},
			Prefix: web.Prefix,
		}

		for _, path := range web.AssetNames() {
			bytes, err := web.Asset(path)
			if err != nil {
				s.logger.Panicf("Unable to parse: path=%s, err=%s", path, err)
			}

			htmlTemplates.New(stripPrefix(path)).Parse(string(bytes))
			textTemplates.New(stripPrefix(path)).Parse(string(bytes))
		}
	}

	staticHandler := http.FileServer(fs)

	r.PathPrefix("/images/").Handler(staticHandler).Methods("GET")
	r.PathPrefix("/styles/").Handler(staticHandler).Methods("GET")
	r.PathPrefix("/scripts/").Handler(staticHandler).Methods("GET")
	r.PathPrefix("/fonts/").Handler(staticHandler).Methods("GET")
	r.PathPrefix("/ico/").Handler(staticHandler).Methods("GET")
	r.HandleFunc("/favicon.ico", staticHandler.ServeHTTP).Methods("GET")
	r.HandleFunc("/robots.txt", staticHandler.ServeHTTP).Methods("GET")

	r.HandleFunc("/{filename:(?:favicon\\.ico|robots\\.txt|health\\.html)}", s.BasicAuthHandler(http.HandlerFunc(s.putHandler))).Methods("PUT")

	r.HandleFunc("/health.html", healthHandler).Methods("GET")
	r.HandleFunc("/", s.viewHandler).Methods("GET")

	r.HandleFunc("/({files:.*}).zip", s.zipHandler).Methods("GET")
	r.HandleFunc("/({files:.*}).tar", s.tarHandler).Methods("GET")
	r.HandleFunc("/({files:.*}).tar.gz", s.tarGzHandler).Methods("GET")

	r.HandleFunc("/{token}/{filename}", s.headHandler).Methods("HEAD")
	r.HandleFunc("/{action:(?:download|get|inline)}/{token}/{filename}", s.headHandler).Methods("HEAD")

	r.HandleFunc("/{token}/{filename}", s.previewHandler).MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) (match bool) {
		match = false

		// The file will show a preview page when opening the link in browser directly or
		// from external link. If the referer url path and current path are the same it will be
		// downloaded.
		if !acceptsHTML(r.Header) {
			return false
		}

		match = (r.Referer() == "")

		u, err := url.Parse(r.Referer())
		if err != nil {
			s.logger.Fatal(err)
			return
		}

		match = match || (u.Path != r.URL.Path)
		return
	}).Methods("GET")

	getHandlerFn := s.getHandler
	if s.rateLimitRequests > 0 {
		getHandlerFn = ratelimit.Request(ratelimit.IP).Rate(s.rateLimitRequests, 60*time.Second).LimitBy(memory.New())(http.HandlerFunc(getHandlerFn)).ServeHTTP
	}

	r.HandleFunc("/{token}/{filename}",
		s.AssignMetadata(MetadataAllowedIP(s.MetadataBasicAuth(http.HandlerFunc(getHandlerFn)))),
	).Methods("GET")
	r.HandleFunc("/{action:(?:download|get|inline)}/{token}/{filename}",
		s.AssignMetadata(MetadataAllowedIP(s.MetadataBasicAuth(http.HandlerFunc(getHandlerFn)))),
	).Methods("GET")

	r.HandleFunc("/{filename}/virustotal", s.virusTotalHandler).Methods("PUT")
	r.HandleFunc("/{filename}/scan", s.scanHandler).Methods("PUT")
	r.HandleFunc("/put/{filename}", s.BasicAuthHandler(http.HandlerFunc(s.putHandler))).Methods("PUT")
	r.HandleFunc("/upload/{filename}", s.BasicAuthHandler(http.HandlerFunc(s.putHandler))).Methods("PUT")
	r.HandleFunc("/{filename}", s.BasicAuthHandler(http.HandlerFunc(s.putHandler))).Methods("PUT")
	r.HandleFunc("/", s.BasicAuthHandler(http.HandlerFunc(s.postHandler))).Methods("POST")
	// r.HandleFunc("/{page}", viewHandler).Methods("GET")

	r.HandleFunc("/{token}/{filename}/{deletionToken}", s.deleteHandler).Methods("DELETE")

	r.NotFoundHandler = http.HandlerFunc(s.notFoundHandler)

	mime.AddExtensionType(".md", "text/x-markdown")

	s.logger.Printf("Transfer.sh server started.\nusing temp folder: %s\nusing storage provider: %s", s.tempPath, s.storage.Type())

	var cors func(http.Handler) http.Handler
	if len(s.CorsDomains) > 0 {
		cors = gorillaHandlers.CORS(
			gorillaHandlers.AllowedHeaders([]string{"*"}),
			gorillaHandlers.AllowedOrigins(strings.Split(s.CorsDomains, ",")),
			gorillaHandlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"}),
		)
	} else {
		cors = func(h http.Handler) http.Handler {
			return h
		}
	}

	h := handlers.PanicHandler(
		IPFilterHandler(
			handlers.LogHandler(
				LoveHandler(
					s.RedirectHandler(cors(r))),
				handlers.NewLogOptions(s.logger.Printf, "_default_"),
			),
			s.ipFilterOptions,
		),
		nil,
	)

	if !s.TLSListenerOnly {
		srvr := &http.Server{
			Addr:    s.ListenerString,
			Handler: h,
		}

		listening = true
		s.logger.Printf("listening on port: %v\n", s.ListenerString)

		go func() {
			srvr.ListenAndServe()
		}()
	}

	if s.TLSListenerString != "" {
		listening = true
		s.logger.Printf("listening on port: %v\n", s.TLSListenerString)

		go func() {
			s := &http.Server{
				Addr:      s.TLSListenerString,
				Handler:   h,
				TLSConfig: s.tlsConfig,
			}

			if err := s.ListenAndServeTLS("", ""); err != nil {
				panic(err)
			}
		}()
	}

	s.logger.Printf("---------------------------")

	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt)
	signal.Notify(term, syscall.SIGTERM)

	if listening {
		<-term
	} else {
		s.logger.Printf("No listener active.")
	}

	s.logger.Printf("Server stopped.")
}

func (d *DefaultServAuthenticator) Set(user, password string) {
	d.user = user
	d.password = password
}

func (d *DefaultServAuthenticator) Authenticate(user, password string) (bool, error) {
	if d.user == user && password == password {
		return true, nil
	}
	return false, nil
}
