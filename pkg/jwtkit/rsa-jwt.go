package jwtkit

import (
	"crypto/rsa"
	"crypto/x509"
	"io/ioutil"
	"log"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/pkcs12"

	"github.com/govinda-attal/kiss-lib/pkg/core/status"
)

// NewRSAVerifier returns rsa verifier which takes public certificate file.
// It be used to verify JWT tokens signed by its counter private key file.
func NewRSAVerifier(pubKeyPath string) (*rsaVerifier, error) {
	verifyBytes, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		return nil, status.ErrInternal().WithError(err)
	}
	verifyKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		return nil, status.ErrInternal().WithError(err)
	}
	v := &rsaVerifier{
		rsaPubKey: verifyKey,
	}
	return v, nil
}

// VerifyToken verifies the JWT token validity and returns claims for the valid token.
func (rv *rsaVerifier) VerifyToken(tokenValue string) (jwt.Claims, error) {

	token, err := jwt.Parse(tokenValue, func(token *jwt.Token) (interface{}, error) {

		return rv.rsaPubKey, nil
	})
	if err != nil {
		log.Println(err)
		return nil, status.ErrUnauthorized().WithError(err)
	}
	if !token.Valid {
		return nil, ErrTokenInValid()
	}

	return token.Claims, nil
}

func NewRSASigner(pkcs12FilePath, pwd, alg string) (*rsaSigner, error) {
	data, err := ioutil.ReadFile(pkcs12FilePath)
	if err != nil {
		return nil, status.ErrInternal().WithError(err)
	}
	priv, crt, err := pkcs12.Decode(data, pwd)
	if err := priv.(*rsa.PrivateKey).Validate(); err != nil {
		return nil, status.ErrInternal().WithError(err)
	}
	s := &rsaSigner{
		rsaPrvKey: priv.(*rsa.PrivateKey),
		crt:       crt,
		signMthd:  jwt.GetSigningMethod(alg),
	}
	return s, nil
}

func (rs *rsaSigner) SignBasic(subject string, expMins time.Duration) (string, error) {

	t := time.Now()
	claims := &jwt.StandardClaims{
		Issuer:    rs.crt.Subject.String(),
		ExpiresAt: t.Add(time.Minute * expMins).Unix(),
		IssuedAt:  t.Unix(),
		Subject:   subject,
	}
	token := jwt.NewWithClaims(rs.signMthd, claims)
	ts, err := token.SignedString(rs.rsaPrvKey)
	if err != nil {
		return "", status.ErrInternal().WithError(err)
	}
	return ts, nil
}

// Verifier can be used to verify JWT bearer token.
type Verifier interface {
	VerifyToken(tokenValue string) (jwt.Claims, error)
}

type rsaVerifier struct {
	rsaPubKey *rsa.PublicKey
}

type Signer interface {
	SignBasic(subject string, expMins time.Duration) (string, error)
}

type rsaSigner struct {
	rsaPrvKey *rsa.PrivateKey
	crt       *x509.Certificate
	signMthd  jwt.SigningMethod
}
