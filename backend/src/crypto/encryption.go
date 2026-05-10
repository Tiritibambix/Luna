package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	goerrors "errors"
	"io"
	"luna-backend/errors"
	"net/http"

	"golang.org/x/crypto/argon2"
)

func encryptionCipher(key []byte) (cipher.AEAD, *errors.ErrorTrace) {
	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.New().Status(http.StatusInternalServerError).
			AddErr(errors.LvlDebug, err).
			Append(errors.LvlWordy, "Could not create AES block cipher from key %v", key).
			Append(errors.LvlWordy, "Could not create AES block cipher")
	}

	aeadCipher, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, errors.New().Status(http.StatusInternalServerError).
			AddErr(errors.LvlDebug, err).
			Append(errors.LvlWordy, "Could not create AES-GCM cipher")
	}

	return aeadCipher, nil
}

func DeriveEncryptionKeyFromPassword(password string, salt []byte, length uint32) []byte {
	return argon2.IDKey(
		[]byte(password),
		salt,
		uint32(DefaultArgon2Settings["time"]),
		uint32(DefaultArgon2Settings["memory"]),
		uint8(DefaultArgon2Settings["threads"]),
		length,
	)
}

type encryptWriter struct {
	reader  *io.PipeReader
	writer  *io.PipeWriter
	errChan chan (*errors.ErrorTrace)
}

func (writer *encryptWriter) Write(payload []byte) (n int, err error) {
	return writer.writer.Write(payload)
}

func (writer *encryptWriter) Close() error {
	err := writer.reader.Close()
	if err != nil {
		return err
	}
	tr := <-writer.errChan
	if tr != nil {
		return tr.SerializeError(errors.LvlDebug)
	}
	err = writer.writer.Close()
	if err != nil {
		return err
	}
	return nil
}

func EncryptStream(stream io.Writer, password string) (io.WriteCloser, *errors.ErrorTrace) {
	salt, tr := GenerateRandomBytes(32)
	if tr != nil {
		return nil, tr.Append(errors.LvlDebug, "Could not generate salt")
	}

	key := DeriveEncryptionKeyFromPassword(password, salt, 32)

	aeadCipher, tr := encryptionCipher(key)
	if tr != nil {
		return nil, tr.Append(errors.LvlDebug, "Could not create cipher")
	}

	nonce, tr := GenerateRandomBytes(aeadCipher.NonceSize())
	if tr != nil {
		return nil, tr.Append(errors.LvlDebug, "Could not generate nonce")
	}

	reader, writer := io.Pipe()
	errChan := make(chan (*errors.ErrorTrace))

	go func() {
		plaintext, err := io.ReadAll(reader)
		if err != nil && !goerrors.Is(err, io.ErrClosedPipe) {
			errChan <- errors.New().Status(http.StatusInternalServerError).
				AddErr(errors.LvlDebug, err).
				Append(errors.LvlDebug, "Could not read plaintext")
			return
		}

		ciphertext := aeadCipher.Seal(nil, nonce, plaintext, salt)

		_, err = stream.Write(salt)
		if err != nil {
			errChan <- errors.New().Status(http.StatusInternalServerError).
				AddErr(errors.LvlDebug, err).
				Append(errors.LvlDebug, "Could not write salt")
			return
		}
		_, err = stream.Write(nonce)
		if err != nil {
			errChan <- errors.New().Status(http.StatusInternalServerError).
				AddErr(errors.LvlDebug, err).
				Append(errors.LvlDebug, "Could not write nonce")
			return
		}
		_, err = stream.Write(ciphertext)
		if err != nil {
			errChan <- errors.New().Status(http.StatusInternalServerError).
				AddErr(errors.LvlDebug, err).
				Append(errors.LvlDebug, "Could not write ciphertext")
			return
		}

		errChan <- nil
	}()

	return &encryptWriter{
		reader:  reader,
		writer:  writer,
		errChan: errChan,
	}, nil
}

type decryptReader struct {
	reader  *io.PipeReader
	writer  *io.PipeWriter
	errChan chan (*errors.ErrorTrace)
}

func (reader *decryptReader) Read(payload []byte) (n int, err error) {
	return reader.reader.Read(payload)
}

func (reader *decryptReader) Close() error {
	err := reader.reader.Close()
	if err != nil {
		return err
	}
	tr := <-reader.errChan
	if tr != nil {
		return tr.SerializeError(errors.LvlDebug)
	}
	return nil
}

func (reader *decryptReader) error(tr *errors.ErrorTrace) {
	reader.writer.CloseWithError(tr.SerializeError(errors.LvlDebug))
	reader.errChan <- tr
}

func DecryptStream(stream io.Reader, password string) io.ReadCloser {
	reader, writer := io.Pipe()
	errChan := make(chan (*errors.ErrorTrace))

	dr := &decryptReader{
		reader:  reader,
		writer:  writer,
		errChan: errChan,
	}

	go func() {
		salt := make([]byte, 32)
		n, err := stream.Read(salt)
		if err != nil {
			dr.error(
				errors.New().Status(http.StatusInternalServerError).
					AddErr(errors.LvlDebug, err).
					Append(errors.LvlDebug, "Could not read salt"),
			)
			return
		}
		if n != 32 {
			dr.error(
				errors.New().Status(http.StatusInternalServerError).
					AddErr(errors.LvlDebug, err).
					Append(errors.LvlWordy, "The backup file does not contain salt").
					Append(errors.LvlPlain, "Malformed backup file"),
			)
			return
		}

		key := DeriveEncryptionKeyFromPassword(password, salt, 32)

		aeadCipher, tr := encryptionCipher(key)
		if tr != nil {
			dr.error(
				tr.Append(errors.LvlDebug, "Could not create cipher"),
			)
			return
		}

		nonce := make([]byte, aeadCipher.NonceSize())
		n, err = stream.Read(nonce)
		if err != nil {
			dr.error(
				errors.New().Status(http.StatusInternalServerError).
					AddErr(errors.LvlDebug, err).
					Append(errors.LvlDebug, "Could not read nonce"),
			)
			return
		}
		if n != aeadCipher.NonceSize() {
			dr.error(
				errors.New().Status(http.StatusBadRequest).
					AddErr(errors.LvlDebug, err).
					Append(errors.LvlWordy, "The backup file does not contain a nonce").
					Append(errors.LvlPlain, "Malformed backup file"),
			)
			return
		}

		ciphertext, err := io.ReadAll(stream)
		if err != nil {
			dr.error(
				errors.New().Status(http.StatusInternalServerError).
					AddErr(errors.LvlDebug, err).
					Append(errors.LvlDebug, "Could not read ciphertext"),
			)
			return
		}

		plaintext, err := aeadCipher.Open(nil, nonce, ciphertext, salt)
		if err != nil {
			dr.error(
				errors.New().Status(http.StatusUnauthorized).
					AddErr(errors.LvlDebug, err).
					Append(errors.LvlDebug, "Could not decrypt ciphertext or failed checksum verification"),
			)
			return
		}

		_, err = writer.Write(plaintext)
		if err != nil {
			dr.error(
				errors.New().Status(http.StatusInternalServerError).
					AddErr(errors.LvlDebug, err).
					Append(errors.LvlDebug, "Could not write plaintext"),
			)
			return
		}

		err = writer.Close()
		if err != nil {
			dr.error(
				errors.New().Status(http.StatusUnauthorized).
					AddErr(errors.LvlDebug, err).
					Append(errors.LvlDebug, "Could not close writer"),
			)
			return
		}

		errChan <- nil
	}()

	return dr
}
