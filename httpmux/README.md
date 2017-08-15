# HTTPmux

The `HTTPmux` is an infrastructure plugin which allows app plugins 
to handle HTTP requests (see following diagram) in this sequence:
1. httpmux starts the HTTP server
2. to serve a HTTP service, the plugin must first implement a handler function
and then register it at a given URL path using the `RegisterHTTPHandler`
method. Behind the scenes, `httpmux` runs a HTTP server inside a goroutine
and registers HTTP handlers by their URL path using a HTTP request 
multiplexer from the `gorilla/mux` package.
3. the HTPP server using `gorilla/mux` asks the previously registered handler to 
   handle a particular HTTP request.

![http](../docs/imgs/http.png)

**Configuration**

- the server's port can be defined using command line flag `http-port` or 
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
