package api

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

const (
	certValidityDuration = 30 * time.Minute
)

type Controller struct {
	Log           *slog.Logger
	allowedTokens map[string][]string
	caSigner      ssh.Signer
}

type requestBody struct {
	PublicKey string `json:"public_key"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type successResponse struct {
	SignedCert string `json:"signed_cert,omitempty"`
}

func NewController(logger *slog.Logger, allowedTokens map[string][]string, caSigner ssh.Signer) *Controller {
	return &Controller{Log: logger, allowedTokens: allowedTokens, caSigner: caSigner}
}

func (c *Controller) Sign(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get authentication token and corresponding principals
	authToken := r.Header.Get("Authorization")
	principals, err := c.getPrincipals(authToken)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error()})
		return
	}

	// Decode request body
	var reqBody requestBody
	err = json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "failed to decode request body"})
		return
	}

	// Parse public key
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(reqBody.PublicKey))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "failed to parse public key"})
		return
	}

	// Validate public key
	if pubKey.Type() != ssh.KeyAlgoED25519 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: "only ed25519 keys are supported"})
		return
	}

	// Sign 
	signedCert, err := signUserKey(pubKey, c.caSigner, principals)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{Error: "failed to sign certificate"})
		return
	}

	// Log certificate details
	issuedAt := time.Unix(int64(signedCert.ValidAfter), 0)
	expiresAt := time.Unix(int64(signedCert.ValidBefore), 0)
	c.Log.Info("certificate issued",
		"issued_at", issuedAt.Format(time.RFC3339),
		"principals", principals,
		"expires_at", expiresAt.Format(time.RFC3339),
		"serial", signedCert.Serial,
	)

	w.WriteHeader(http.StatusOK)
	certBytes := ssh.MarshalAuthorizedKey(signedCert)
	json.NewEncoder(w).Encode(successResponse{SignedCert: string(certBytes)})
}

func (c *Controller) getPrincipals(token string) ([]string, error) {
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	} else {
		return nil, fmt.Errorf("Invalid auth token syntax")
	}
	principals, exists := c.allowedTokens[token]
	if !exists {
		return nil, fmt.Errorf("access token not valid or has no principals")
	}
	return principals, nil
}

func signUserKey(
	userPubKey ssh.PublicKey,
	caSigner ssh.Signer,
	principals []string,
) (*ssh.Certificate, error) {

	cert := &ssh.Certificate{
		Key:             userPubKey,
		CertType:        ssh.UserCert,
		Serial:          uint64(time.Now().UnixNano()),
		ValidPrincipals: principals,
		ValidAfter:      uint64(time.Now().Unix()),
		ValidBefore:     uint64(time.Now().Add(certValidityDuration).Unix()),
	}

	if err := cert.SignCert(rand.Reader, caSigner); err != nil {
		return nil, err
	}

	return cert, nil
}
