package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Claims struct {
	Subject string
	Role    string
	Email   string
}

type SupabaseVerifier struct {
	jwksURL  string
	issuer   string
	audience string

	client *http.Client

	mu         sync.RWMutex
	ecKeys     map[string]*ecdsa.PublicKey
	cacheUntil time.Time
}

func NewSupabaseVerifier(jwksURL, issuer, audience string) *SupabaseVerifier {
	return &SupabaseVerifier{
		jwksURL:  jwksURL,
		issuer:   issuer,
		audience: audience,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		ecKeys: make(map[string]*ecdsa.PublicKey),
	}
}

func (v *SupabaseVerifier) Verify(ctx context.Context, token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("token mal formado")
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errors.New("header JWT inválido")
	}

	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
		Typ string `json:"typ"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, errors.New("header JWT inválido")
	}
	if header.Alg != "ES256" {
		return nil, errors.New("algoritmo JWT no soportado")
	}
	if strings.TrimSpace(header.Kid) == "" {
		return nil, errors.New("kid ausente en token")
	}

	signingInput := parts[0] + "." + parts[1]
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, errors.New("firma JWT inválida")
	}

	sum := sha256.Sum256([]byte(signingInput))
	key, err := v.getECKey(ctx, header.Kid)
	if err != nil {
		return nil, err
	}
	// JWS ES256 usa firma raw R||S de 64 bytes
	if len(signature) != 64 {
		return nil, errors.New("firma JWT inválida")
	}
	r := new(big.Int).SetBytes(signature[:32])
	s := new(big.Int).SetBytes(signature[32:])
	if !ecdsa.Verify(key, sum[:], r, s) {
		return nil, errors.New("firma JWT inválida")
	}

	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("claims JWT inválidos")
	}

	var raw struct {
		Sub   string          `json:"sub"`
		Role  string          `json:"role"`
		Email string          `json:"email"`
		Iss   string          `json:"iss"`
		Aud   json.RawMessage `json:"aud"`
		Exp   int64           `json:"exp"`
		Nbf   int64           `json:"nbf"`
	}
	if err := json.Unmarshal(claimsBytes, &raw); err != nil {
		return nil, errors.New("claims JWT inválidos")
	}

	now := time.Now().Unix()
	if raw.Exp == 0 || now >= raw.Exp {
		return nil, errors.New("token expirado")
	}
	if raw.Nbf != 0 && now < raw.Nbf {
		return nil, errors.New("token aún no válido")
	}
	if raw.Iss != v.issuer {
		return nil, errors.New("issuer inválido")
	}
	if !audienceMatches(raw.Aud, v.audience) {
		return nil, errors.New("audience inválida")
	}
	if strings.TrimSpace(raw.Sub) == "" {
		return nil, errors.New("subject inválido")
	}

	return &Claims{
		Subject: raw.Sub,
		Role:    raw.Role,
		Email:   raw.Email,
	}, nil
}

func audienceMatches(raw json.RawMessage, expected string) bool {
	if strings.TrimSpace(expected) == "" {
		return false
	}

	var audString string
	if err := json.Unmarshal(raw, &audString); err == nil {
		return audString == expected
	}

	var audSlice []string
	if err := json.Unmarshal(raw, &audSlice); err == nil {
		for _, a := range audSlice {
			if a == expected {
				return true
			}
		}
	}

	return false
}

func (v *SupabaseVerifier) getECKey(ctx context.Context, kid string) (*ecdsa.PublicKey, error) {
	v.mu.RLock()
	key, ok := v.ecKeys[kid]
	validCache := time.Now().Before(v.cacheUntil)
	v.mu.RUnlock()

	if ok && validCache {
		return key, nil
	}

	if err := v.refreshKeys(ctx); err != nil {
		return nil, err
	}

	v.mu.RLock()
	defer v.mu.RUnlock()
	key, ok = v.ecKeys[kid]
	if !ok {
		return nil, fmt.Errorf("no se encontró key EC para kid=%s", kid)
	}
	return key, nil
}

func (v *SupabaseVerifier) refreshKeys(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.jwksURL, nil)
	if err != nil {
		return err
	}

	resp, err := v.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error obteniendo JWKS: status=%d", resp.StatusCode)
	}

	var jwks struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			Crv string `json:"crv"`
			X   string `json:"x"`
			Y   string `json:"y"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return err
	}

	parsedEC := make(map[string]*ecdsa.PublicKey)
	for _, k := range jwks.Keys {
		if strings.TrimSpace(k.Kid) == "" || k.Kty != "EC" {
			continue
		}

		if k.Crv != "P-256" || k.X == "" || k.Y == "" {
			continue
		}

		xBytes, err := base64.RawURLEncoding.DecodeString(k.X)
		if err != nil {
			continue
		}
		yBytes, err := base64.RawURLEncoding.DecodeString(k.Y)
		if err != nil {
			continue
		}

		x := new(big.Int).SetBytes(xBytes)
		y := new(big.Int).SetBytes(yBytes)
		if !elliptic.P256().IsOnCurve(x, y) {
			continue
		}

		parsedEC[k.Kid] = &ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     x,
			Y:     y,
		}
	}

	if len(parsedEC) == 0 {
		return errors.New("JWKS sin claves válidas")
	}

	v.mu.Lock()
	v.ecKeys = parsedEC
	v.cacheUntil = time.Now().Add(10 * time.Minute)
	v.mu.Unlock()

	return nil
}
