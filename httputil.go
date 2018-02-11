package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func serveGopherJS(port int, srcRoot string, module ...string) error {
	if len(srcRoot) == 0 || srcRoot[0] != '/' {
		srcRoot = "/" + srcRoot
	}

	proxy, err := proxyGopherJS(port)
	if err != nil {
		return err
	}

	// handle for sources
	http.Handle(srcRoot, proxy)

	// modules
	for _, m := range module {
		pfx := srcRoot + "/" + m
		http.Handle("/"+m+".js", addPrefix(pfx, proxy))
		http.Handle("/"+m+".js.map", addPrefix(pfx, proxy))
	}

	return nil
}

// proxyGopherJS starts a gopherjs on the specified port,
// waits for it to be available and returns a reverse proxy
// handler for it.
func proxyGopherJS(port int) (http.Handler, error) {
	ports := fmt.Sprintf(":%d", port)
	cmd := exec.Command("gopherjs", "serve", "--http="+ports)

	// set gopherjs GOOS to darwin, see
	// https://github.com/gopherjs/gopherjs/issues/688
	if runtime.GOOS == "windows" {
		cmd.Env = append(os.Environ(), "GOOS=darwin")
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	u, err := url.Parse("http://localhost" + ports + "/")
	if err != nil {
		return nil, err
	}

	uu, err := u.Parse("/encoding.js")
	if err != nil {
		return nil, err
	}

	err = tryURL(uu, 30*time.Second)

	return httputil.NewSingleHostReverseProxy(u), err
}

func addPrefix(prefix string, h http.Handler) http.Handler {
	if prefix == "" {
		return h
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r2 := new(http.Request)
		*r2 = *r
		r2.URL = new(url.URL)
		*r2.URL = *r.URL
		r2.URL.Path = prefix + r.URL.Path
		h.ServeHTTP(w, r2)
	})
}

// templateDir is like http.Dir but applies
// the template arguments to html files.
type templateDir struct {
	root string
	data interface{} // template data
}

func (td *templateDir) Open(name string) (http.File, error) {
	log.Println("templateDir:", name)
	f, err := http.Dir(td.root).Open(name)
	if err != nil {
		return nil, err
	}
	if !strings.HasSuffix(name, ".html") {
		return f, nil
	}

	defer f.Close()

	src, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	t, err := template.New(name).Parse(string(src))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	buf := new(bytes.Buffer)
	if err := t.Execute(buf, td.data); err != nil {
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	return newFakeFile(fi, bytes.NewReader(buf.Bytes())), nil
}

type fakeFile struct {
	name    string
	modTime time.Time

	r *bytes.Reader
}

func newFakeFile(fi os.FileInfo, r *bytes.Reader) *fakeFile {
	return &fakeFile{
		name:    fi.Name(),
		modTime: fi.ModTime(),
		r:       r,
	}
}

func (f *fakeFile) Close() error                                 { return nil }
func (f *fakeFile) Read(p []byte) (n int, err error)             { return f.r.Read(p) }
func (f *fakeFile) Seek(offset int64, whence int) (int64, error) { return f.r.Seek(offset, whence) }
func (f *fakeFile) Readdir(count int) ([]os.FileInfo, error)     { return nil, os.ErrInvalid }
func (f *fakeFile) Stat() (os.FileInfo, error)                   { return f, nil }

// fakeFile as os.FileInfo
func (f *fakeFile) Name() string       { return f.name }
func (f *fakeFile) Size() int64        { return f.r.Size() }
func (f *fakeFile) Mode() os.FileMode  { return 0666 }
func (f *fakeFile) ModTime() time.Time { return f.modTime }
func (f *fakeFile) IsDir() bool        { return false }
func (f *fakeFile) Sys() interface{}   { return nil }

func handleWithPrefix(pfx string, h http.Handler) {
	n := len(pfx) - 1
	if n < 0 || pfx[n] != '/' {
		panic("prefix must end in '/'")
	}
	http.Handle(pfx, http.StripPrefix(pfx[:n], h))
}

// https://gist.github.com/hyg/9c4afcd91fe24316cbf0
func openbrowser(addr string) error {
	var u string
	if x, err := url.Parse(addr); err == nil {
		if x.Host == "" && strings.HasPrefix(x.Scheme, "http") {
			x.Host = "localhost"
		}
		u = x.String()
	} else {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return fmt.Errorf("unrecognised addr %v", addr)
		}
		if host == "" {
			host = "localhost"
		}
		u = fmt.Sprintf("http://%s:%s", host, port)
	}

	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", u).Start()
	case "windows":
		err = exec.Command("rundll32", "u.dll,FileProtocolHandler", u).Start()
	case "darwin":
		err = exec.Command("open", u).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	return err
}

func tryURL(u *url.URL, timeout time.Duration) error {
	start := time.Now()
	for {
		_, err := http.Get(u.String())
		if err == nil {
			return err
		}
		elapsed := time.Now().Sub(start)
		if elapsed > timeout {
			return err
		}
		time.Sleep(time.Second)
	}
	panic("unreachable")
}
