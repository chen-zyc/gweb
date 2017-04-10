package gweb

type Router interface {
	Add(path string, handler Handlers)
	Find(path string) (handler Handlers, params Params, tsr bool)
	FindCaseInsensitivePath(path string, fixTrailingSlash bool) (ciPath []byte, found bool)
}

func NewRouter() Router { return NewPrefixTree() }
