package util

import (
	"errors"
)

//var databaseFileUrlRegex = regexp.MustCompile(`^/api/files/([a-f0-9-]{36})$`)

func IsValidUsername(username string) error {
	if len(username) < 3 {
		return errors.New("username must be at least 3 characters long")
	}
	if len(username) > 25 {
		return errors.New("username must be at most 25 characters long")
	}
	return nil
}

func IsValidPassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	if len(password) > 1000 {
		return errors.New("password must be at most 1000 characters long")
	}
	return nil
}

//func IsDatabaseFileUrl(url *types.Url) (types.ID, error) {
//	urlStr := url.String()
//	matches := databaseFileUrlRegex.FindStringSubmatch(urlStr)
//	if len(matches) != 2 {
//		return types.ID{}, errors.New("invalid database file url")
//	}
//	id, err := types.IdFromString(matches[1])
//	if err != nil {
//		return types.ID{}, errors.New("invalid database file url")
//	}
//	return id, nil
//}
