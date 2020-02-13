Nomad Skeleton Device Plugin
==================

Skeleton project for [Nomad device plugins](https://www.nomadproject.io/docs/internals/plugins/devices.html).

This project is intended for bootstrapping development of a new device plugin.

- Website: https://www.nomadproject.io
- Mailing list: [Google Groups](http://groups.google.com/group/nomad-tool)

Requirements
------------

- [Nomad](https://www.nomadproject.io/downloads.html) 0.9+
- [Go](https://golang.org/doc/install) 1.11 or later (to build the plugin)

Building the Skeleton Plugin
---------------------
[Generate](https://github.com/hashicorp/nomad-skeleton-device-plugin/generate)
a new repository in your account from this template by clicking the `Use this
template` button above.

Clone the repository somewhere in your computer. This project uses
[Go modules](https://blog.golang.org/using-go-modules) so you will need to set
the environment variable `GO111MODULE=on` or work outside your `GOPATH` if it
is set to `auto` or not declared.

```sh
$ git clone git@github.com:<ORG>/<REPO>git
```

Enter the plugin directory and update the paths in `go.mod` and `main.go` to
match your repository path.

```diff
// go.mod

- module github.com/hashicorp/nomad-skeleton-device-plugin
+ module github.com/<ORG>/<REPO>
...
```

```diff
// main.go

package main

import (
    log "github.com/hashicorp/go-hclog"
    "github.com/hashicorp/nomad/plugins"

-   "github.com/hashicorp/nomad-skeleton-device-plugin/device"
+   "github.com/<REPO>/<ORG>/device"
)
...
```

Build the skeleton plugin.

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
