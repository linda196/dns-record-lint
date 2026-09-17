# dns-record-lint

A Go library that reads DNS zone file records and points out the mistakes
that resolvers won't tell you about until something breaks: a CNAME sharing
a name with other records, a TTL nobody ever set, and (eventually) the
usual list of zone file footguns.

There's no CLI here on purpose. This is meant to be embedded — in a
provisioning tool that generates zone files, in a CI check that runs
against a `terraform plan` for a DNS zone, in whatever you're already
building. It gives you records in, a report out, and lets you decide how
to surface it.

## Usage

```go
package main

import (
	"fmt"

	"github.com/linda196/dns-record-lint"
)

func main() {
	lines := []string{
		"www.example.com.  3600  IN  CNAME  example.com.",
		"www.example.com.        IN  A      192.0.2.10",
		"mail.example.com. 300   IN  MX     10 mail.example.com.",
	}

	var records []dnslint.Record
	for _, line := range lines {
		rec, err := dnslint.ParseRecord(line)
		if err != nil {
			panic(err)
		}
		records = append(records, rec)
	}

	report := dnslint.Lint(records, dnslint.DefaultRules()...)
	fmt.Println(report.Text())
}
```

Output:

```
[ERROR] www.example.com A: name has a CNAME plus other records, which RFC 1034 forbids
[ERROR] www.example.com CNAME: name has a CNAME plus other records, which RFC 1034 forbids
[INFO] mail.example.com MX: no explicit TTL; value comes from zone or SOA default
```

## The `--json` output mode

`Report` renders itself two ways: `Text()` for a terminal and `JSON()` for
anything that needs to parse the result. `Format(asJSON bool)` picks
between them, which is exactly the shape a CLI wrapper's `--json` flag
needs:

```go
var asJSON = flag.Bool("json", false, "emit the report as JSON")

func main() {
	flag.Parse()
	// ... build records, run dnslint.Lint ...
	out, err := report.Format(*asJSON)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(out)
}
```

With `--json` the same report above comes out as:

```json
{
  "issues": [
    {
      "name": "www.example.com",
      "type": "A",
      "severity": "error",
      "message": "name has a CNAME plus other records, which RFC 1034 forbids"
    }
  ]
}
```

## Status

Early. `ParseRecord` handles single-line records for the common types (A,
AAAA, CNAME, MX, NS, TXT, PTR, SRV, CAA) but not multi-line SOA blocks or
zone file directives like `$ORIGIN` and `$TTL` yet — see the roadmap below.

## License

MIT, see [LICENSE](LICENSE).
