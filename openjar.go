package openjar

import (
	log "bitbucket.org/yanghu/logger"
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

type CookiesMap map[string][]*http.Cookie

// OpenJar implements http.CookieJar interface. It's a simple wrap
// on cookiejar.Jar, store a copy of cookies in a map so we can easily
// serialize cookies.
type OpenJar struct {
	*cookiejar.Jar
	Store CookiesMap
}

func New() *OpenJar {
	jar, _ := cookiejar.New(nil)
	return &OpenJar{Jar: jar, Store: make(CookiesMap)}
}

func (jar *OpenJar) Cookies(u *url.URL) (cookies []*http.Cookie) {
	return jar.Jar.Cookies(u)
}

// SetCookies stores a copy of the cookies in map. the keys are
// scheme|host
func (jar *OpenJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.Jar.SetCookies(u, cookies)
	// also store cookie in store
	jar.UpdateStore(u)
}

// StoreCookies saves cookies in map. map key is deduced from url
func (jar *OpenJar) UpdateStore(u *url.URL) {
	key := jar.urlKey(u)
	jar.Store[key] = jar.Jar.Cookies(u)
}

// urlKey creates key from an url.URL for map storate.
// only scheme and host will be used in CookieJar's SetCookies method.
func (jar *OpenJar) urlKey(u *url.URL) string {
	return fmt.Sprintf("%s|%s", u.Scheme, u.Host)
}

// urlFromKey decode keys to *url.URL so that it can be used for http.CookieJar
func (jar *OpenJar) urlFromKey(key string) *url.URL {
	fields := strings.Split(key, "|")
	return &url.URL{Scheme: fields[0], Host: fields[1]}
}

// FillJar puts all cookies from map into the internal Jar.
// urls are created from keys
func (jar *OpenJar) FillJar() {
	var u *url.URL
	for key, cookies := range jar.Store {
		u = jar.urlFromKey(key)
		jar.SetCookies(u, cookies)
	}
}

func (jar *OpenJar) MarshalBinary() ([]byte, error) {
	// a simple encoding: calling gob encoder
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err := enc.Encode(jar.Store)
	return b.Bytes(), err
}

func (jar *OpenJar) UnmarshalBinary(data []byte) error {
	b := bytes.NewBuffer(data)
	dec := gob.NewDecoder(b)
	err := dec.Decode(&jar.Store)
	jar.FillJar()
	return err
}

// Encode and decode only operates Store, which is a map
func (jar *OpenJar) Encode(w io.Writer) error {
	enc := gob.NewEncoder(w)
	return enc.Encode(jar.Store)
}

func (jar *OpenJar) Decode(r io.Reader) error {
	dec := gob.NewDecoder(r)
	err := dec.Decode(&jar.Store)
	if err != nil {
		log.Error(err, "")
		return err
	}
	jar.FillJar() // update cookiejar after decoding.
	return nil
}

func (jar *OpenJar) String() string {
	var (
		u         *url.URL
		buf       bytes.Buffer
		cookieCnt int
	)
	for key, cookies := range jar.Store {
		u = jar.urlFromKey(key)
		cookieCnt = len(cookies)
		fmt.Fprintf(&buf, "Cookies for %s\n", u.Host)
		fmt.Fprintf(&buf, "Cookie count: %d\n", cookieCnt)
		for i, cookie := range cookies {
			fmt.Fprintf(&buf, "--------cookie [%d] --------\n", i)
			fmt.Fprintf(&buf, "Name\t= %s\n", cookie.Name)
			fmt.Fprintf(&buf, "Value\t= %s\n", cookie.Value)
			fmt.Fprintf(&buf, "Path\t= %s\n", cookie.Path)
			fmt.Fprintf(&buf, "Domain\t= %s\n", cookie.Domain)
			fmt.Fprintf(&buf, "Expires\t= %s\n", cookie.Expires)
			fmt.Fprintf(&buf, "RawExpires\t= %s\n", cookie.RawExpires)
		}
	}
	return buf.String()
}
