package handlers

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"luna-backend/api/internal/util"
	"luna-backend/crypto"
	"luna-backend/errors"
	"luna-backend/files"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
)

var archivedKeyFileRegex = regexp.MustCompile("keys/([^\\s\\./]+)\\.key")

func CreateBackup(c *gin.Context) {
	u := util.GetUtil(c)

	// Current time
	curTime := time.Now()

	// Create database backup
	dbBackup, err := u.DbBackups.CreateBackup()
	if err != nil {
		u.Error(err.Append(errors.LvlPlain, "Could not create a backup"))
		return
	}

	// Get all cryptographic keys
	keyNames, err := crypto.ListKeys(u.Config)
	if err != nil {
		u.Error(err.Append(errors.LvlPlain, "Could not create a backup"))
		return
	}

	keys := make([]string, len(keyNames))
	for i, name := range keyNames {
		rawKey, err := crypto.GetSymmetricKey(u.Config, name)
		if err != nil {
			u.Error(err.Append(errors.LvlPlain, "Could not create a backup"))
			return
		}
		keys[i] = base64.StdEncoding.EncodeToString(rawKey)
	}

	// Create backup archive structure
	type backupEntry struct {
		Name string
		Body []byte
	}
	backupFiles := []backupEntry{
		{"postgres.dump", []byte(dbBackup)},
	}
	for i, name := range keyNames {
		backupFiles = append(backupFiles, backupEntry{
			Name: fmt.Sprintf("keys/%s.key", name),
			Body: []byte(keys[i]),
		})
	}

	// Tar all files
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	tarWrite := tar.NewWriter(gzipWriter)

	for _, file := range backupFiles {
		hdr := &tar.Header{
			Name:       file.Name,
			Mode:       0600,
			Size:       int64(len(file.Body)),
			ModTime:    curTime,
			AccessTime: curTime,
			ChangeTime: curTime,
		}
		if err := tarWrite.WriteHeader(hdr); err != nil {
			u.Error(errors.New().Status(http.StatusInternalServerError).
				AddErr(errors.LvlDebug, err).
				Append(errors.LvlWordy, "Could not write file header").
				Append(errors.LvlPlain, "Could not create backup archive"),
			)
		}
		if _, err := tarWrite.Write(file.Body); err != nil {
			u.Error(errors.New().Status(http.StatusInternalServerError).
				AddErr(errors.LvlDebug, err).
				Append(errors.LvlWordy, "Could not write file contents").
				Append(errors.LvlPlain, "Could not create backup archive"),
			)
		}
	}
	if err := tarWrite.Close(); err != nil {
		u.Error(errors.New().Status(http.StatusInternalServerError).
			AddErr(errors.LvlDebug, err).
			Append(errors.LvlWordy, "Could not close tar writer").
			Append(errors.LvlPlain, "Could not create backup archive"),
		)
	}
	if err := gzipWriter.Close(); err != nil {
		u.Error(errors.New().Status(http.StatusInternalServerError).
			AddErr(errors.LvlDebug, err).
			Append(errors.LvlWordy, "Could not close gzip writer").
			Append(errors.LvlPlain, "Could not create backup archive"),
		)
	}

	// Respond with the backup archive
	u.ResponseWithFile(files.NewVolatileFile(fmt.Sprintf("luna-backup-%s.tar.gz", curTime.Format("2006-01-02-15-04-05")), buf.Bytes()), "application/gzip")
}

func RestoreBackup(c *gin.Context) {
	u := util.GetUtil(c)

	// Receive file
	backupFileHeader, fileErr := c.FormFile("backup_file")
	switch fileErr {
	case nil:
		break
	case http.ErrMissingFile:
		u.Error(errors.New().Status(http.StatusBadRequest).
			AddErr(errors.LvlDebug, fileErr).
			Append(errors.LvlPlain, "Missing backup file"),
		)
		return
	default:
		u.Error(errors.New().Status(http.StatusBadRequest).
			AddErr(errors.LvlDebug, fileErr).
			Append(errors.LvlPlain, "Invalid form data"),
		)
		return
	}

	file, err := backupFileHeader.Open()
	if err != nil {
		u.Error(errors.New().Status(http.StatusInternalServerError).
			AddErr(errors.LvlDebug, err).
			Append(errors.LvlPlain, "Could not open backup file"),
		)
		return
	}

	buf := make([]byte, backupFileHeader.Size)
	_, err = file.Read(buf)
	if err != nil {
		u.Error(errors.New().Status(http.StatusInternalServerError).
			AddErr(errors.LvlDebug, err).
			Append(errors.LvlPlain, "Could not read backup file"),
		)
		return
	}

	gzipReader, err := gzip.NewReader(bytes.NewReader(buf))
	if err != nil {
		u.Error(errors.New().Status(http.StatusInternalServerError).
			AddErr(errors.LvlDebug, err).
			Append(errors.LvlWordy, "Failed to create a gzip reader").
			Append(errors.LvlPlain, "Could not decompress backup file"),
		)
		return
	}
	tarReader := tar.NewReader(gzipReader)

	// Read all files
	for {
		tarHeader, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			u.Error(errors.New().Status(http.StatusInternalServerError).
				AddErr(errors.LvlDebug, err).
				Append(errors.LvlWordy, "Tar header ended abruptly").
				Append(errors.LvlPlain, "Could not decode backup file"),
			)
			return
		}

		var body bytes.Buffer
		_, err = io.Copy(&body, tarReader)
		if err != nil {
			u.Error(errors.New().Status(http.StatusInternalServerError).
				AddErr(errors.LvlDebug, err).
				Append(errors.LvlWordy, "File ended abruptly").
				Append(errors.LvlPlain, "Could not decode backup file"),
			)
			return
		}

		switch name := tarHeader.Name; {
		case name == "postgres.dump":
			tr := u.DbBackups.RestoreBackup(&body)
			if tr != nil {
				u.Error(tr.Append(errors.LvlPlain, "Could not restore the backup"))
				return
			}

		case archivedKeyFileRegex.MatchString(name):
			matches := archivedKeyFileRegex.FindStringSubmatch(name)
			if len(matches) != 2 {
				panic("amount of capture groups does not match the compiled regular expression")
			}
			keyName := matches[1]
			encodedKey, err := io.ReadAll(&body)
			if err != nil {
				u.Error(errors.New().Status(http.StatusInternalServerError).
					AddErr(errors.LvlDebug, err).
					Append(errors.LvlDebug, "Could not read key %s", keyName).
					Append(errors.LvlWordy, "File was malformed").
					Append(errors.LvlPlain, "Could not decode backup file"),
				)
				return
			}
			decodedKey, err := base64.StdEncoding.DecodeString(string(encodedKey))
			if err != nil {
				u.Error(errors.New().Status(http.StatusInternalServerError).
					AddErr(errors.LvlDebug, err).
					Append(errors.LvlDebug, "Could not decode key %s", keyName).
					Append(errors.LvlWordy, "File was malformed").
					Append(errors.LvlPlain, "Could not decode backup file"),
				)
				return
			}
			crypto.OverwriteSymmetricKey(u.Config, keyName, decodedKey)

		default:
			u.Warn(errors.New().Status(http.StatusBadRequest).
				Append(errors.LvlDebug, "The backup archive contains an unknown file %s", name).
				AltStr(errors.LvlWordy, "The backup archive contains an unknown file").
				Append(errors.LvlPlain, "The backup could not be restored fully. Was it made for a more recent version of Luna?"),
			)
		}
	}

	u.Success(nil)
}
