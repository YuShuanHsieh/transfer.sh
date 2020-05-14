package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	apiauth "github.com/dutchcoders/transfer.sh/api-auth"

	redisStorage "github.com/dutchcoders/transfer.sh/redis-storage"

	"github.com/dutchcoders/transfer.sh/server"
	"github.com/fatih/color"
	"github.com/urfave/cli"
	"google.golang.org/api/googleapi"
)

var Version = "1.1.7"
var helpTemplate = `NAME:
{{.Name}} - {{.Usage}}

DESCRIPTION:
{{.Description}}

USAGE:
{{.Name}} {{if .Flags}}[flags] {{end}}command{{if .Flags}}{{end}} [arguments...]

COMMANDS:
{{range .Commands}}{{join .Names ", "}}{{ "\t" }}{{.Usage}}
{{end}}{{if .Flags}}
FLAGS:
{{range .Flags}}{{.}}
{{end}}{{end}}
VERSION:
` + Version +
	`{{ "\n"}}`

var globalFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "listener",
		Usage: "127.0.0.1:8080",
		Value: "127.0.0.1:8080",
		EnvVar: "LISTENER",
	},
	// redirect to https?
	// hostnames
	cli.StringFlag{
		Name:  "profile-listener",
		Usage: "127.0.0.1:6060",
		Value: "",
		EnvVar: "PROFILE_LISTENER",
	},
	cli.BoolFlag{
		Name:  "force-https",
		Usage: "",
		EnvVar: "FORCE_HTTPS",
	},
	cli.StringFlag{
		Name:  "tls-listener",
		Usage: "127.0.0.1:8443",
		Value: "",
		EnvVar: "TLS_LISTENER",
	},
	cli.BoolFlag{
		Name:  "tls-listener-only",
		Usage: "",
		EnvVar: "TLS_LISTENER_ONLY",
	},
	cli.StringFlag{
		Name:  "tls-cert-file",
		Value: "",
		EnvVar: "TLS_CERT_FILE",
	},
	cli.StringFlag{
		Name:  "tls-private-key",
		Value: "",
		EnvVar: "TLS_PRIVATE_KEY",
	},
	cli.StringFlag{
		Name:  "temp-path",
		Usage: "path to temp files",
		Value: os.TempDir(),
		EnvVar: "TEMP_PATH",
	},
	cli.StringFlag{
		Name:  "web-path",
		Usage: "path to static web files",
		Value: "",
		EnvVar: "WEB_PATH",
	},
	cli.StringFlag{
		Name:  "proxy-path",
		Usage: "path prefix when service is run behind a proxy",
		Value: "",
		EnvVar: "PROXY_PATH",
	},
	cli.StringFlag{
		Name:  "proxy-port",
		Usage: "port of the proxy when the service is run behind a proxy",
		Value: "",
		EnvVar: "PROXY_PORT",
	},
	cli.StringFlag{
		Name:  "ga-key",
		Usage: "key for google analytics (front end)",
		Value: "",
		EnvVar: "GA_KEY",
	},
	cli.StringFlag{
		Name:  "uservoice-key",
		Usage: "key for user voice (front end)",
		Value: "",
		EnvVar: "USERVOICE_KEY",
	},
	cli.StringFlag{
		Name:  "api-endpoint",
		Usage: "the endpoint for the api authenticator",
		Value: "",
	},
	cli.StringFlag{
		Name:  "api-headers",
		Usage: "additional headers for the api authenticator",
		Value: "",
	},
	cli.StringFlag{
		Name:  "provider",
		Usage: "s3|gdrive|local",
		Value: "",
		EnvVar: "PROVIDER",
	},
	cli.StringFlag{
		Name:  "meta-provider",
		Usage: "s3|gdrive|local|redis",
		Value: "",
	},
	cli.StringFlag{
		Name:  "redis-addr",
		Usage: "The address of redis server",
		Value: "",
	},
	cli.StringFlag{
		Name:  "redis-pwd",
		Usage: "The password of redis server",
		Value: "",
	},
	cli.StringFlag{
		Name:   "s3-endpoint",
		Usage:  "",
		Value:  "",
		EnvVar: "S3_ENDPOINT",
	},
	cli.StringFlag{
		Name:   "s3-region",
		Usage:  "",
		Value:  "eu-west-1",
		EnvVar: "S3_REGION",
	},
	cli.StringFlag{
		Name:   "aws-access-key",
		Usage:  "",
		Value:  "",
		EnvVar: "AWS_ACCESS_KEY",
	},
	cli.StringFlag{
		Name:   "aws-secret-key",
		Usage:  "",
		Value:  "",
		EnvVar: "AWS_SECRET_KEY",
	},
	cli.StringFlag{
		Name:   "bucket",
		Usage:  "",
		Value:  "",
		EnvVar: "BUCKET",
	},
	cli.BoolFlag{
		Name:  "s3-no-multipart",
		Usage: "Disables S3 Multipart Puts",
		EnvVar: "S3_NO_MULTIPART",
	},
	cli.BoolFlag{
		Name:  "s3-path-style",
		Usage: "Forces path style URLs, required for Minio.",
		EnvVar: "S3_PATH_STYLE",
	},
	cli.StringFlag{
		Name:  "gdrive-client-json-filepath",
		Usage: "",
		Value: "",
		EnvVar: "GDRIVE_CLIENT_JSON_FILEPATH",
	},
	cli.StringFlag{
		Name:  "gdrive-local-config-path",
		Usage: "",
		Value: "",
		EnvVar: "GDRIVE_LOCAL_CONFIG_PATH",
	},
	cli.IntFlag{
		Name:  "gdrive-chunk-size",
		Usage: "",
		Value: googleapi.DefaultUploadChunkSize / 1024 / 1024,
		EnvVar: "GDRIVE_CHUNK_SIZE",
	},
	cli.IntFlag{
		Name:   "rate-limit",
		Usage:  "requests per minute",
		Value:  0,
		EnvVar: "RATE_LIMIT",
	},
	cli.StringFlag{
		Name:   "lets-encrypt-hosts",
		Usage:  "host1, host2",
		Value:  "",
		EnvVar: "HOSTS",
	},
	cli.StringFlag{
		Name:  "log",
		Usage: "/var/log/transfersh.log",
		Value: "",
		EnvVar: "LOG",
	},
	cli.StringFlag{
		Name:  "basedir",
		Usage: "path to storage",
		Value: "",
		EnvVar: "BASEDIR",
	},
	cli.StringFlag{
		Name:   "clamav-host",
		Usage:  "clamav-host",
		Value:  "",
		EnvVar: "CLAMAV_HOST",
	},
	cli.StringFlag{
		Name:   "virustotal-key",
		Usage:  "virustotal-key",
		Value:  "",
		EnvVar: "VIRUSTOTAL_KEY",
	},
	cli.BoolFlag{
		Name:  "profiler",
		Usage: "enable profiling",
		EnvVar: "PROFILER",
	},
	cli.StringFlag{
		Name:  "http-auth-user",
		Usage: "user for http basic auth",
		Value: "",
		EnvVar: "HTTP_AUTH_USER",
	},
	cli.StringFlag{
		Name:  "http-auth-pass",
		Usage: "pass for http basic auth",
		Value: "",
		EnvVar: "HTTP_AUTH_PASS",
	},
	cli.StringFlag{
		Name:  "ip-whitelist",
		Usage: "comma separated list of ips allowed to connect to the service",
		Value: "",
		EnvVar: "IP_WHITELIST",
	},
	cli.StringFlag{
		Name:  "ip-blacklist",
		Usage: "comma separated list of ips not allowed to connect to the service",
		Value: "",
		EnvVar: "IP_BLACKLIST",
	},
	cli.StringFlag{
		Name:  "cors-domains",
		Usage: "comma separated list of domains allowed for CORS requests",
		Value: "",
		EnvVar: "CORS_DOMAINS",
	},
}

