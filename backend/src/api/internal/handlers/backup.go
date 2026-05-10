package handlers

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"luna-backend/api/internal/util"
	"luna-backend/crypto"
	"luna-backend/errors"
	"luna-backend/files"
	"luna-backend/types"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
)

var archivedKeyFileRegex = regexp.MustCompile("keys/([^\\s\\./]+)\\.key")

type backupMetadata struct {
	Version   types.Version `json:"version"`
	Timestamp time.Time     `json:"timestamp"`
}

func CreateBackup(c *gin.Context) {
	u := util.GetUtil(c)

	// Current time
	curTime := time.Now()

	// Create backup metadata
	meta := backupMetadata{
		Version:   u.Config.Version,
		Timestamp: curTime,
	}
	marshalledMeta, err := json.Marshal(meta)
	if err != nil {
		u.Error(errors.New().Status(http.StatusInternalServerError).
			AddErr(errors.LvlDebug, err).
			Append(errors.LvlWordy, "Could not marshal backup metadata").
			Append(errors.LvlPlain, "Could not create backup archive"),
		)
		return
	}

	// Create database backup
	dbBackup, tr := u.DbBackups.CreateBackup()
	if tr != nil {
		u.Error(tr.Append(errors.LvlPlain, "Could not create a backup"))
		return
	}

	// Get all cryptographic keys
	keyNames, tr := crypto.ListKeys(u.Config)
	if tr != nil {
		u.Error(tr.Append(errors.LvlPlain, "Could not create a backup"))
		return
	}

	keys := make([]string, len(keyNames))
	for i, name := range keyNames {
		rawKey, tr := crypto.GetSymmetricKey(u.Config, name)
		if tr != nil {
			u.Error(tr.Append(errors.LvlPlain, "Could not create a backup"))
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
		{"metadata.json", marshalledMeta},
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
	var encryptWriter io.WriteCloser
	if password := c.PostForm("backup_password"); len(password) != 0 {
		encryptWriter, tr = crypto.EncryptStream(&buf, password)
		if tr != nil {
			u.Error(tr.Append(errors.LvlPlain, "Could not encrypt the backup"))
			return
		}
		defer encryptWriter.Close()
	}
	var gzipWriter *gzip.Writer
	if encryptWriter == nil {
		gzipWriter = gzip.NewWriter(&buf)
	} else {
		gzipWriter = gzip.NewWriter(encryptWriter)
	}
	defer gzipWriter.Close()
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	for _, file := range backupFiles {
		tarHeader := &tar.Header{
			Name:       file.Name,
			Mode:       0600,
			Size:       int64(len(file.Body)),
			ModTime:    curTime,
			AccessTime: curTime,
			ChangeTime: curTime,
		}
		if err := tarWriter.WriteHeader(tarHeader); err != nil {
			u.Error(errors.New().Status(http.StatusInternalServerError).
				AddErr(errors.LvlDebug, err).
				Append(errors.LvlWordy, "Could not write file header").
				Append(errors.LvlPlain, "Could not create backup archive"),
			)
			return
		}
		if _, err := tarWriter.Write(file.Body); err != nil {
			u.Error(errors.New().Status(http.StatusInternalServerError).
				AddErr(errors.LvlDebug, err).
				Append(errors.LvlWordy, "Could not write file contents").
				Append(errors.LvlPlain, "Could not create backup archive"),
			)
			return
		}
	}
	if err := tarWriter.Close(); err != nil {
		u.Error(errors.New().Status(http.StatusInternalServerError).
			AddErr(errors.LvlDebug, err).
			Append(errors.LvlWordy, "Could not close tar writer").
			Append(errors.LvlPlain, "Could not create backup archive"),
		)
		return
	}
	if err := gzipWriter.Close(); err != nil {
		u.Error(errors.New().Status(http.StatusInternalServerError).
			AddErr(errors.LvlDebug, err).
			Append(errors.LvlWordy, "Could not close gzip writer").
			Append(errors.LvlPlain, "Could not create backup archive"),
		)
		return
	}
	if encryptWriter != nil {
		if err := encryptWriter.Close(); err != nil {
			u.Error(errors.New().Status(http.StatusInternalServerError).
				AddErr(errors.LvlDebug, err).
				Append(errors.LvlWordy, "Could not close encryption writer").
				Append(errors.LvlPlain, "Could not create backup archive"),
			)
			return
		}
	}

	// Respond with the backup archive
	var extension string
	if encryptWriter == nil {
		extension = "tar.gz"
	} else {
		extension = "tar.gz.encrypted"
	}
	u.ResponseWithFile(files.NewVolatileFile(fmt.Sprintf("luna-backup-%s.%s", curTime.Format("2006-01-02-15-04-05"), extension), buf.Bytes()), "application/gzip")
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

	var decryptReader io.ReadCloser
	if password := c.PostForm("backup_password"); len(password) != 0 {
		decryptReader = crypto.DecryptStream(bytes.NewReader(buf), password)
		defer decryptReader.Close()
	}
	var gzipReader *gzip.Reader
	if decryptReader == nil {
		gzipReader, err = gzip.NewReader(bytes.NewReader(buf))
	} else {
		gzipReader, err = gzip.NewReader(decryptReader)
	}
	if err != nil {
		u.Error(errors.New().Status(http.StatusInternalServerError).
			AddErr(errors.LvlDebug, err).
			Append(errors.LvlWordy, "Failed to create a gzip reader").
			Append(errors.LvlPlain, "Could not decompress backup file"),
		)
		return
	}
	defer gzipReader.Close()
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
		case name == "metadata.json":
			marshalledMeta, err := io.ReadAll(&body)
			if err != nil {
				u.Error(errors.New().Status(http.StatusInternalServerError).
					AddErr(errors.LvlDebug, err).
					Append(errors.LvlWordy, "Could not read backup metadata").
					Append(errors.LvlWordy, "File was malformed").
					Append(errors.LvlPlain, "Could not decode backup file"),
				)
				return
			}
			meta := backupMetadata{}
			err = json.Unmarshal(marshalledMeta, &meta)
			if err != nil {
				u.Error(errors.New().Status(http.StatusInternalServerError).
					AddErr(errors.LvlDebug, err).
					Append(errors.LvlWordy, "Could not unmarshal backup metadata").
					Append(errors.LvlWordy, "File was malformed").
					Append(errors.LvlPlain, "Could not decode backup file"),
				)
				return
			}
			if meta.Version.IsGreaterThan(&u.Config.Version) {
				u.Error(errors.New().Status(http.StatusBadRequest).
					AddErr(errors.LvlDebug, err).
					Append(errors.LvlWordy, "The backup was created for a higher version of Luna (%s)", meta.Version.String()).
					AltStr(errors.LvlPlain, "The backup was created for a higher version of Luna").
					Append(errors.LvlPlain, "Could not restore backup"),
				)
				return
			}

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

	if err := gzipReader.Close(); err != nil {
		u.Warn(errors.New().Status(http.StatusInternalServerError).
			AddErr(errors.LvlDebug, err).
			Append(errors.LvlWordy, "Could not close gzip writer"),
		)
		return
	}
	if decryptReader != nil {
		if err := decryptReader.Close(); err != nil {
			u.Warn(errors.New().Status(http.StatusInternalServerError).
				AddErr(errors.LvlDebug, err).
				Append(errors.LvlWordy, "Could not close encryption writer"),
			)
			return
		}
	}

	u.Success(nil)
}
