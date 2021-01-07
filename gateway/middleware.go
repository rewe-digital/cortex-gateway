package gateway

import (
	"crypto/rsa"
	"encoding/base64"
	"flag"
	"net/http"
	"strings"

	"github.com/cortexproject/cortex/pkg/util"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/weaveworks/common/middleware"
)

var (
	jwtSecret    string
	jwtSecretEncoded bool
	jwtValidationAlgo string
	authFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "cortex_gateway",
		Name:      "failed_authentications_total",
		Help:      "The total number of failed authentications.",
	}, []string{"reason"})
	authSuccess = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "cortex_gateway",
		Name:      "succeeded_authentications_total",
		Help:      "The total number of succeeded authentications.",
	}, []string{"tenant"})
)

func init() {
	flag.StringVar(&jwtSecret, "gateway.auth.jwt-secret", "", "Secret to sign JSON Web Tokens")
	flag.BoolVar(&jwtSecretEncoded, "gateway.auth.jwt-secret-encoded", false, "Whether to base64-decode secret before use")
	flag.StringVar(&jwtValidationAlgo, "gateway.auth.token-validation.algo", "HMAC", "JWT Validation algorithm: HMAC or RSA")
}


// AuthenticateTenant validates the Bearer Token and attaches the TenantID to the request
var AuthenticateTenant = middleware.Func(func(next http.Handler) http.Handler {

	// Initialize JWT secret
	jwtSecretBytes := []byte(jwtSecret)
	var jwtRsaKey rsa.PublicKey
	var jwtRsaKeyPtr *rsa.PublicKey
	var err error

	// Optionally decode JWT secret
	if jwtSecretEncoded {
		secretDecoded, err := base64.StdEncoding.DecodeString(jwtSecret)
		if err != nil {
			logger := log.With(util.Logger)
			level.Warn(logger).Log("msg", "base64 secret encoding is " +
				"enabled, but secret cant decode", "secret_encoding", "base64")
		} else {
			jwtSecretBytes = secretDecoded
		}
	}

	// Optionally parse RSA key
	var algoLower = strings.ToLower(jwtValidationAlgo)
	if algoLower == "rsa" {
		jwtRsaKeyPtr, err = jwt.ParseRSAPublicKeyFromPEM(jwtSecretBytes)
		if err != nil {
			logger := log.With(util.Logger)
			level.Warn(logger).Log("msg", "can not load rsa key ", "secret_encoding", "pem")
		} else {
			jwtRsaKey = *jwtRsaKeyPtr
		}
	}


	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := log.With(util.WithContext(r.Context(), util.Logger), "ip_address", r.RemoteAddr)
		level.Debug(logger).Log("msg", "authenticating request", "route", r.RequestURI)

		tokenString := r.Header.Get("Authorization") // Get operation is case insensitive
		if tokenString == "" {
			level.Info(logger).Log("msg", "no bearer token provided")
			http.Error(w, "No bearer token provided", http.StatusUnauthorized)
			authFailures.WithLabelValues("no_token").Inc()
			return
		}

		// Try to parse and validate JWT
		te := &tenant{}

		if algoLower == "rsa" {
			err = jwtToTenant(r, te, logger, algoLower, &jwtRsaKey)
		} else {
			err = jwtToTenant(r, te, logger, algoLower, jwtSecretBytes)
		}

		// If Tenant's Valid method returns false an error will be set as well, hence there is no need
		// to additionally check the parsed token for "Valid"
		if err != nil {
			level.Info(logger).Log("msg", "invalid bearer token", "err", err.Error())
			http.Error(w, "Invalid bearer token", http.StatusUnauthorized)
			authFailures.WithLabelValues("token_not_valid").Inc()
			return
		}

		// Token is valid
		authSuccess.WithLabelValues(te.TenantID).Inc()
		r.Header.Set("X-Scope-OrgID", te.TenantID)
		next.ServeHTTP(w, r)
	})
})
