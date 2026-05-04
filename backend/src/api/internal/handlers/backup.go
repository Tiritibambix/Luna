package handlers

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"luna-backend/api/internal/util"
	"luna-backend/crypto"
	"luna-backend/errors"
	"luna-backend/files"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateBackup(c *gin.Context) {
	u := util.GetUtil(c)

	// Current time
	curTime := time.Now()

	// Create database backup
	dbBackup, err := u.Tx.Backups().CreateBackup()
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

	keys := make([][]byte, len(keyNames))
	for i, name := range keyNames {
		keys[i], err = crypto.GetSymmetricKey(u.Config, name)
		if err != nil {
			u.Error(err.Append(errors.LvlPlain, "Could not create a backup"))
			return
		}
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
			Body: keys[i],
		})
	}

	// Tar all files
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	for _, file := range backupFiles {
		hdr := &tar.Header{
			Name:       file.Name,
			Mode:       0600,
			Size:       int64(len(file.Body)),
			ModTime:    curTime,
			AccessTime: curTime,
			ChangeTime: curTime,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			u.Error(errors.New().Status(http.StatusInternalServerError).
				AddErr(errors.LvlDebug, err).
				Append(errors.LvlWordy, "Could not write file header").
				Append(errors.LvlPlain, "Could not create backup archive"),
			)
		}
		if _, err := tw.Write(file.Body); err != nil {
			u.Error(errors.New().Status(http.StatusInternalServerError).
				AddErr(errors.LvlDebug, err).
				Append(errors.LvlWordy, "Could not write file contents").
				Append(errors.LvlPlain, "Could not create backup archive"),
			)
		}
	}
	if err := tw.Close(); err != nil {
		u.Error(errors.New().Status(http.StatusInternalServerError).
			AddErr(errors.LvlDebug, err).
			Append(errors.LvlWordy, "Could not close tar writer").
			Append(errors.LvlPlain, "Could not create backup archive"),
		)
	}
	if err := gw.Close(); err != nil {
		u.Error(errors.New().Status(http.StatusInternalServerError).
			AddErr(errors.LvlDebug, err).
			Append(errors.LvlWordy, "Could not close gzip writer").
			Append(errors.LvlPlain, "Could not create backup archive"),
		)
	}

	// Respond with the backup archive
	u.ResponseWithFile(files.NewVolatileFile(fmt.Sprintf("luna-backup-%s.tar.gz", curTime.Format("2006-01-02-15-04-05")), buf.Bytes()), "application/gzip")
}
