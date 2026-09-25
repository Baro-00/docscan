package hashing

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

type Hashes struct {
	MD5    string
	SHA256 string
}

func File(path string) (Hashes, error) {
	f, err := os.Open(path)
	if err != nil {
		return Hashes{}, err
	}
	defer f.Close()

	md5Hash := md5.New()
	sha256Hash := sha256.New()

	writer := io.MultiWriter(md5Hash, sha256Hash)

	if _, err := io.Copy(writer, f); err != nil {
		return Hashes{}, err
	}

	return Hashes{
		MD5:    hex.EncodeToString(md5Hash.Sum(nil)),
		SHA256: hex.EncodeToString(sha256Hash.Sum(nil)),
	}, nil
}
