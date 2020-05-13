package server

import (
	"crypto/tls"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/acme/autocert"

	context "golang.org/x/net/context"
)

type OptionFn func(*Server)

func ClamavHost(s string) OptionFn {
	return func(srvr *Server) {
		srvr.ClamAVDaemonHost = s
	}
}

func VirustotalKey(s string) OptionFn {
	return func(srvr *Server) {
		srvr.VirusTotalKey = s
	}
}

func Listener(s string) OptionFn {
	return func(srvr *Server) {
		srvr.ListenerString = s
	}

}

func CorsDomains(s string) OptionFn {
	return func(srvr *Server) {
		srvr.CorsDomains = s
	}

}

func GoogleAnalytics(gaKey string) OptionFn {
	return func(srvr *Server) {
		srvr.gaKey = gaKey
	}
}

func UserVoice(userVoiceKey string) OptionFn {
	return func(srvr *Server) {
		srvr.userVoiceKey = userVoiceKey
	}
}

func TLSListener(s string, t bool) OptionFn {
	return func(srvr *Server) {
		srvr.TLSListenerString = s
		srvr.TLSListenerOnly = t
	}

}

func ProfileListener(s string) OptionFn {
	return func(srvr *Server) {
		srvr.ProfileListenerString = s
	}
}

func WebPath(s string) OptionFn {
	return func(srvr *Server) {
		if s[len(s)-1:] != "/" {
			s = s + string(filepath.Separator)
		}

		srvr.webPath = s
	}
}

func ProxyPath(s string) OptionFn {
	return func(srvr *Server) {
		if s[len(s)-1:] != "/" {
			s = s + string(filepath.Separator)
		}

		srvr.proxyPath = s
	}
}

func TempPath(s string) OptionFn {
	return func(srvr *Server) {
		if s[len(s)-1:] != "/" {
			s = s + string(filepath.Separator)
		}

		srvr.tempPath = s
	}
}

func LogFile(logger *log.Logger, s string) OptionFn {
	return func(srvr *Server) {
		f, err := os.OpenFile(s, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}

		logger.SetOutput(f)
		srvr.logger = logger
	}
}

func Logger(logger *log.Logger) OptionFn {
	return func(srvr *Server) {
		srvr.logger = logger
	}
}

func RateLimit(requests int) OptionFn {
	return func(srvr *Server) {
		srvr.rateLimitRequests = requests
	}
}

func ForceHTTPs() OptionFn {
	return func(srvr *Server) {
		srvr.forceHTTPs = true
	}
}

func EnableProfiler() OptionFn {
	return func(srvr *Server) {
		srvr.profilerEnabled = true
	}
}

func UseMetaStorage(s Storage) OptionFn {
	return func(srvr *Server) {
		srvr.metadataStorage = s
	}
}

func UseStorage(s Storage) OptionFn {
	return func(srvr *Server) {
		srvr.storage = s
	}
}

func UseLetsEncrypt(hosts []string) OptionFn {
	return func(srvr *Server) {
		cacheDir := "./cache/"

		m := autocert.Manager{
			Prompt: autocert.AcceptTOS,
			Cache:  autocert.DirCache(cacheDir),
			HostPolicy: func(_ context.Context, host string) error {
				found := false

				for _, h := range hosts {
					found = found || strings.HasSuffix(host, h)
				}

				if !found {
					return errors.New("acme/autocert: host not configured")
				}

				return nil
			},
		}

		srvr.tlsConfig = &tls.Config{
			GetCertificate: m.GetCertificate,
		}
	}
}

func TLSConfig(cert, pk string) OptionFn {
	certificate, err := tls.LoadX509KeyPair(cert, pk)
	return func(srvr *Server) {
		srvr.tlsConfig = &tls.Config{
			GetCertificate: func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
				return &certificate, err
			},
		}
	}
}

func AuthCredential(key string, auth Authenticator) OptionFn {
	return func(srvr *Server) {
		srvr.auths[key] = auth
	}
}

func FilterOptions(options IPFilterOptions) OptionFn {
	for i, allowedIP := range options.AllowedIPs {
		options.AllowedIPs[i] = strings.TrimSpace(allowedIP)
	}

	for i, blockedIP := range options.BlockedIPs {
		options.BlockedIPs[i] = strings.TrimSpace(blockedIP)
	}

	return func(srvr *Server) {
		srvr.ipFilterOptions = &options
	}
}
