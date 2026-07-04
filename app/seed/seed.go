package seed

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"

	"github.com/heidericklucas/go-security-swe-bench/app/auth"
	"github.com/heidericklucas/go-security-swe-bench/app/store"
)

// testRSAPrivPEM is a COMMITTED 2048-bit RSA private key
const testRSAPrivPEM = `-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEAyH5AhmI4tPF31Fe+b09A2TDwVMYM8HhOw/605AF4wuMSKm5N
+MMYa+2h4OWGHDZ1vP40Ug5ky2NV4B5Be0WEQAfHJLW3/4WljpDN2Fbl0cX2QPy9
yEEhnSrcXGRn4NcgWpJZImRjDlkf7ul5w4O6KWANtEcPo/O5jnXSkTf1fe5vGnNG
d9Jyzxl8X0O00A5TKsPDSXlqsuJwcGApzTYH8C1a/40imHZ7eYYev2KM9g4ThcUh
JV5+DUa7EfqM4WohYLDGN4PTobmI5wESEPkce8ZIq+aqZ/otITT/hsOzWuyJBzyr
u4Sv/2Zgq8jjnOwuZYMYZxQB4LUnkGX9DJnexQIDAQABAoIBADsderIE6Pp1DiN2
gah5QkIn21zrjmoi1vqUMcreojg4UqVfI69a+urrxKS2mE3eQuXoQA4Hv3F2xx3P
XfhWXXGxuWaaC/gT9GYuTPtiV9371CmCPAT9K0eXmSTG2Bgj5h6+cvigh9J1teQw
RB5BS1UixLeokjSByu71z5HQ4Znw+py+l/OtRBdEIIA7fk1mdOOK1wjQx57Ag9cA
OUu45jV8v8ylRqNf440PNXDd04iLPjXndcwh3cJZOtWTR96Tp420SlkDDJmRlt4o
9fnrq3wOc8CEvPOCa8BJsmNOSPYRIuEAufs1Exxf4pQ4lYYw4mOAh/0q6/A0hQXV
wWkkmNsCgYEA/shCVcCOlAfDd0HUuRXeEtSt9Q4jMJBSrOPr1BKBLAVnVK29FFKp
ZxrAL2v8yvh0IY/GDjo1bkxvrcS2hxiRZjthtVYxJKiL2cQgoSaSL9FX+d8t+Kff
UlplRszxhbGx/PYPN9+Vk7qQjaYFq1WHuma0n+WwjGsFRlORNhIIBNMCgYEAyXOR
LAmQ8Ab71HMrNkxzbb6EXeK8YlloMMXFhaRQiC8RcUKA5q9vQeQ0bJXyJI09ivRx
r/Q06dWWNOV6es6b9c9uFD9wYK6aL+J/8KK8AwMnhN9NVP2NJjcaZ9AdhFGmlhSx
iwtQr2qZXEqeFQWhRH01N3RcND6CWYM4qjElLwcCgYBPeHOIf+l5Lvq/Rh9uI+4C
/afNGj3LthizqNw0aBk2e/EBLrgdkLMaX/O2Vv6g6OKAXXIvmeR0pQ7oqsUsWnNv
6fHOODZX8uK8aDqtSXSryaahYAXc27AC5gNVFDP5ubWE69NPYEQtsjHInRKDoMgX
UlXg0ipBglhA2Wwf63vo1QKBgGterzdbeeaUslBPee6c3MDXVFF98Y4xvPiR0G8k
Xs/W1mMZYP6S7qed69scKE0XAoA3PQjdDf67mi1jSd5+lF11S2R7U3gUV5XCERK+
iZrJaZzGL79mzH5bzDUclT7bDgDb0q0bQMvd8xNfqdXmhEeuvFhsKmCmGG+BIwiC
3SCtAoGAC/3ckgyQUCCSVFeoTxA2sVwyYocvfL9EohMQ/kzPBPe0b4WaObAJM6rm
Us+VxjexYSbl707dNp8gApPP+wHI7+I896PvaUiMUeAaKwEFpwb9PAKr7k1/wm3D
PQc+zOuYdn++yTsVQWeNKh5Zw+k6ZnVRJpNiX0TvLlguOnVdWaE=
-----END RSA PRIVATE KEY-----`

func FixedRSAKeys() auth.Keys {
	block, _ := pem.Decode([]byte(testRSAPrivPEM))
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		key, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			panic("seed: bad test RSA key: " + err.Error())
		}
		priv = key.(*rsa.PrivateKey)
	}
	return auth.Keys{RSAPrivate: priv, RSAPublic: &priv.PublicKey}
}

type Data struct {
	AliceID, BobID, CarolID int64
	BobPrivateNoteID        int64
	SentinelToken           string
}

func Load(s *store.Store, filesDir string) (Data, error) {
	alice, _ := s.CreateUser(store.User{Username: "alice", PasswordHash: "alice-pw"})
	bob, _ := s.CreateUser(store.User{Username: "bob", PasswordHash: "bob-pw"})
	carol, _ := s.CreateUser(store.User{Username: "carol", PasswordHash: "carol-pw"})
	_, _ = s.CreateNote(store.Note{OwnerID: alice, Title: "groceries", Body: "milk, eggs", Private: false})
	bobNote, _ := s.CreateNote(store.Note{OwnerID: bob, Title: "diary", Body: "BOB_SECRET_NOTE", Private: true})

	if err := os.MkdirAll(filesDir, 0o755); err != nil {
		return Data{}, err
	}
	if err := os.WriteFile(filepath.Join(filesDir, "avatar.png"), []byte("PNGDATA"), 0o600); err != nil {
		return Data{}, err
	}
	return Data{AliceID: alice, BobID: bob, CarolID: carol, BobPrivateNoteID: bobNote, SentinelToken: ""}, nil
}
