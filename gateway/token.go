package gateway

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	jwtReq "github.com/dgrijalva/jwt-go/request"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"net/http"
)

// Given request with JWT in Bearer token, tenant container and JWT secret value,
// fill tenant container with JWT claim data and validate JWT signature.
func jwtToTenant(r *http.Request, te *tenant, logger log.Logger, algo string, jwtSecretValue interface{}) error {

	// HMAC validation bits
	var validateJwtHmac = func(token *jwt.Token) (interface{}, error) {
		// Only HMAC algorithms accepted - algorithm validation is super important!
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			level.Info(logger).Log("msg", "unexpected signing method", "used_method", token.Header["alg"])
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return jwtSecretValue, nil
	}

	// JWT validation bits
	var validateJwtRsa = func(token *jwt.Token) (interface{}, error) {
		// Only HMAC algorithms accepted - algorithm validation is super important!
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			level.Info(logger).Log("msg", "unexpected signing method", "used_method", token.Header["alg"])
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return jwtSecretValue, nil
	}
	var validator func(token *jwt.Token)(interface{}, error)

	// Validate JWT
	if algo == "hmac" {
		validator = validateJwtHmac
	} else {
		if algo == "rsa" {
			validator = validateJwtRsa
		} else {
			level.Warn(logger).Log("msg", "invalid validator algo")
			validator =  func(token *jwt.Token)(interface{}, error) {
				return nil, fmt.Errorf("Unknown signing method: %v", jwtValidationAlgo)
			}
		}
	}

	_, err := jwtReq.ParseFromRequest(
		r,
		jwtReq.AuthorizationHeaderExtractor,
		validator,
		jwtReq.WithClaims(te))

	return err

}
