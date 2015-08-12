package openjar

import (
	"bytes"
	"errors"
	// "fmt"
	"github.com/stretchr/testify/assert"
	// "net/http"
	// "net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"testing"
)

var CookieGobFile = "../testData/ToBeReadCookies.gob"
var sampleURL = &url.URL{Scheme: "http", Host: "www.google.com"}

// sampleCookies loads some []*http.Cookies from a gob encoded file
func populateJar(f string, jar *OpenJar) error {
	file, err := os.Open(CookieGobFile)
	if err != nil {
		return err
	}
	cookies := ReadCookies(file)
	file.Close()
	if len(cookies) != 2 {
		return errors.New("read cookie wrong")
	}
	jar.SetCookies(sampleURL, cookies)
	return nil
}

func TestOpenJar(t *testing.T) {
	jar := New()
	// test SetCookies and Cookies of jar
	populateJar(CookieGobFile, jar)
	cookies := jar.Cookies(sampleURL)
	assert.True(t, len(cookies) == 2, "cookie length is %d", len(cookies))
	assert.True(t, strings.Contains(cookies[0].Name, "JSESSIONID"), "cookie should be equal")
	// check if Store is also populated
	assert.True(t, len(jar.Store) == 1, "Store should have cookies")

	// test FillJar. Populate cookiejar from Store map
	jar2 := New()
	//copy map from jar1 to jar2
	for k, v := range jar.Store {
		jar2.Store[k] = v
	}
	jar2.FillJar()
	cookies = jar2.Cookies(sampleURL)
	assert.True(t, len(cookies) == 2, "jar2 cookie length is %d", len(cookies))
	assert.True(t, strings.Contains(cookies[0].Name, "JSESSIONID"), "cookie should be equal")
}

func TestSerialize(t *testing.T) {
	jar := New()
	populateJar(CookieGobFile, jar)
	// test encoding and decoding
	var buf bytes.Buffer
	jar.Encode(&buf)
	jar2 := New()
	jar2.Decode(&buf)

	cookies := jar2.Cookies(sampleURL)
	assert.True(t, len(cookies) == 2, "jar2 cookie length is %d", len(cookies))
	assert.True(t, strings.Contains(cookies[0].Name, "JSESSIONID"), "cookie should be equal")
}
