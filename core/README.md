## Plugin API

This package provides the API between Agent Core and Agent Plugins (illustrated also on following diagram).

```
                                       +-----------------------+
+------------------+                   |                       |
|                  |                   |      Agent Plugin     |
|                  |                   |                       |
|    Agent Core    |                   +-----------------------+
|     (setup)      |        +--------->| Plugin global var     |
|                  |        |          |   + Init() error      |
|                  |        |          |   + AfterInit() error |
|                  |        |          |   + Close() error     |
|                  |        |          +-----------------------+
+------------------+        |
|                  +--------+          +-----------------------+
|    Init Plugin   |                   |                       |
|                  +--------+          |      Agent Plugin     |
+------------------+        |          |                       |
                            |          +-----------------------+
                            +--------->| Plugin global var     |
                                       |   + Init() error      |
                                       |   + AfterInit() error |
                                       |   + Close() error     |
                                       +-----------------------+
```
