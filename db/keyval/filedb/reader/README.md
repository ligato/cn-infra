# Filesystem reader

Reader exposes API to work with OS file system, like file path verification, transforming configuration from files,
validation and file system notification event watching.

Package `fsnotify` is used to obtain events. The file validator currently supports only files with `.json` or `.yaml`
extensions. Files with proper data, but without extension are ignored.