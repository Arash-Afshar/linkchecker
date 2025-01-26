# Link Checker

A Go tool that validates links in PDF documents.

Checks if links in documents are:
- Live (accessible via HTTP/HTTPS)
- Valid (match whitelist/blacklist rules)

## Usage

- You can provide both `--pdfs` and `--directory` but at least one of them must be provided
- You can provide either `--whitelist-domains` or `--blacklist-domains` or neither but not both
- The liveness check is done by sending a GET request and checking if the response is 200 OK

### As a Library
```go
import "github.com/Arash-Afshar/linkchecker"
results, err := linkchecker.CheckLinks(config)
```

where the config is a `linkchecker.Config` struct:

```go
type Config struct {
	Pdfs             []string `help:"PDF files to process"`
	Directory        string   `help:"Directory containing PDF files to process"`
	WhitelistDomains []string `help:"List of allowed domains"`
	BlacklistDomains []string `help:"List of blocked domains"`
}
```

### As a CLI

```bash
make build
./linkchecker --pdfs test-data/test.pdf --whitelist-domains a.com,b.com
```


### Tests

To run the tests you need to install latex so that the test pdf can be built. You do NOT need to install latex to use the library.

```bash
make test-deps
make test
```
