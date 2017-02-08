package fakes

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
)

type FakeUAA struct {
	server *httptest.Server
	lock   sync.Mutex

	tokenType   string
	accessToken string

	requested bool
}

func NewFakeUAA(tokenType string, accessToken string) *FakeUAA {
	return &FakeUAA{
		tokenType:   tokenType,
		accessToken: accessToken,
	}
}

func (f *FakeUAA) Start() {
	f.server = httptest.NewUnstartedServer(f)
	f.server.Start()
}

func (f *FakeUAA) Close() {
	f.server.Close()
}

func (f *FakeUAA) URL() string {
	return f.server.URL
}

func (f *FakeUAA) Requested() bool {
	f.lock.Lock()
	defer f.lock.Unlock()
	return f.requested
}

func (f *FakeUAA) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	rw.Write([]byte(fmt.Sprintf(`
		{
			"token_type": "%s",
			"access_token": "%s"
		}
	`, f.tokenType, f.accessToken)))
	f.lock.Lock()
	f.requested = true
	f.lock.Unlock()
}

func (f *FakeUAA) AuthToken() string {
	if f.tokenType == "" && f.accessToken == "" {
		return ""
	}
	return fmt.Sprintf("%s %s", f.tokenType, f.accessToken)
}
