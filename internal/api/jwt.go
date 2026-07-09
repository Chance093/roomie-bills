package api

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/Chance093/roomie-bills/internal/lib"
	"github.com/golang-jwt/jwt/v4"
	"github.com/plaid/plaid-go/v43/plaid"
)

var maxAge = 5 * time.Minute

func verifyWebhook(webhookBody []byte, headers map[string]string) (bool, error) {
	tokenString := getHeaderCI(headers, "Plaid-Verification")
	if tokenString == "" {
		return false, errors.New("missing Plaid-Verification header")
	}

	// Decode JWT header (unverified) to extract alg and kid
	parser := jwt.Parser{}
	unverified, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return false, fmt.Errorf("parse unverified token: %w", err)
	}
	if unverified.Method.Alg() != jwt.SigningMethodES256.Alg() {
		return false, fmt.Errorf("unexpected alg %q (want ES256)", unverified.Method.Alg())
	}
	kid, _ := unverified.Header["kid"].(string)
	if kid == "" {
		return false, errors.New("missing kid in JWT header")
	}

	// Get verification key for kid via /webhook_verification_key/get
	pc := lib.NewPlaidClient()
	jwk, err := pc.GetJWK(kid)
	if err != nil {
		return false, fmt.Errorf("get JWK: %w", err)
	}
	pubKey, err := jwkToECDSAPublicKey(jwk)
	if err != nil {
		return false, fmt.Errorf("jwk->ecdsa: %w", err)
	}

	// Verify JWT signature
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodES256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return pubKey, nil
	})
	if err != nil || !token.Valid {
		return false, fmt.Errorf("invalid token: %w", err)
	}

	// Verify that the webhook is not more than 5 minutes old
	iatVal, ok := claims["iat"]
	if !ok {
		return false, errors.New("missing iat")
	}
	var iat time.Time
	switch v := iatVal.(type) {
	case float64:
		iat = time.Unix(int64(v), 0)
	case int64:
		iat = time.Unix(v, 0)
	default:
		return false, errors.New("invalid iat type")
	}
	if time.Since(iat) > maxAge {
		return false, errors.New("token too old (>5m)")
	}

	// Verify body hash integrity
	wantHash, ok := claims["request_body_sha256"].(string)
	if !ok || wantHash == "" {
		return false, errors.New("missing request_body_sha256")
	}
	sum := sha256.Sum256(webhookBody)
	gotHex := strings.ToLower(hex.EncodeToString(sum[:]))
	if subtle.ConstantTimeCompare([]byte(gotHex), []byte(strings.ToLower(wantHash))) != 1 {
		return false, errors.New("body hash mismatch")
	}

	return true, nil
}

// helper method
func jwkToECDSAPublicKey(jwk *plaid.JWKPublicKey) (*ecdsa.PublicKey, error) {
	if jwk == nil || jwk.X == "" || jwk.Y == "" ||
		jwk.Kty != "EC" || jwk.Crv != "P-256" {
		return nil, errors.New("invalid/unsupported JWK")
	}
	xBytes, err := base64.RawURLEncoding.DecodeString(jwk.X)
	if err != nil {
		return nil, fmt.Errorf("decode x: %w", err)
	}
	yBytes, err := base64.RawURLEncoding.DecodeString(jwk.Y)
	if err != nil {
		return nil, fmt.Errorf("decode y: %w", err)
	}
	return &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}, nil
}

func getHeaderCI(h map[string]string, name string) string {
	lname := strings.ToLower(name)
	for k, v := range h {
		if strings.ToLower(k) == lname {
			return v
		}
	}
	return ""
}
