# Process manager example

The example consists of three parts:
* simple process managing like creating a process, start, stop, restart of status watching
* more advanced status handling as attaching to running processes, reading of process status file or automatic restarts
* all about templates, how to create and use it

The example uses [test-application](test-process/test-process.go) called test-process, which is managed during
the example.

Note: in order to run a part of the example which handles templates, process manager config file needs to be provided
to the example, otherwise it will be skipped.

```
./process-manager-plugin -process-manager-config=<path-to-file>
```

All about process manager config file can be found in process manager [readme](../../process/README.md#Templates)