type Cmd struct {
	*cli.App
}

func VersionAction(c *cli.Context) {
	fmt.Println(color.YellowString(fmt.Sprintf("transfer.sh: Easy file sharing from the command line")))
}

func New() *Cmd {
	logger := log.New(os.Stdout, "[transfer.sh]", log.LstdFlags)

	app := cli.NewApp()
	app.Name = "transfer.sh"
	app.Author = ""
	app.Usage = "transfer.sh"
	app.Description = `Easy file sharing from the command line`
	app.Version = Version
	app.Flags = globalFlags
	app.CustomAppHelpTemplate = helpTemplate
	app.Commands = []cli.Command{
		{
			Name:   "version",
			Action: VersionAction,
		},
	}

	app.Before = func(c *cli.Context) error {
		return nil
	}

	app.Action = func(c *cli.Context) {
		options := []server.OptionFn{}
		if v := c.String("listener"); v != "" {
			options = append(options, server.Listener(v))
		}

		if v := c.String("cors-domains"); v != "" {
			options = append(options, server.CorsDomains(v))
		}

		if v := c.String("tls-listener"); v == "" {
		} else if c.Bool("tls-listener-only") {
			options = append(options, server.TLSListener(v, true))
		} else {
			options = append(options, server.TLSListener(v, false))
		}

		if v := c.String("profile-listener"); v != "" {
			options = append(options, server.ProfileListener(v))
		}

		if v := c.String("web-path"); v != "" {
			options = append(options, server.WebPath(v))
		}

		if v := c.String("proxy-path"); v != "" {
			options = append(options, server.ProxyPath(v))
		}

		if v := c.String("proxy-port"); v != "" {
			options = append(options, server.ProxyPort(v))
		}

		if v := c.String("ga-key"); v != "" {
			options = append(options, server.GoogleAnalytics(v))
		}

		if v := c.String("uservoice-key"); v != "" {
			options = append(options, server.UserVoice(v))
		}

		if v := c.String("temp-path"); v != "" {
			options = append(options, server.TempPath(v))
		}

		if v := c.String("log"); v != "" {
			options = append(options, server.LogFile(logger, v))
		} else {
			options = append(options, server.Logger(logger))
		}

		if v := c.String("lets-encrypt-hosts"); v != "" {
			options = append(options, server.UseLetsEncrypt(strings.Split(v, ",")))
		}

		if v := c.String("virustotal-key"); v != "" {
			options = append(options, server.VirustotalKey(v))
		}

		if v := c.String("clamav-host"); v != "" {
			options = append(options, server.ClamavHost(v))
		}

		if v := c.Int("rate-limit"); v > 0 {
			options = append(options, server.RateLimit(v))
		}

		if cert := c.String("tls-cert-file"); cert == "" {
		} else if pk := c.String("tls-private-key"); pk == "" {
		} else {
			options = append(options, server.TLSConfig(cert, pk))
		}

		if c.Bool("profiler") {
			options = append(options, server.EnableProfiler())
		}

		if c.Bool("force-https") {
			options = append(options, server.ForceHTTPs())
		}

		if httpAuthUser := c.String("http-auth-user"); httpAuthUser == "" {
		} else if httpAuthPass := c.String("http-auth-pass"); httpAuthPass == "" {
		} else {
			var authenticator server.DefaultServAuthenticator
			authenticator.Set(httpAuthUser, httpAuthPass)
			options = append(options, server.AuthCredential(server.ServerAuthKey, &authenticator))
		}

		applyIPFilter := false
		ipFilterOptions := server.IPFilterOptions{}
		if ipWhitelist := c.String("ip-whitelist"); ipWhitelist != "" {
			applyIPFilter = true
			ipFilterOptions.AllowedIPs = strings.Split(ipWhitelist, ",")
			ipFilterOptions.BlockByDefault = true
		}

		if ipBlacklist := c.String("ip-blacklist"); ipBlacklist != "" {
			applyIPFilter = true
			ipFilterOptions.BlockedIPs = strings.Split(ipBlacklist, ",")
		}

		if applyIPFilter {
			options = append(options, server.FilterOptions(ipFilterOptions))
		}

		fileStorage := getStorage(c, c.String("provider"), logger)
		if fileStorage == nil {
			panic("Provider not set or invalid.")
		}

		options = append(options, server.UseStorage(fileStorage))

		var metaStorage server.Storage
		if metaProvider := c.String("meta-provider"); metaProvider != "" {
			metaStorage = getStorage(c, metaProvider, logger)
		} else {
			metaStorage = fileStorage
		}

		if metaStorage == nil {
			panic("Metadata Provider not set or invalid.")
		}

		if endpoint := c.String("api-endpoint"); endpoint != "" {
			var headerM map[string]string
			if headerStr := c.String("api-headers"); headerStr != "" {
				headerM = make(map[string]string)
				headers := strings.Split(headerStr, ",")
				for _, header := range headers {
					pair := strings.SplitN(strings.TrimSpace(header), "=", 2)
					headerM[pair[0]] = pair[1]
				}
			}
			options = append(options, server.UseAPIAuthenticator(apiauth.APIConfig{
				Endpoint: endpoint,
				Headers:  headerM,
			}))
		}

		options = append(options, server.UseMetaStorage(metaStorage))

		srvr, err := server.New(
			options...,
		)

		if err != nil {
			logger.Println(color.RedString("Error starting server: %s", err.Error()))
			return
		}

		srvr.Run()
	}

	return &Cmd{
		App: app,
	}
}

