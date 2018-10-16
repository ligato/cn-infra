# Filesystem reader

Reader exposes API to work with OS file system, like file path verification, transforming configuration from files
or validation.

Package `fsnotify` is used to obtain events. File validator currently checks following:

* ignores empty events (path is an empty string)
* ignores older revisions (backups) of given file (with '~' at the end of the file name)
* ignores all temporary files with extension '.sw*'
* ignores full-numeric file names (created by some editors where the file is opened)

