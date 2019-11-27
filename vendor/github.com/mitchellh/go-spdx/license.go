package spdx

// LicenseList is the structure for a list of licenses provided by the SPDX API
type LicenseList struct {
	// Version is the raw version string of the license list.
	Version string `json:"licenseListVersion"`

	// Licenses is the list of known licenses.
	Licenses []*LicenseInfo `json:"licenses"`
}

// LicenseInfo is a single software license.
//
// Basic descriptions are documented in the fields below. For a full
// description of the fields, see the official SPDX specification here:
// https://github.com/spdx/license-list-data/blob/master/accessingLicenses.md
type LicenseInfo struct {
	ID          string   `json:"licenseId"`
	Name        string   `json:"name"`
	Text        string   `json:"licenseText"`
	Deprecated  bool     `json:"isDeprecatedLicenseId"`
	OSIApproved bool     `json:"isOsiApproved"`
	SeeAlso     []string `json:"seeAlso"`
}

// License looks up the license in the list with the given ID. If the license
// is not found, nil is returned.
//
// Note that licenses in a LicenseList are usually missing fields such as Text.
// To fully populate a Licenese, call Client.Licence with the ID.
func (l *LicenseList) License(id string) *LicenseInfo {
	for _, v := range l.Licenses {
		if v != nil && v.ID == id {
			return v
		}
	}

	return nil
}
