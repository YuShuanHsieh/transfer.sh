package server

import (
	"net"
	"time"
)

type AuthType string

const (
	IP       AuthType = "IP"
	METADATA AuthType = "METADATA"
	API      AuthType = "API"
)

type Metadata struct {
	// ContentType is the original uploading content type
	ContentType string
	// Secret as knowledge to delete file
	// Secret string
	// Downloads is the actual number of downloads
	Downloads int
	// MaxDownloads contains the maximum numbers of downloads
	MaxDownloads int
	// MaxDate contains the max age of the file
	MaxDate time.Time
	// DeletionToken contains the token to match against for deletion
	DeletionToken string

	AuthTypes []AuthType
	// Basic Auth for downloading
	User string
	// Basic Auth for downloading
	Password string
	// IP filter
	IP []net.IP
	// Network filter
	Nets []*net.IPNet
}

func (m *Metadata) Authenticate(user, password string) (bool, error) {
	if user == m.User && m.Password == password {
		return true, nil
	}
	return false, nil
}

func (m *Metadata) AuthRequired() bool {
	return len(m.AuthTypes) != 0
}

func (m *Metadata) IPFilterEnabled() bool {
	for _, v := range m.AuthTypes {
		if v == IP {
			return true
		}
	}
	return false
}

func (m *Metadata) AllowedIP(remoteAddr string) bool {
	if !m.IPFilterEnabled() {
		return true
	}
	host, _, err := net.SplitHostPort(remoteAddr)
	remoteIP := net.ParseIP(host)
	if err != nil || remoteIP == nil {
		return false
	}
	for _, ip := range m.IP {
		if ip.Equal(remoteIP) {
			return true
		}
	}
	for _, net := range m.Nets {
		if net.Contains(remoteIP) {
			return true
		}
	}
	return false
}
