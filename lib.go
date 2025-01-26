package linkchecker

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

type Config struct {
	Pdfs             []string `help:"PDF files to process"`
	Directory        string   `help:"Directory containing PDF files to process"`
	WhitelistDomains []string `help:"List of allowed domains"`
	BlacklistDomains []string `help:"List of blocked domains"`
}

type Link struct {
	URL     string
	IsValid bool
	IsLive  bool
}

func (l *Link) String() string {
	validMark := "\033[31m✗\033[0m"
	if l.IsValid {
		validMark = "\033[32m✓\033[0m"
	}
	liveMark := "\033[31m✗\033[0m"
	if l.IsLive {
		liveMark = "\033[32m✓\033[0m"
	}
	linkColored := "\033[31m" + l.URL + "\033[0m"
	if l.IsValid && l.IsLive {
		linkColored = "\033[32m" + l.URL + "\033[0m"
	}
	return fmt.Sprintf("%s %s %s", validMark, liveMark, linkColored)
}

func ValidateConfig(config *Config) error {
	if len(config.Pdfs) == 0 && config.Directory == "" {
		return errors.New("provide at least one PDF file or a directory containing PDF files")
	}

	pdfFiles := config.Pdfs
	if config.Directory != "" {
		pdfFiles, err := getPDFsFromDirectory(config.Directory)
		if err != nil {
			return err
		}
		config.Pdfs = append(config.Pdfs, pdfFiles...)
	}

	if len(pdfFiles) == 0 {
		return errors.New("no PDF files found")
	}

	return nil
}

func getPDFsFromDirectory(directory string) ([]string, error) {
	pdfFiles, err := filepath.Glob(filepath.Join(directory, "*.pdf"))
	if err != nil {
		return nil, fmt.Errorf("getting PDFs from directory %s: %v", directory, err)
	}
	return pdfFiles, nil
}

func extractsLinks(pdfFiles []string) ([]Link, error) {
	conf := api.LoadConfiguration()
	allLinks := []Link{}

	for _, pdfFile := range pdfFiles {
		f, err := os.Open(pdfFile)
		if err != nil {
			return nil, fmt.Errorf("opening PDF file %s: %v", pdfFile, err)
		}
		defer f.Close()
		links, err := extractLink(f, conf)
		if err != nil {
			return nil, fmt.Errorf("extracting links from PDF file %s: %v", pdfFile, err)
		}
		allLinks = append(allLinks, links...)
	}
	return allLinks, nil
}

func extractLink(f io.ReadSeeker, conf *model.Configuration) ([]Link, error) {
	links := []Link{}
	annots, err := api.Annotations(f, nil, conf)
	if err != nil {
		return nil, fmt.Errorf("extracting annotations: %v", err)
	}

	for _, pageAnnot := range annots {
		for annotType, annot := range pageAnnot {
			if annotType == model.AnnLink {
				for _, link := range annot.Map {
					link, ok := link.(model.LinkAnnotation)
					if !ok {
						continue
					}
					if link.URI != "" {
						links = append(links, Link{
							URL:     link.URI,
							IsValid: false,
						})
					}
				}
			}
		}
	}

	return links, nil
}

func validateLinks(links []Link, whitelistDomains []string, blacklistDomains []string) {
	if len(whitelistDomains) > 0 && len(blacklistDomains) > 0 {
		for i := range links {
			links[i].IsValid = false
		}
		return
	}

	if len(whitelistDomains) == 0 && len(blacklistDomains) == 0 {
		for i := range links {
			links[i].IsValid = true
		}
		return
	}

	for i := range links {
		l, err := url.Parse(links[i].URL)
		if err != nil {
			links[i].IsValid = false
			continue
		}
		host := l.Host
		if host == "" {
			host = links[i].URL
		}

		defaultIsValid := true
		defaultList := blacklistDomains
		if len(whitelistDomains) > 0 {
			defaultIsValid = false
			defaultList = whitelistDomains
		}
		for _, domain := range defaultList {
			links[i].IsValid = defaultIsValid
			if strings.HasSuffix(host, domain) {
				links[i].IsValid = !defaultIsValid
				break
			}
		}
	}
}

func livenesscheck(links []Link) {
	for i := range links {
		resp, err := http.Get(links[i].URL)
		if err != nil {
			links[i].IsLive = false
			continue
		}
		links[i].IsLive = resp.StatusCode == http.StatusOK
	}
}

func CheckLinks(config *Config) ([]Link, error) {
	links, err := extractsLinks(config.Pdfs)
	if err != nil {
		return nil, fmt.Errorf("error extracting links: %v", err)
	}

	validateLinks(links, config.WhitelistDomains, config.BlacklistDomains)
	livenesscheck(links)
	return links, nil
}
