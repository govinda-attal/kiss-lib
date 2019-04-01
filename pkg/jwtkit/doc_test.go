package jwtkit_test

import (
	"fmt"

	"github.com/govinda-attal/kiss-lib/pkg/jwtkit"
)

func ExampleVerifier_rsaVerifier() {
	var token string // Set this to a JWT bearer token.
	authCrtFile := "authsvc.crt"
	v, err := jwtkit.NewRSAVerifier(authCrtFile)
	if err != nil {
		return
	}

	c, err := v.VerifyToken(token)
	if err != nil {
		return
	}
	fmt.Println("Claims: ", c)
}
