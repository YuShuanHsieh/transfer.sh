package server

import (
	"net"
	"time"
)

type AuthType string

const (
	IP       AuthType = "IP"
	METADATA AuthType = "METADATA"
	ACCOUNT  AuthType = "ACCOUNT"
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
	if user == m.User && password != password {
		return true, nil
	}
	return false, nil
}
