package spdx

// DefaultClient is the default Client used by the top-level functions
// List, License, etc.
var DefaultClient = &Client{}

// List is the same as Client.List, but operates on DefaultClient.
func List() (*LicenseList, error) { return DefaultClient.List() }

// License is the same as Client.License, but operates on DefaultClient.
func License(id string) (*LicenseInfo, error) { return DefaultClient.License(id) }
