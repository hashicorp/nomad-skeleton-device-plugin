module github.com/hashicorp/nomad-skeleton-device-plugin

go 1.12

require (
	github.com/LK4D4/joincontext v0.0.0-20171026170139-1724345da6d5 // indirect
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/gorhill/cronexpr v0.0.0-20180427100037-88b0669f7d75 // indirect
	github.com/hashicorp/consul/api v1.2.0 // indirect
	github.com/hashicorp/go-hclog v0.9.1
	github.com/hashicorp/go-immutable-radix v1.1.0 // indirect
	github.com/hashicorp/go-version v1.2.0 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/hashicorp/hcl2 v0.0.0-20191002203319-fb75b3253c80 // indirect
	github.com/hashicorp/nomad v0.10.0-beta1.0.20191119152219-a9490506dc2a
	github.com/hashicorp/raft v1.1.1 // indirect
	github.com/hashicorp/vault/api v1.0.4 // indirect
	github.com/kr/pretty v0.1.0
	github.com/mitchellh/hashstructure v1.0.0 // indirect
	github.com/shirou/gopsutil v0.0.0-00010101000000-000000000000 // indirect
	github.com/shirou/w32 v0.0.0-20160930032740-bb4de0191aa4 // indirect
	github.com/ugorji/go v0.0.0-00010101000000-000000000000 // indirect
	github.com/zclconf/go-cty v1.1.0 // indirect
	golang.org/x/crypto v0.0.0-20191029031824-8986dd9e96cf // indirect
	google.golang.org/grpc v1.22.0
)

// don't use shirou/gopsutil, use the hashicorp fork
replace github.com/shirou/gopsutil => github.com/hashicorp/gopsutil v0.0.0-20180427102116-62d5761ddb7d

// don't use ugorji/go, use the hashicorp fork
replace github.com/ugorji/go => github.com/hashicorp/go-msgpack v0.0.0-20190927083313-23165f7bc3c2

// fix the version of hashicorp/go-msgpack to 96ddbed8d05b
replace github.com/hashicorp/go-msgpack => github.com/hashicorp/go-msgpack v0.0.0-20191101193846-96ddbed8d05b
