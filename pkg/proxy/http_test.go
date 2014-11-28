package proxy

import (
	"net"
	"net/http"
	"net/url"
	"strconv"
	"testing"
)

const NORMAL_PORT = 12354
const PROXY_PORT = 12355

type serverCb func(w http.ResponseWriter, r *http.Request)

type server struct {
	listener *net.Listener
	callback serverCb
}

func TestHttpRewriteHost(t *testing.T) {
	client := &http.Client{}
	settings := &Settings{false, true}

	incoming := make(chan string, 1)
	proxyHost := "localhost:" + strconv.Itoa(PROXY_PORT)
	normalHost := "localhost:" + strconv.Itoa(NORMAL_PORT)

	// Create normal server
	normal := newServer(func(w http.ResponseWriter, r *http.Request) {
		incoming <- r.Host
		w.Write([]byte("peace out"))
	})

	// Create proxy server
	proxy := newServer(func(w http.ResponseWriter, r *http.Request) {
		proxyURL := url.URL{Scheme: "http", Host: proxyHost, Path: "/foo"}
		destURL := url.URL{Scheme: "http", Host: normalHost, Path: "/bar"}
		ctx := Context{r, w, &proxyURL, &destURL, settings}
		ProxyHTTP(&ctx, client)
	})

	// Start servers
	if err := normal.start(NORMAL_PORT); err != nil {
		t.Fatalf("Failed to listen on port %d: %s", NORMAL_PORT, err.Error())
	}
	defer normal.stop()
	if err := proxy.start(PROXY_PORT); err != nil {
		t.Fatalf("Failed to listen on port %d: %s", PROXY_PORT, err.Error())
	}
	defer proxy.stop()

	// Host rewrite enabled
	res, err := http.Get("http://" + proxyHost + "/foo")
	if err != nil {
		t.Fatal("Failed to make request:", err)
	}
	res.Body.Close()
	if gotHost := <-incoming; gotHost != normalHost {
		t.Errorf("(Host rewrite enabled) expected %s got %s", normalHost,
			gotHost)
	}

	// Host rewrite disabled
	settings.RewriteHost = false
	res, err = http.Get("http://" + proxyHost + "/foo")
	if err != nil {
		t.Fatal("Failed to make request:", err)
	}
	res.Body.Close()
	if gotHost := <-incoming; gotHost != proxyHost {
		t.Errorf("(Host rewrite disabled) expected %s got %s", proxyHost,
			gotHost)
	}
}

func newServer(callback serverCb) *server {
	return &server{nil, callback}
}

func (self *server) start(port int) error {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return err
	}

	self.listener = &listener
	go http.Serve(listener, self)
	return nil
}

func (self *server) stop() {
	(*self.listener).Close()
	self.listener = nil
}

func (self *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	self.callback(w, r)
}
