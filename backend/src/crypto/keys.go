package crypto

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"luna-backend/config"
	"luna-backend/errors"
	"net/http"
	"os"
	"strings"

	"golang.org/x/crypto/hkdf"
)

const keyExtension = ".key"

func GenerateSymmetricKey(commonConfig *config.CommonConfig, name string) ([]byte, *errors.ErrorTrace) {
	secret, tr := GenerateRandomBytes(64)
	if tr != nil {
		return nil, tr.
			Append(errors.LvlDebug, "Could not generate symmetric key %v", name).
			AltStr(errors.LvlWordy, "Could not generate symmetric key")
	}

	encodedSecret := base64.StdEncoding.EncodeToString(secret)

	path := fmt.Sprintf("%s/%s%s", commonConfig.Env.GetKeysPath(), name, keyExtension)
	err := os.WriteFile(path, []byte(encodedSecret), 0660)
	if err != nil {
		return nil, errors.New().Status(http.StatusInternalServerError).
			AddErr(errors.LvlDebug, err).
			Append(errors.LvlDebug, "Could not write key file %v at %v", path, commonConfig.Env.GetKeysPath()).
			AltStr(errors.LvlWordy, "Could not write key file").
			Append(errors.LvlDebug, "Could not generate symmetric key %v", name).
			AltStr(errors.LvlWordy, "Could not generate symmetric key")
	}

	return secret, nil
}

func OverwriteSymmetricKey(commonConfig *config.CommonConfig, name string, key []byte) *errors.ErrorTrace {
	encodedSecret := base64.StdEncoding.EncodeToString(key)

	path := fmt.Sprintf("%s/%s%s", commonConfig.Env.GetKeysPath(), name, keyExtension)
	err := os.WriteFile(path, []byte(encodedSecret), 0660)
	if err != nil {
		return errors.New().Status(http.StatusInternalServerError).
			AddErr(errors.LvlDebug, err).
			Append(errors.LvlDebug, "Could not write key file %v at %v", path, commonConfig.Env.GetKeysPath()).
			AltStr(errors.LvlWordy, "Could not write key file").
			Append(errors.LvlDebug, "Could not generate symmetric key %v", name).
			AltStr(errors.LvlWordy, "Could not generate symmetric key")
	}

	return nil
}

func GetSymmetricKey(commonConfig *config.CommonConfig, name string) ([]byte, *errors.ErrorTrace) {
	path := fmt.Sprintf("%s/%s%s", commonConfig.Env.GetKeysPath(), name, keyExtension)

	_, err := os.Stat(path)
	if err == nil {
		encodedSecret, err := os.ReadFile(path)
		if err != nil {
			return nil, errors.New().Status(http.StatusInternalServerError).
				AddErr(errors.LvlDebug, err).
				Append(errors.LvlDebug, "Could not read key file %v at %v", path, commonConfig.Env.GetKeysPath()).
				AltStr(errors.LvlWordy, "Could not read key file").
				Append(errors.LvlDebug, "Could not get symmetric key %v", name).
				AltStr(errors.LvlWordy, "Could not get symmetric key")
		}

		secret, err := base64.StdEncoding.DecodeString(string(encodedSecret))
		if err != nil {
			return nil, errors.New().Status(http.StatusInternalServerError).
				AddErr(errors.LvlDebug, err).
				Append(errors.LvlDebug, "Could not decode key file %v at %v", path, commonConfig.Env.GetKeysPath()).
				AltStr(errors.LvlWordy, "Could not decode key file").
				Append(errors.LvlDebug, "Could not get symmetric key %v", name).
				AltStr(errors.LvlWordy, "Could not get symmetric key")
		}

		return secret, nil
	} else if os.IsNotExist(err) {
		return GenerateSymmetricKey(commonConfig, name)
	} else {
		return nil, errors.New().Status(http.StatusInternalServerError).
			AddErr(errors.LvlDebug, err).
			Append(errors.LvlDebug, "Could not access key file %v at %v", path, commonConfig.Env.GetKeysPath()).
			AltStr(errors.LvlWordy, "Could not access key file").
			Append(errors.LvlDebug, "Could not get symmetric key %v", name).
			AltStr(errors.LvlWordy, "Could not get symmetric key")
	}
}

func ListKeys(commonConfig *config.CommonConfig) ([]string, *errors.ErrorTrace) {
	path := commonConfig.Env.GetKeysPath()
	entries, err := os.ReadDir(path)

	if err != nil {
		return nil, errors.New().Status(http.StatusInternalServerError).
			AddErr(errors.LvlDebug, err).
			Append(errors.LvlDebug, "Could not read keys directory at %v", path).
			AltStr(errors.LvlWordy, "Could not read keys directory").
			Append(errors.LvlWordy, "Could not list keys")
	}

	keyNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		entryName := entry.Name()

		if !entry.Type().IsRegular() || !strings.HasSuffix(entryName, keyExtension) {
			continue
		}
		keyNames = append(keyNames, entryName[:len(entryName)-len(keyExtension)])
	}

	return keyNames, nil
}

func DeriveKey(secret []byte, salt []byte) ([]byte, error) {
	generator := hkdf.New(sha256.New, secret, salt, nil)
	newSecret := make([]byte, 64)
	_, err := generator.Read(newSecret)
	if err != nil {
		return nil, err
	}
	return newSecret, nil
}
