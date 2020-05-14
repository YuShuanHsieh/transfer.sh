# Download Service

This service forked from [https://github.com/dutchcoders/transfer.sh](https://github.com/dutchcoders/transfer.sh)
Easy and fast file sharing from the command-line. This code contains the server with everything you need to create your own instance.

## Start Service

```
go run main.go --provider="local" --meta-provider="redis" --redis-addr="localhost:6379" --basedir="./files" --http-auth-user="cherie" --http-auth-pass="cherie"
```

Or use API authenticator for files downloading

```
go run main.go --provider="local" --basedir="./files" --http-auth-user="cherie" --http-auth-pass="cherie" --api-endpoint=http://localhost:1313/auth
```

## Usage

### Upload:
```bash
$ curl -X PUT -F "file=@./examples.md" http://<your.domain.com>/hello.txt
```
### Upload with download restrictions

#### Form

Parameter | Description | Example 
--- | --- | ---
ip | allowed IPs or CIDR | 127.0.0.1,0.0.0.0/0
user | user for HTTP basic Auth | user
password | password for HTTP basic Auth | password
api | user account for the api authenticator | cherie

Example

```bash
curl -X POST -H "Content-Type: multipart/form-data" \
-F "user=user" \
-F "password=password" \
-F "file=@./examples.md" "http://localhost:8080/examples.md"
```

API
```bash
curl -X POST -H "Content-Type: multipart/form-data" \
-F "api=username" \
-F "file=@./examples.md" "http://localhost:8080/examples.md"
```

### Deleting
```bash
$ curl -X DELETE <X-Url-Delete Response Header URL>
```

## Request Headers

### Max-Downloads
```bash
$ curl -X PUT -F "file=@./examples.md" http://<your.domain.com>/hello.txt -H "Max-Downloads: 1" # Limit the number of downloads
```

### Max-Days
```bash
$ curl -X PUT -F "file=@./examples.md" http://<your.domain.com>/hello.txt -H "Max-Days: 1" # Set the number of days before deletion
```

## Response Headers

### X-Url-Delete

The URL used to request the deletion of a file. Returned as a response header.
```bash
curl -sD - -X PUT -F "file=@./examples.md" http://<your.domain.com>/hello.txt | grep 'X-Url-Delete'
X-Url-Delete: http://<your.domain.com>/BAYh0/hello.txt/PDw0NHPcqU
```

## Parameters

Parameter | Description | Value | Env
--- | --- | --- | ---
listener | port to use for http (:80) | |
profile-listener | port to use for profiler (:6060)| |
force-https | redirect to https | false |
tls-listener | port to use for https (:443) | |
tls-listener-only | flag to enable tls listener only | |
tls-cert-file | path to tls certificate | |
tls-private-key | path to tls private key | |
http-auth-user | user for basic http auth on upload | |
http-auth-pass | pass for basic http auth on upload | |
ip-whitelist | comma separated list of ips allowed to connect to the service | |
ip-blacklist | comma separated list of ips not allowed to connect to the service | |
temp-path | path to temp folder | system temp |
web-path | path to static web files (for development or custom front end) | |
proxy-path | path prefix when service is run behind a proxy | |
ga-key | google analytics key for the front end | |
uservoice-key | user voice key for the front end  | |
api-endpoint | the endpoint for api authenticator | | 
api-headers | the HTTP(s) headers for api authenticator | | 
provider | which storage provider to use | (s3, gdrive or local) |
meta-provider | which storage provider to use | (s3, gdrive, local, redis) |
redis-addr | The address of redis server | localhost:6379 | 
redis-pwd | The password of redis server | | 
aws-access-key | aws access key | | AWS_ACCESS_KEY
aws-secret-key | aws access key | | AWS_SECRET_KEY
bucket | aws bucket | | BUCKET
s3-endpoint | Custom S3 endpoint. | |
s3-region | region of the s3 bucket | eu-west-1 | S3_REGION
s3-no-multipart | disables s3 multipart upload | false | |
s3-path-style | Forces path style URLs, required for Minio. | false | |
basedir | path storage for local/gdrive provider| |
gdrive-client-json-filepath | path to oauth client json config for gdrive provider| |
gdrive-local-config-path | path to store local transfer.sh config cache for gdrive provider| |
gdrive-chunk-size | chunk size for gdrive upload in megabytes, must be lower than available memory (8 MB) | |
lets-encrypt-hosts | hosts to use for lets encrypt certificates (comma seperated) | |
log | path to log file| |
cors-domains | comma separated list of domains for CORS, setting it enable CORS | |
cors-domains | comma separated list of domains for CORS, setting it enable CORS | | CORS_DOMAINS |
clamav-host | host for clamav feature  | | CLAMAV_HOST |
rate-limit | request per minute  | | RATE_LIMIT |

If you want to use TLS using lets encrypt certificates, set lets-encrypt-hosts to your domain, set tls-listener to :443 and enable force-https.
If you want to use TLS using your own certificates, set tls-listener to :443, force-https, tls-cert-file and tls-private-key.

## Development

## Build

If on go < 1.11
```bash
go get -u -v ./...
```

```bash
go build -o transfersh main.go
```

## S3 Usage

For the usage with a AWS S3 Bucket, you just need to specify the following options:
- provider
- aws-access-key
- aws-secret-key
- bucket
- s3-region

If you specify the s3-region, you don't need to set the endpoint URL since the correct endpoint will used automatically.

### Custom S3 providers

To use a custom non-AWS S3 provider, you need to specify the endpoint as definied from your cloud provider.

## Google Drive Usage

For the usage with Google drive, you need to specify the following options:
- provider
- gdrive-client-json-filepath
- gdrive-local-config-path
- basedir

### Creating Gdrive Client Json

You need to create an Oauth Client id from console.cloud.google.com
download the file and place into a safe directory

### Usage example

```go run main.go --provider gdrive --basedir /tmp/ --gdrive-client-json-filepath /[credential_dir] --gdrive-local-config-path [directory_to_save_config] ```

## Contributions

Contributions are welcome.

## Creators

**Remco Verhoef**
- <https://twitter.com/remco_verhoef>
- <https://twitter.com/dutchcoders>

**Uvis Grinfelds**

## Maintainer

Cherie Hsieh (cherie@pufsecurity.com)

## Copyright and license

Code and documentation copyright 2011-2018 Remco Verhoef.
Code released under [the MIT license](LICENSE).
