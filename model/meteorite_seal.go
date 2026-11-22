package model

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"os"

	"gorm.io/gorm"
)

// meteorite: API keys (users') and channel keys (upstreams') are stored sealed, never in plain
// text (R-005, R-019). A value is sealed with AES-256-GCM under METEORITE_SEAL_KEY. The nonce is
// derived from the value, so the same key always seals to the same text: that keeps the lookup
// "WHERE key = ?" working, and it is safe because every sealed value is a distinct random key.
// A stored value without the prefix is an older plain one; it is read as it is and sealed the
// next time it is saved.
const meteoriteSealPrefix = "mk1:"

var errMeteoriteSealKeyMissing = errors.New("METEORITE_SEAL_KEY is not set")

func meteoriteSealKeys() (encKey, nonceKey []byte, err error) {
	master := os.Getenv("METEORITE_SEAL_KEY")
	if len(master) < 32 {
		return nil, nil, errMeteoriteSealKeyMissing
	}
	enc := sha256.Sum256([]byte("meteorite-seal-enc|" + master))
	nonce := sha256.Sum256([]byte("meteorite-seal-nonce|" + master))
	return enc[:], nonce[:], nil
}

// MeteoriteSeal returns the sealed form of value; an empty or already sealed value is returned
// as it is.
func MeteoriteSeal(value string) (string, error) {
	if value == "" || len(value) >= len(meteoriteSealPrefix) && value[:len(meteoriteSealPrefix)] == meteoriteSealPrefix {
		return value, nil
	}
	encKey, nonceKey, err := meteoriteSealKeys()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(encKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, nonceKey)
	mac.Write([]byte(value))
	nonce := mac.Sum(nil)[:gcm.NonceSize()]
	sealed := gcm.Seal(nonce, nonce, []byte(value), nil)
	return meteoriteSealPrefix + base64.RawURLEncoding.EncodeToString(sealed), nil
}

// MeteoriteOpen returns the plain form of a stored value. A value that is not sealed is returned
// as it is; a sealed value that cannot be opened is returned empty, so it matches nothing.
func MeteoriteOpen(stored string) string {
	if len(stored) < len(meteoriteSealPrefix) || stored[:len(meteoriteSealPrefix)] != meteoriteSealPrefix {
		return stored
	}
	raw, err := base64.RawURLEncoding.DecodeString(stored[len(meteoriteSealPrefix):])
	if err != nil {
		return ""
	}
	encKey, _, err := meteoriteSealKeys()
	if err != nil {
		return ""
	}
	block, err := aes.NewCipher(encKey)
	if err != nil {
		return ""
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil || len(raw) < gcm.NonceSize() {
		return ""
	}
	plain, err := gcm.Open(nil, raw[:gcm.NonceSize()], raw[gcm.NonceSize():], nil)
	if err != nil {
		return ""
	}
	return string(plain)
}

func (token *Token) BeforeSave(tx *gorm.DB) error {
	sealed, err := MeteoriteSeal(token.Key)
	token.Key = sealed
	return err
}

func (token *Token) AfterSave(tx *gorm.DB) error {
	token.Key = MeteoriteOpen(token.Key)
	return nil
}

func (token *Token) AfterFind(tx *gorm.DB) error {
	token.Key = MeteoriteOpen(token.Key)
	return nil
}

func (channel *Channel) BeforeSave(tx *gorm.DB) error {
	sealed, err := MeteoriteSeal(channel.Key)
	channel.Key = sealed
	return err
}

func (channel *Channel) AfterSave(tx *gorm.DB) error {
	channel.Key = MeteoriteOpen(channel.Key)
	return nil
}

func (channel *Channel) AfterFind(tx *gorm.DB) error {
	channel.Key = MeteoriteOpen(channel.Key)
	return nil
}

// MeteoriteSealStoredKeys seals every API key and channel key that is still stored in plain
// text. It runs after the database migration at every start and changes nothing once all keys
// are sealed.
func MeteoriteSealStoredKeys() error {
	var tokens []struct {
		Id     int
		Secret string
	}
	if err := DB.Raw("SELECT id, " + commonKeyCol + " AS secret FROM tokens WHERE " + commonKeyCol + " NOT LIKE 'mk1:%'").Scan(&tokens).Error; err != nil {
		return err
	}
	for _, row := range tokens {
		sealed, err := MeteoriteSeal(row.Secret)
		if err != nil {
			return err
		}
		if err := DB.Exec("UPDATE tokens SET "+commonKeyCol+" = ? WHERE id = ?", sealed, row.Id).Error; err != nil {
			return err
		}
	}
	var channels []struct {
		Id     int
		Secret string
	}
	if err := DB.Raw("SELECT id, " + commonKeyCol + " AS secret FROM channels WHERE " + commonKeyCol + " NOT LIKE 'mk1:%' AND " + commonKeyCol + " <> ''").Scan(&channels).Error; err != nil {
		return err
	}
	for _, row := range channels {
		sealed, err := MeteoriteSeal(row.Secret)
		if err != nil {
			return err
		}
		if err := DB.Exec("UPDATE channels SET "+commonKeyCol+" = ? WHERE id = ?", sealed, row.Id).Error; err != nil {
			return err
		}
	}
	return nil
}
