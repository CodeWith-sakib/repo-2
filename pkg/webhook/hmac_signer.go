package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
)

type HashAlgorithm string

const (
	AlgorithmSHA256 HashAlgorithm = "sha256"
	AlgorithmSHA512 HashAlgorithm = "sha512"
)

type PayloadSigner struct {
	secret    []byte
	algorithm HashAlgorithm
}

func NewPayloadSigner(secret string, algo HashAlgorithm) *PayloadSigner {
	if algo != AlgorithmSHA512 {
		algo = AlgorithmSHA256
	}
	return &PayloadSigner{
		secret:    []byte(secret),
		algorithm: algo,
	}
}

func (s *PayloadSigner) Sign(payload []byte) (string, error) {
	var h hash.Hash
	switch s.algorithm {
	case AlgorithmSHA256:
		h = hmac.New(sha256.New, s.secret)
	case AlgorithmSHA512:
		h = hmac.New(sha512.New, s.secret)
	default:
		return "", fmt.Errorf("unsupported hash algorithm: %s", s.algorithm)
	}

	h.Write(payload)
	sum := h.Sum(nil)
	return fmt.Sprintf("%s=%s", s.algorithm, hex.EncodeToString(sum)), nil
}

func (s *PayloadSigner) Verify(payload []byte, headerValue string) bool {
	expectedSig, err := s.Sign(payload)
	if err != nil {
		return false
	}
	return hmac.Equal([]byte(expectedSig), []byte(headerValue))
}
