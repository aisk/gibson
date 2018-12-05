package vox

import (
	"encoding/json"
	"net/http"
)

// An Application is a container which includes middlewares and config, and implemented the GO's net/http.Handler interface https://golang.org/pkg/net/http/#Handler.
type Application struct {
	middlewares []Handler
	configs     map[string]string
}

// New returns a new vox Application.
func New() *Application {
	app := &Application{
		middlewares: []Handler{},
		configs:     map[string]string{},
	}
	return app
}

// Use a vox middleware.
func (app *Application) Use(handler Handler) {
	app.middlewares = append(app.middlewares, handler)
}

// SetConfig sets an application level variable.
func (app *Application) SetConfig(key, value string) {
	app.configs[key] = value
}

// GetConfig application level variable by key.
func (app *Application) GetConfig(key string) string {
	return app.configs[key]
}

func (app *Application) ServeHTTP(rw http.ResponseWriter, rq *http.Request) {
	handler := compose(app.middlewares)
	req := createRequest(rq)
	res := createResponse(rw)
	res.request = req
	handler(req, res)
	if !res.DontRespond {
		respond(res)
	}
}

// Run the Vox application.
func (app *Application) Run(addr string) error {
	return http.ListenAndServe(addr, app)
}

func compose(middlewares []Handler) Handler {
	return func(req *Request, res *Response) {
		next := func() {}
		for i := len(middlewares) - 1; i >= 0; i-- {
			func(i int, nenext func()) {
				next = func() {
					req.Next = nenext
					middlewares[i](req, res)
				}
			}(i, next)
		}
		next()
	}
}

func respond(res *Response) {
	res.setImplicit()

	res.Writer.WriteHeader(res.Status)

	switch v := res.Body.(type) {
	case []byte:
		res.Writer.Write(v)
	case string:
		res.Writer.Write([]byte(v))
	// TODO: support io.Reader type
	default:
		body, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		res.Writer.Write(body)
	}
}
