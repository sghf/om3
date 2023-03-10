package daemonenv

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/opensvc/om3/core/rawconfig"
)

var (
	RawPort  = 1214
	HttpPort = 1215

	HeaderNode        = "o-node"
	HeaderMultiplexed = "o-multiplexed"

	DrainChanDuration = 40 * time.Millisecond

	// ReadyDuration define the minimum time to wait during the startup of an instance object by imon
	// The ready duration impacts the durations involved during daemon cluster split analyse (see nmon spit
	// detection details).
	ReadyDuration = 5 * time.Second
)

func CAKeyFile() string {
	return filepath.Join(rawconfig.Paths.Certs, "ca_private_key")
}

func CACertChainFile() string {
	return filepath.Join(rawconfig.Paths.Certs, "ca_certificate_chain")
}

func CAsCertFile() string {
	return filepath.Join(rawconfig.Paths.Certs, "ca_certificates")
}

func CertChainFile() string {
	return filepath.Join(rawconfig.Paths.Certs, "certificate_chain")
}

func CertFile() string {
	return filepath.Join(rawconfig.Paths.Certs, "certificate")
}

func KeyFile() string {
	return filepath.Join(rawconfig.Paths.Certs, "private_key")
}

func UrlInetRaw() string {
	return fmt.Sprintf("raw://localhost:%d", RawPort)
}

func UrlInetHttp() string {
	return fmt.Sprintf("https://localhost:%d", HttpPort)
}

func UrlHttpNode(node string) string {
	return fmt.Sprintf("https://%s:%d", node, HttpPort)
}

func UrlUxRaw() string {
	return "raw://" + PathUxRaw()
}

func UrlUxHttp() string {
	return "http://" + PathUxHttp()
}

func PathUxProfile() string {
	return filepath.Join(rawconfig.Paths.Lsnr, "profile.sock")
}

func PathUxRaw() string {
	return filepath.Join(rawconfig.Paths.Lsnr, "lsnr.sock")
}

func PathUxHttp() string {
	return filepath.Join(rawconfig.Paths.Lsnr, "h2.sock")
}
