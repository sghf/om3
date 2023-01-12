package listener

import (
	"io/fs"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com/rs/zerolog/log"

	"opensvc.com/opensvc/core/keyop"
	"opensvc.com/opensvc/core/object"
	"opensvc.com/opensvc/core/path"
	"opensvc.com/opensvc/core/rawconfig"
	"opensvc.com/opensvc/daemon/daemonenv"
	"opensvc.com/opensvc/util/file"
	"opensvc.com/opensvc/util/filesystems"
	"opensvc.com/opensvc/util/findmnt"
)

func startCertFS() error {
	if err := mountCertFS(); err != nil {
		return err
	}

	if err := installCaFiles(); err != nil {
		return err
	}

	if err := installCertFiles(); err != nil {
		return err
	}

	return nil
}

func stopCertFS() error {
	tmpfs := filesystems.FromType("tmpfs")
	return tmpfs.Umount(rawconfig.Paths.Certs)
}

func mountCertFS() error {
	if v, err := findmnt.Has("none", rawconfig.Paths.Certs); err != nil {
		if err1, ok := err.(*exec.Error); ok {
			if err1.Name == "findmnt" && err1.Err == exec.ErrNotFound {
				// fallback when findmnt is not present
				if !file.Exists(rawconfig.Paths.Certs) {
					err := os.MkdirAll(rawconfig.Paths.Certs, 0700)
					if err != nil {
						return err
					}
				}
				return nil
			}
			return nil
		}
		return err
	} else if v {
		return nil
	}
	tmpfs := filesystems.FromType("tmpfs")
	if err := tmpfs.Mount("none", rawconfig.Paths.Certs, "rw,nosuid,nodev,noexec,relatime,size=1m"); err != nil {
		return err
	}
	return nil
}

func installCaFiles() error {
	var (
		caPath path.T
	)
	caPath, err := getSecCaPath()
	if err != nil {
		return err
	}
	if !caPath.Exists() {
		log.Logger.Info().Msgf("bootstrap initial %s", caPath)
		if err := bootStrapCaPath(caPath); err != nil {
			return err
		}
	}
	caSec, err := object.NewSec(caPath, object.WithVolatile(true))
	if err != nil {
		return err
	}

	err, usr, grp, fmode, dmode := getCertFilesModes()
	if err != nil {
		return err
	}

	// ca_certificates for jwt
	dst := daemonenv.CAKeyFile()

	if err := caSec.InstallKeyTo("private_key", dst, &fmode, &dmode, usr, grp); err != nil {
		return err
	} else {
		log.Logger.Info().Msgf("installed %s", dst)
	}

	dst = daemonenv.CACertChainFile()
	if err := caSec.InstallKeyTo("certificate_chain", dst, &fmode, &dmode, usr, grp); err != nil {
		return err
	} else {
		log.Logger.Info().Msgf("installed %s", dst)
	}

	// ca_certificates
	var b []byte
	validCA := make([]string, 0)
	caList := []string{caPath.String()}
	caList = append(caList, strings.Fields(rawconfig.ClusterSection().CASecPaths)...)
	for _, p := range caList {
		caPath, err := path.Parse(p)
		if err != nil {
			log.Logger.Warn().Err(err).Msgf("parse ca %s", p)
			continue
		}
		if !caPath.Exists() {
			log.Logger.Warn().Msgf("skip %s ca: sec object does not exist", caPath)
			continue
		}
		caSec, err := object.NewSec(caPath, object.WithVolatile(true))
		if err != nil {
			return err
		}
		chain, err := caSec.DecodeKey("certificate_chain")
		if err != nil {
			return err
		}
		b = append(b, chain...)
		validCA = append(validCA, p)
	}
	if len(b) > 0 {
		dst := daemonenv.CAsCertFile()
		if err := os.WriteFile(dst, b, fmode); err != nil {
			return err
		}
		log.Logger.Info().Strs("ca", validCA).Msgf("installed %s", dst)
	}

	// TODO: ca_crl
	return nil
}

func installCertFiles() error {
	certPath, err := getSecCertPath()
	if err != nil {
		return err
	}
	caPath, err := getSecCaPath()
	if err != nil {
		return err
	}
	if !certPath.Exists() {
		log.Logger.Info().Msgf("bootstrap initial %s", certPath)
		if err := bootStrapCertPath(certPath, caPath); err != nil {
			return err
		}
	}
	certSec, err := object.NewSec(certPath, object.WithVolatile(true))
	if err != nil {
		return err
	}
	err, usr, grp, fmode, dmode := getCertFilesModes()
	if err != nil {
		return err
	}
	dst := daemonenv.KeyFile()
	if err := certSec.InstallKeyTo("private_key", dst, &fmode, &dmode, usr, grp); err != nil {
		return err
	} else {
		log.Logger.Info().Msgf("installed %s", dst)
	}
	dst = daemonenv.CertChainFile()
	if err := certSec.InstallKeyTo("certificate_chain", dst, &fmode, &dmode, usr, grp); err != nil {
		return err
	} else {
		log.Logger.Info().Msgf("installed %s", dst)
	}

	dst = daemonenv.CertFile()
	if err := certSec.InstallKeyTo("certificate", dst, &fmode, &dmode, usr, grp); err != nil {
		return err
	} else {
		log.Logger.Info().Msgf("installed %s", dst)
	}
	return nil
}

func getCertFilesModes() (err error, usr *user.User, grp *user.Group, fmode, dmode fs.FileMode) {
	usr, err = user.Lookup("root")
	if err != nil {
		return
	}
	grp, err = user.LookupGroupId(usr.Gid)
	if err != nil {
		return
	}
	fmode = 0600
	dmode = 0700
	return
}

func bootStrapCaPath(p path.T) error {
	caSec, err := object.NewSec(p, object.WithVolatile(false))
	if err != nil {
		return err
	}
	return caSec.GenCert()
}

func bootStrapCertPath(p path.T, caPath path.T) error {
	certSec, err := object.NewSec(p, object.WithVolatile(false))
	if err != nil {
		return err
	}
	op := keyop.Parse("ca=" + caPath.String())
	if err := certSec.Config().Set(*op); err != nil {
		return err
	}
	return certSec.GenCert()
}

func getSecCaPath() (path.T, error) {
	return path.Parse("system/sec/ca-" + rawconfig.ClusterSection().Name)
}

func getSecCertPath() (path.T, error) {
	return path.Parse("system/sec/cert-" + rawconfig.ClusterSection().Name)
}
