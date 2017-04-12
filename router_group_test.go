package gweb

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGroupBasic(t *testing.T) {
	s := NewServer()
	g := s.Group("/hello", emptyHandler)
	g.Global(emptyHandler)

	assert.Len(t, g.globalHandlers, 2)
	assert.Equal(t, g.basePath, "/hello")
	assert.Equal(t, g.s, s)

	g2 := g.Group("world", emptyHandler)
	g2.Global(emptyHandler)

	assert.Len(t, g2.globalHandlers, 4)
	assert.Equal(t, g2.basePath, "/hello/world")
	assert.Equal(t, g2.s, s)
}

func TestGroupBasicHandler(t *testing.T) {
	performRequestInGroup(t, MethodGet)
	performRequestInGroup(t, MethodPost)
	performRequestInGroup(t, MethodPut)
	performRequestInGroup(t, MethodPatch)
	performRequestInGroup(t, MethodDelete)
	performRequestInGroup(t, MethodHead)
	performRequestInGroup(t, MethodOptions)
}

func performRequestInGroup(t *testing.T, method string) {
	s := NewServer()
	v1 := s.Group("v1", emptyHandler)
	assert.Equal(t, v1.basePath, "/v1")

	login := v1.Group("/login/", emptyHandler, emptyHandler)
	assert.Equal(t, login.basePath, "/v1/login/")

	handler := func(c *Context) {
		c.String(http.StatusBadRequest, "the method was %s and index %d", c.req.Method, c.curHandlerIndex)
	}

	v1.Handle(method, "/test", handler)
	login.Handle(method, "/test", handler)

	w := performRequest(s, method, "/v1/login/test")
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Equal(t, w.Body.String(), "the method was "+method+" and index 3")

	w = performRequest(s, method, "/v1/test")
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Equal(t, w.Body.String(), "the method was "+method+" and index 1")
}

func emptyHandler(_ *Context) {}

func performRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
