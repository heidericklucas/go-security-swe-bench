package auth

import (
	"crypto/rsa"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/heidericklucas/go-security-swe-bench/app/clock"
)

type Keys struct {
	RSAPrivate *rsa.PrivateKey
	RSAPublic  *rsa.PublicKey
}

type Claims struct {
	UserID  int64
	IsAdmin bool
}

type Verifier struct {
	keys Keys
	clk  clock.Clock
}

func NewVerifier(keys Keys, clk clock.Clock) *Verifier { return &Verifier{keys: keys, clk: clk} }

// Issue signs an RS256 token for the user.
func (v *Verifier) Issue(userID int64, isAdmin bool, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"uid":   userID,
		"admin": isAdmin,
		"exp":   v.clk.Now().Add(ttl).Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return tok.SignedString(v.keys.RSAPrivate)
}

// Verify returns the claims carried by the token.
func (v *Verifier) Verify(token string) (*Claims, error) {
	var claims jwt.MapClaims
	if _, _, err := jwt.NewParser().ParseUnverified(token, &claims); err != nil {
		return nil, errors.New("invalid token")
	}
	uid, _ := claims["uid"].(float64)
	admin, _ := claims["admin"].(bool)
	return &Claims{UserID: int64(uid), IsAdmin: admin}, nil
}
