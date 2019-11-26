package spdx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/hashicorp/go-cleanhttp"
)

const (
	DefaultListURL    = "https://spdx.org/licenses/licenses.json"
	DefaultDetailsURL = "https://spdx.org/licenses/%[1]s.json"
)

// Client is an API client for accessing SPDX data.
//
// Configure any fields on the struct prior to calling any functions. After
// calling functions, do not access the fields again.
//
// The functions on Client are safe to call concurrently.
type Client struct {
	// HTTP is the HTTP client to use for requests. If this is nil, then
	// a default new HTTP client will be used.
	HTTP *http.Client

	// ListURL and DetailsURL are the URLs for listing licenses and accessing
	// a single license, respectively. If these are not set, they will default
	// to the default values specified in constants (i.e. DefaultListURL).
	//
	// For DetailsURL, use the placeholder "%[1]s" to interpolate the SPDX ID.
	ListURL    string
	DetailsURL string

	once sync.Once
}

// List returns the list of licenses.
//
// If err == nil, then *LicenseList will always be non-nil.
func (c *Client) List() (*LicenseList, error) {
	c.once.Do(c.init)

	resp, err := c.HTTP.Get(c.ListURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result LicenseList
	return &result, json.NewDecoder(resp.Body).Decode(&result)
}

// License returns the license by ID. This often includes more detailed
// information than List such as the full license text.
//
// The ID is usually case sensitive. Please ensure the ID is set exactly
// to the SPDX ID, including casing.
//
// If err == nil, then *License will always be non-nil.
func (c *Client) License(id string) (*LicenseInfo, error) {
	c.once.Do(c.init)

	resp, err := c.HTTP.Get(fmt.Sprintf(c.DetailsURL, id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result LicenseInfo
	return &result, json.NewDecoder(resp.Body).Decode(&result)
}

func (c *Client) init() {
	if c.HTTP == nil {
		c.HTTP = cleanhttp.DefaultClient()
	}

	if c.ListURL == "" {
		c.ListURL = DefaultListURL
	}
	if c.DetailsURL == "" {
		c.DetailsURL = DefaultDetailsURL
	}
}
