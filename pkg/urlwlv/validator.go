package urlwlv

import (
	"errors"
	"net/url"
	"path"
	"strings"
)

var (
	ErrInvalidURL       = errors.New("invalid URL")
	ErrInvalidProtocol  = errors.New("protocol not allowed")
	ErrInvalidDomain    = errors.New("domain not allowed")
	ErrInvalidExtension = errors.New("extension not allowed")
)

type Validator struct {
	protocols  map[string]struct{}
	domains    map[string]struct{}
	extensions map[string]struct{}
}

func NewValidator(protocols, domains, extensions []string) *Validator {
	v := &Validator{
		protocols:  make(map[string]struct{}),
		domains:    make(map[string]struct{}),
		extensions: make(map[string]struct{}),
	}
	for _, p := range protocols {
		p = strings.ToLower(strings.TrimSpace(p))
		if p != "" {
			v.protocols[p] = struct{}{}
		}
	}
	for _, d := range domains {
		d = normalizeDomain(d)
		if d != "" {
			v.domains[d] = struct{}{}
		}
	}
	for _, ext := range extensions {
		ext = strings.ToLower(strings.TrimSpace(ext))
		if ext != "" && !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		v.extensions[ext] = struct{}{}
	}

	return v
}

func (v *Validator) Validate(urlStr string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ErrInvalidURL
	}

	if len(v.protocols) > 0 {
		scheme := strings.ToLower(u.Scheme)
		if _, ok := v.protocols[scheme]; !ok {
			return ErrInvalidProtocol
		}
	}

	if len(v.domains) > 0 {
		host := normalizeDomain(u.Hostname())
		if _, ok := v.domains[host]; !ok {
			return ErrInvalidDomain
		}
	}

	if len(v.extensions) > 0 {
		ext := strings.ToLower(path.Ext(u.Path))
		if _, ok := v.extensions[ext]; !ok {
			return ErrInvalidExtension
		}
	}

	return nil
}

func normalizeDomain(domain string) string {
	domain = strings.ToLower(strings.TrimSpace(domain))
	if strings.HasPrefix(domain, "www.") {
		return domain[4:]
	}
	return domain
}
