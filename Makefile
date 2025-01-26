
.PHONY: test
test:
	rm -rf "test-data/*.{pdf,aux,log,out}"
	cd test-data && pdflatex test.tex && pdflatex test.tex && cd -
	go test -v ./...

.PHONY: build
build:
	go build -o linkchecker cmd/main.go

.PHONY: test-deps
test-deps:
	apt-get install texlive-latex-recommended
