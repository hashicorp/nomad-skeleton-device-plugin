Nomad Skeleton Device Plugin
==================

Skeleton project for [Nomad device plugins](https://www.nomadproject.io/docs/internals/plugins/devices.html).

This project is intended for bootstrapping development of a new device plugin.

- Website: https://www.nomadproject.io
- Mailing list: [Google Groups](http://groups.google.com/group/nomad-tool)

Requirements
------------

- [Nomad](https://www.nomadproject.io/downloads.html) 0.9+
- [Go](https://golang.org/doc/install) 1.11 or later (to build the provider plugin)

Building the Skeleton Plugin
---------------------

Clone the repository. This project uses [go modules](https://github.com/golang/go/wiki/Modules); you will need to
set `GO111MODULE=on` (or `auto`, depending on your Go version and whether you are working inside your GOPATH).

```sh
$ git clone git@github.com:hashicorp/nomad-skeleton-device-plugin.git
```

Enter the provider directory and build the skeleton provider:

```sh
$ make build
```

Running the Plugin in Development
---------------------

You can test this plugin (and your own device plugins) in development using the
[plugin launcher](https://github.com/hashicorp/nomad/tree/master/plugins/shared/cmd/launcher). The makefile provides
a target for this:

```sh
$ make eval
```

Deploying Device Plugins in Nomad
----------------------

Copy the plugin binary to the
[plugins directory](https://www.nomadproject.io/docs/configuration/index.html#plugin_dir) and
[configure the plugin](https://www.nomadproject.io/docs/configuration/plugin.html) in the client config. Then use the
[device stanza](https://www.nomadproject.io/docs/job-specification/device.html) in the job file to schedule with
device support. (Note, the skeleton plugin is not intended for use in Nomad.)