func getStorage(c *cli.Context, provider string, logger *log.Logger) server.Storage {
	switch provider {
	case "s3":
		if accessKey := c.String("aws-access-key"); accessKey == "" {
			panic("access-key not set.")
		} else if secretKey := c.String("aws-secret-key"); secretKey == "" {
			panic("secret-key not set.")
		} else if bucket := c.String("bucket"); bucket == "" {
			panic("bucket not set.")
		} else if storage, err := server.NewS3Storage(accessKey, secretKey, bucket, c.String("s3-region"), c.String("s3-endpoint"), logger, c.Bool("s3-no-multipart"), c.Bool("s3-path-style")); err != nil {
			panic(err)
		} else {
			return storage
		}
	case "gdrive":
		chunkSize := c.Int("gdrive-chunk-size")

		if clientJsonFilepath := c.String("gdrive-client-json-filepath"); clientJsonFilepath == "" {
			panic("client-json-filepath not set.")
		} else if localConfigPath := c.String("gdrive-local-config-path"); localConfigPath == "" {
			panic("local-config-path not set.")
		} else if basedir := c.String("basedir"); basedir == "" {
			panic("basedir not set.")
		} else if storage, err := server.NewGDriveStorage(clientJsonFilepath, localConfigPath, basedir, chunkSize, logger); err != nil {
			panic(err)
		} else {
			return storage
		}
	case "local":
		if v := c.String("basedir"); v == "" {
			panic("basedir not set.")
		} else if storage, err := server.NewLocalStorage(v, logger); err != nil {
			panic(err)
		} else {
			return storage
		}
	case "redis":
		redisAddr := c.String("redis-addr")
		redisPwd := c.String("redis-pwd")
		if redisAddr == "" {
			return nil
		}
		return redisStorage.New(redisAddr, redisPwd)
	default:
		return nil
	}
}
