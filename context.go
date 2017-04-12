package gweb

import (
	"github.com/chen-zyc/gweb/render"
	"net/http"
	"net/url"
)

// Param is a single URL parameter, consisting of a key and a value.
type Param struct {
	Key   string
	Value string
}

// Params is a Param-slice, as returned by the router.
// The slice is ordered, the first URL parameter is also the first slice value.
// It is therefore safe to read values by the index.
type Params []Param

// ByName returns the value of the first Param which key matches the given name.
// If no matching Param is found, an empty string is returned.
func (ps Params) ByName(name string) string {
	for i := range ps {
		if ps[i].Key == name {
			return ps[i].Value
		}
	}
	return ""
}

type Context struct {
	req             *http.Request
	resp            http.ResponseWriter
	params          Params
	handlers        Handlers
	curHandlerIndex int
	userData        map[string]interface{}
}

func (c *Context) reset(req *http.Request, resp http.ResponseWriter) {
	c.req = req
	c.resp = resp
	c.params = c.params[:0]
	c.handlers = nil
	c.curHandlerIndex = -1
}

func (c *Context) Next() {
	c.curHandlerIndex++
	// 使用for，这样的话即使handler没有主动调用 Next，也能够保证剩余的handler被执行。
	for ; c.curHandlerIndex < len(c.handlers); c.curHandlerIndex++ {
		c.handlers[c.curHandlerIndex](c)
	}
}

// =================================
// ======= input data ==============
// =================================

func (c *Context) Param(name string) string {
	return c.params.ByName(name)
}

func (c *Context) GetQueryArray(key string) (arr []string, exist bool) {
	arr, exist = c.req.URL.Query()[key]
	return
}

func (c *Context) QueryArray(key string) (arr []string) {
	arr, _ = c.GetQueryArray(key)
	return
}

func (c *Context) GetQuery(key string) (val string, exist bool) {
	if arr, ok := c.GetQueryArray(key); ok {
		exist = true
		if len(arr) > 0 {
			val = arr[0]
		}
	}
	return
}

func (c *Context) Query(key string) (val string) {
	val, _ = c.GetQuery(key)
	return
}

func (c *Context) DefaultQuery(key string, defVal string) string {
	val := c.Query(key)
	if val == "" {
		val = defVal
	}
	return val
}

func (c *Context) GetPostFormArray(key string) ([]string, bool) {
	c.req.ParseForm()
	arr, exist := c.req.PostForm[key]
	if exist && len(arr) > 0 {
		return arr, true
	}
	c.req.ParseMultipartForm(32 << 20) // 32MB
	if c.req.MultipartForm != nil && c.req.MultipartForm.File != nil {
		if arr = c.req.MultipartForm.Value[key]; len(arr) > 0 {
			return arr, true
		}
	}

	return nil, false
}

func (c *Context) PostFormArray(key string) []string {
	arr, _ := c.GetPostFormArray(key)
	return arr
}

func (c *Context) GetPostForm(key string) (string, bool) {
	arr, exist := c.GetPostFormArray(key)
	if len(arr) > 0 {
		return arr[0], exist
	}
	return "", exist
}

func (c *Context) PostForm(key string) string {
	val, _ := c.GetPostForm(key)
	return val
}

// =================================
// ======= response ================
// =================================

func (c *Context) Status(code int) {
	c.resp.WriteHeader(code)
}

func (c *Context) Header(key, value string) {
	if value == "" {
		c.resp.Header().Del(key)
	} else {
		c.resp.Header().Set(key, value)
	}
}

func (c *Context) Render(r render.Render) {
	if err := r.Render(c.resp); err != nil {
		panic(err)
	}
}

func (c *Context) String(code int, format string, args ...interface{}) {
	c.Status(code)
	c.Render(render.StringRender(format, args...))
}

func (c *Context) JSON(code int, obj interface{}) {
	c.Status(code)
	c.Render(render.JSONRender(obj))
}

func (c *Context) XML(code int, obj interface{}) {
	c.Status(code)
	c.Render(render.XMLRender(obj))
}

func (c *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if path == "" {
		path = "/"
	}
	http.SetCookie(c.resp, &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
}

func (c *Context) Cookie(name string) (string, error) {
	cookie, err := c.req.Cookie(name)
	if err != nil {
		return "", err
	}
	val, err := url.QueryUnescape(cookie.Value)
	return val, err
}

func (c *Context) RawCookie(name string) (*http.Cookie, error) {
	return c.req.Cookie(name)
}

// =================================
// ======= user data ===============
// =================================

func (c *Context) UserData(key string) (data interface{}, exist bool) {
	if c.userData != nil {
		data, exist = c.userData[key]
		return
	}
	return
}

func (c *Context) SetUserData(key string, val interface{}) {
	if c.userData == nil {
		c.userData = make(map[string]interface{})
	}
	c.userData[key] = val
}
