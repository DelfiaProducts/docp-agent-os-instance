package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GenerateUniqHash return uniq hash
func GenerateUniqHash() string {
	uuidValue := uuid.New()
	timestamp := time.Now().UnixNano()
	dataToHash := fmt.Sprintf("%s%d", uuidValue.String(), timestamp)
	hash := sha256.Sum256([]byte(dataToHash))
	return hex.EncodeToString(hash[:])
}

// GenerateMd5Hash return hash from slice data
func GenerateMd5Hash(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}
