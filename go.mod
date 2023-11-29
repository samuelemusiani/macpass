module github.com/musianisamuele/macpass

go 1.21.3

require (
	github.com/coreos/go-iptables v0.7.0
	github.com/ghodss/yaml v1.0.0
	github.com/jcmturner/gokrb5/v8 v8.4.4
	github.com/mattn/go-sqlite3 v1.14.17
	golang.org/x/term v0.13.0
	gotest.tools/v3 v3.5.1
	internal/comunication v1.0.0
)

replace internal/comunication => ./internal/comunication

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	golang.org/x/crypto v0.13.0 // indirect
	golang.org/x/net v0.15.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
)
