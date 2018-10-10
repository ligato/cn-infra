# Filesystem configuration reader plugin

The filesystem plugin allows to use the file system of a operating system as a key-value data store. The filesystem
plugin watches for pre-defined files or directories, reads a configuration and sends response events according
to changes.

All the configuration is resynced in the beginning (as for standard key-value data store). Configuration files
then can be added, updated, moved, renamed or removed, plugin makes all the necessary changes.

## Configuration

All files/directories used as a data store must be defined in configuration file. Location of the file
can be defined either by the command line flag `filesystem-config` or set via the `FILESYSTEM_CONFIG`
environment variable.

## Data structure

Plugin currently supports only json-formatted data. The format of the file is as follows:

```
{
    "data": [
        {
            "key": "<key>",
            "value": {
                <proto-modelled data>
            }
        },
        {
            "key": "<key>",
            "value": {
                <proto-modelled data>
            }
        },
        ...
    ]
}

``` 

Key has to contain also instance prefix with micro service label, so plugin knows which parts of the configuration 
are intended for it. All configuration is stored internally in local database. It allows to compare events and 
respond with correct 'previous' value for a given key. 

