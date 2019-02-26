# Tutorial: Hello World

In this tutorial we will create a simple Ligato control plane agent that 
contains a single `Helloworld` plugin that prints "Hello World" to the log.

We start with the plugin. Every plugin must implement the `Plugin` interface
defined in the `github.com/cn-infra/infra` package:
```go
type Plugin interface {
	// Init is called in the agent`s startup phase.
	Init() error
	// Close is called in the agent`s cleanup phase.
	Close() error
	// String returns unique name of the plugin.
	String() string
}
```

Let's define the above methods for our `HelloWorld` plugin:

```go
type HelloWorld struct{}

func (p *HelloWorld) String() string {
	return "HelloWorld"
}

func (p *HelloWorld) Init() error {
	log.Println("Hello World!")
	return nil
}

func (p *HelloWorld) Close() error {
	log.Println("Goodbye World!")
	return nil
}
```
Note that the `HelloWorld` struct is empty - our simple plugin does not 
have any data, so we just need an empty structure that satisfies the 
`Plugin` interface.

Some plugins may require additional initialization that can only be
performed after the base system is up (for example, ...). If your plugin
needs this, you can optionally define the `AfterInit` method for your
plugin. It will be executed after the `Init` method has been called for
all plugins. The `AfterInit` method comes from the `PostInit` interface
defined in the `github.com/cn-infra/infra` package as:

```go
type PostInit interface {
	// AfterInit is called once Init() of all plugins have returned without error.
	AfterInit() error
}
```

Next, in our main function we create an instance of the `HelloWorld` plugin. Then we 
create a new agent and tell it about the `HelloWorld` plugin:

```go
func main() {
    	p := new(HelloWorld)    
    	a := agent.NewAgent(agent.Plugins(p))
    	// ...
}
```

We use agent options tell the agent about its plugin(s). The option `agent.Plugins` 
adds single plugin instance to the list of plugins and the option `agent.AllPlugins` 
adds plugin instance along with all of its dependencies.

Now we can start the agent using `Run()` method which will initialize all of its
plugins by calling their `Init` and `AfterInit` methods, then wait until interrupt
and then stops all of this plugins by calling their `Close` methods.

```go
if err := a.Run(); err != nil {
	log.Fatalln(err)
}
```

Complete working example can be found at [examples/tutorials/01_hello-world](https://github.com/ligato/cn-infra/blob/master/examples/tutorials/01_hello-world).
