package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

func createSHA256Hash(data []byte) []byte {
	hasher := sha256.New()

	hasher.Write(data)

	return hasher.Sum(nil)
}

func compareHash(x, y []byte) bool {
	if len(x) != len(y) {
		return false
	}

	return subtle.ConstantTimeCompare(x, y) == 1
}

type Checksum struct {
	Hash []byte `json:"hash"`
}

type Changelog struct {
	Hash []byte `json:"hash"`
	Diff int
}

func getChecksum() (*Checksum, error) {
	checksumFd, err := os.OpenFile("checksum.json", os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		return nil, err
	}

	defer checksumFd.Close()

	checkSumFC, err := io.ReadAll(checksumFd)

	if err != nil {
		return nil, err
	}

	var checksum Checksum

	if len(checkSumFC) > 0 {
		err = json.Unmarshal(checkSumFC, &checksum)

		if err != nil {
			return nil, err
		}
	}

	return &checksum, err
}

func generateChecksum(content []byte) *Checksum {
	checksumHash := createSHA256Hash(content)
	return &Checksum{Hash: checksumHash}
}

func scanFileForChanges(content []byte, checksum *Checksum) *Changelog {

	newChecksumHash := createSHA256Hash(content)

	changeLog := &Changelog{}
	changeLog.Hash = newChecksumHash

	if checksum == nil {
		return changeLog
	}

	if len(checksum.Hash) != 0 && !compareHash(newChecksumHash, checksum.Hash) {
		changeLog.Diff = 1
	}

	return changeLog
}

func updateChecksum(changelog *Changelog) (err error) {
	checksum := &Checksum{Hash: changelog.Hash}

	cJson, err := json.Marshal(checksum)

	err = os.WriteFile("checksum.json", cJson, 0644)

	return
}

func main() {

	checksum, err := getChecksum()

	if err != nil {
		log.Fatal(err)
	}

	txtFd, err := os.Open("file.txt")

	if err != nil {
		log.Fatal(err)
	}

	defer txtFd.Close()

	txtContent, err := io.ReadAll(txtFd)

	changelog := scanFileForChanges(txtContent, checksum)

	err = updateChecksum(changelog)

	if err != nil {
		log.Fatal(err)
	}

	if changelog.Diff > 0 {
		fmt.Println("==> File changes detected <==")
	} else {
		fmt.Println("==> No file changes detected <==")
	}

}
