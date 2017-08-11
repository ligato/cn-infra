# HTTPmux

The `HTTPmux` is a Plugin which allows other plugins to handle HTTP requests.

**API**

To serve an HTTP service, plugin must first implement a handler function
and then register it at a given URL path using the `RegisterHTTPHandler`
method. Behind the scenes, `httpmux` runs an HTTP server inside a goroutine
and registers HTTP handlers by their URL path using an HTTP request 
multiplexer from the `gorilla/mux` package.

![http](../docs/imgs/http.png)

**Configuration**

- the server's port can be defined using commandline flag `http-port` or 
  the environment variable HTTP_PORT.

**Example**

The following example demonstrates the usage of the `httpmux` plugin API:
```
// httpExampleHandler returns a very simple HTTP request handler.
func httpExampleHandler(formatter *render.Render) http.HandlerFunc {

    // An example HTTP request handler which prints out attributes of 
    // a trivial Go structure in JSON format.
    return func(w http.ResponseWriter, req *http.Request) {
        formatter.JSON(w, http.StatusOK, struct{ Example string }{"This is an example"})
    }
}

// Register our HTTP request handler as a GET method serving at 
// the URL path "/example".
httpmux.RegisterHTTPHandler("/example", httpExampleHandler, "GET")
```

Once the handler is registered with `httpmux` and the agent is running, 
you can use `curl` to verify that it is operating properly:
```
$ curl -X GET http://localhost:9191/example
{
  "Example": "This is an example"
}
```

**Dependencies**

- [Logging](../logging/plugin)