package plugin

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"math/big"
	"strconv"
	"strings"
	"time"
)

var pluginPubKey = ""

func SetPluginPubKey(pubKey string) {
	pluginPubKey = pubKey
}

// 插件判断授权
func HasAuth(code string) error {
	auth, err := verifiedPluginAuth(code)
	if err != nil {
		return err
	}
	switch auth.Type {
	case "expired":
		expired, err := strconv.ParseInt(auth.Value, 10, 64)
		if err != nil || expired <= 0 || time.Now().Unix() > expired {
			return errors.New("official license unauthorized")
		}
	case "has":
		if auth.Value != "1" {
			return errors.New("official license unauthorized")
		}
	default:
		return errors.New("official license unauthorized")
	}
	return nil
}

// 插件返回授权次数
func HasAuthByTimes(code string) (int64, error) {
	auth, err := verifiedPluginAuth(code)
	if err != nil {
		return 0, err
	}
	if auth.Type != "times" {
		return 0, errors.New("official license unauthorized")
	}
	times, err := strconv.ParseInt(auth.Value, 10, 64)
	if err != nil || times <= 0 {
		return 0, errors.New("official license unauthorized")
	}
	return times, nil
}

type pluginAuthCipherResult struct {
	AuthID     string
	AuthCipher string
}

type pluginAuthPlain struct {
	AuthID string
	Code   string
	Type   string
	Value  string
}

func pluginAuthCipher(code string) (*pluginAuthCipherResult, error) {
	res, err := officialLicensePostData("/official_license/auth_cipher", map[string]interface{}{
		"code": code,
	})
	if err != nil {
		return nil, err
	}
	authCipher, ok := res["authCipher"].(string)
	if !ok || authCipher == "" {
		return nil, errors.New("official license auth cipher not found")
	}
	authID, ok := res["authId"].(string)
	if !ok || authID == "" {
		return nil, errors.New("official license auth id not found")
	}
	return &pluginAuthCipherResult{AuthID: authID, AuthCipher: authCipher}, nil
}

func verifiedPluginAuth(code string) (*pluginAuthPlain, error) {
	res, err := pluginAuthCipher(code)
	if err != nil {
		return nil, err
	}
	plain, err := decryptPluginAuthCipher(res.AuthCipher)
	if err != nil {
		return nil, err
	}
	if plain.AuthID != res.AuthID || plain.Code != code {
		return nil, errors.New("official license unauthorized")
	}
	return plain, nil
}

func decryptPluginAuthCipher(cipherText string) (*pluginAuthPlain, error) {
	pub, err := parsePluginPublicKey(pluginPubKey)
	if err != nil {
		return nil, err
	}
	data, err := publicDecryptBase64(pub, cipherText)
	if err != nil {
		return nil, err
	}
	plain := &pluginAuthPlain{}
	for _, line := range strings.Split(string(data), "\n") {
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch key {
		case "auth_id":
			plain.AuthID = value
		case "code":
			plain.Code = value
		case "type":
			plain.Type = value
		case "value":
			plain.Value = value
		}
	}
	if plain.AuthID == "" || plain.Code == "" || plain.Type == "" || plain.Value == "" {
		return nil, errors.New("official license auth cipher invalid")
	}
	return plain, nil
}

func parsePluginPublicKey(src string) (*rsa.PublicKey, error) {
	src = strings.TrimSpace(src)
	if src == "" {
		return nil, errors.New("plugin public key is empty")
	}
	block, _ := pem.Decode([]byte(src))
	if block == nil {
		return nil, errors.New("plugin public key invalid")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	key, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("plugin public key invalid")
	}
	return key, nil
}

func publicDecryptBase64(pub *rsa.PublicKey, cipherText string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return nil, err
	}
	return publicDecryptPKCS1v15(pub, data)
}

func publicDecryptPKCS1v15(pub *rsa.PublicKey, cipherText []byte) ([]byte, error) {
	k := pub.Size()
	if len(cipherText) != k {
		return nil, errors.New("official license auth cipher invalid")
	}
	c := new(big.Int).SetBytes(cipherText)
	if c.Cmp(pub.N) > 0 {
		return nil, errors.New("official license auth cipher invalid")
	}
	m := new(big.Int).Exp(c, big.NewInt(int64(pub.E)), pub.N)
	em := m.FillBytes(make([]byte, k))
	if len(em) < 11 || em[0] != 0 || em[1] != 1 {
		return nil, errors.New("official license auth cipher invalid")
	}
	i := 2
	for ; i < len(em); i++ {
		if em[i] == 0 {
			break
		}
		if em[i] != 0xff {
			return nil, errors.New("official license auth cipher invalid")
		}
	}
	if i < 10 || i >= len(em) {
		return nil, errors.New("official license auth cipher invalid")
	}
	return em[i+1:], nil
}

func officialLicensePostData(uri string, data map[string]interface{}) (map[string]interface{}, error) {
	res, err := httpPost(uri, data)
	if err != nil {
		return nil, err
	}
	if code, ok := res["code"].(float64); ok && int(code) == 0 {
		data, ok := res["data"].(map[string]interface{})
		if !ok {
			return nil, errors.New("official license response data not found")
		}
		return data, nil
	}
	if msg, ok := res["msg"].(string); ok && msg != "" {
		return nil, errors.New(msg)
	}
	return nil, errors.New("official license unauthorized")
}
