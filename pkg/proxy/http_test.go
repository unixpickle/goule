package proxy

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"testing"
)

const NORMAL_PORT = 12354
const PROXY_PORT = 12355
const NORMAL_PORT_2 = 12356
const PROXY_PORT_2 = 12357

type serverCb func(w http.ResponseWriter, r *http.Request)

type server struct {
	listener *net.Listener
	callback serverCb
}

func TestHttpRewriteHost(t *testing.T) {
	client := new(http.Client)
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

func TestHttpProxy(t *testing.T) {
	client := new(http.Client)
	settings := new(Settings)

	incoming := make(chan resInfo, 1)
	proxyHost := "localhost:" + strconv.Itoa(PROXY_PORT_2)
	normalHost := "localhost:" + strconv.Itoa(NORMAL_PORT_2)

	// Create normal server
	normal := newServer(func(w http.ResponseWriter, r *http.Request) {
		if data, err := ioutil.ReadAll(r.Body); err != nil {
			t.Fatal("Failed to read incoming data:", err)
		} else {
			incoming <- resInfo{data, r.Header}
		}
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(404)
		w.Write([]byte("Some kind of 404 page!"))
	})

	// Create proxy server
	proxy := newServer(func(w http.ResponseWriter, r *http.Request) {
		proxyURL := url.URL{Scheme: "http", Host: proxyHost, Path: "/foo"}
		destURL := url.URL{Scheme: "http", Host: normalHost, Path: "/bar"}
		ctx := Context{r, w, &proxyURL, &destURL, settings}
		ProxyHTTP(&ctx, client)
	})

	// Start servers
	if err := normal.start(NORMAL_PORT_2); err != nil {
		t.Fatalf("Failed to listen on port %d: %s", NORMAL_PORT_2, err.Error())
	}
	defer normal.stop()
	if err := proxy.start(PROXY_PORT_2); err != nil {
		t.Fatalf("Failed to listen on port %d: %s", PROXY_PORT_2, err.Error())
	}
	defer proxy.stop()

	// Get data
	res, err := http.Post("http://" + proxyHost + "/foo", "text/plain",
		bytes.NewBuffer([]byte("Request data")))
	if err != nil {
		t.Fatal("Failed to make request:", err)
	}
	gotData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal("Failed to read data:", gotData)
	}
	res.Body.Close()
	if !bytes.Equal(gotData, []byte("Some kind of 404 page!")) {
		t.Error("Got unexpected body:", string(gotData))
	}
	if res.StatusCode != 404 {
		t.Error("Expected status 404, got:", res.StatusCode)
	}
	
	info := <-incoming
	if !bytes.Equal(info.data, []byte("Request data")) {
		t.Error("Sent unexpected post data:", string(info.data))
	}
	if info.head.Get("Content-Type") != "text/plain" {
		t.Error("Invalid content-type:", info.head.Get("Content-Type"))
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

type resInfo struct {
	data []byte
	head http.Header
}
