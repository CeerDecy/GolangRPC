package token

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"github/CeerDecy/RpcFrameWork/crpc"
	"time"
)

const CrpcToken = "crpc_token"

type JwtHandler struct {
	Alg            string           // 算法方式
	TimeOut        time.Duration    // 过期时间
	RefreshTimeOut time.Duration    // 过期时间
	TimeFunc       func() time.Time // 时间函数
	Key            []byte
	PrivateKey     string
	RefreshKey     string
	SendCookie     bool
	Authenticator  func(ctx *crpc.Context) (map[string]any, error)
	CookieName     string
	CookieMaxAge   int64
	CookieDomain   string
	CookieSecure   bool
	CookieHttpOnly bool
}
type JwtResponse struct {
	Toke         string `json:"toke"`
	RefreshToken string `json:"refreshToken"`
}

// LoginHandler 登录处理
func (j *JwtHandler) LoginHandler(ctx *crpc.Context) (*JwtResponse, error) {
	data, err := j.Authenticator(ctx)
	if err != nil {
		return nil, err
	}
	if j.Alg == "" {
		j.Alg = "HS256"
	}
	// A部分
	method := jwt.GetSigningMethod(j.Alg)
	token := jwt.New(method)
	// B部分
	claims := token.Claims.(jwt.MapClaims)
	if data != nil {
		for k, v := range data {
			claims[k] = v
		}
	}
	if j.TimeFunc == nil {
		j.TimeFunc = func() time.Time {
			return time.Now()
		}
	}
	expire := j.TimeFunc().Add(j.TimeOut)
	// 设置过期时间
	claims["exp"] = expire.Unix()
	// 设置发布时间
	claims["iat"] = j.TimeFunc().Unix()
	// C部分
	var signedString string
	if j.usingPublicKeyAlgo() {
		signedString, err = token.SignedString(j.PrivateKey)
	} else {
		signedString, err = token.SignedString(j.Key)
	}
	if err != nil {
		return nil, err
	}
	response := &JwtResponse{
		Toke: signedString,
	}
	// refreshToken
	refreshToken, err := j.refreshToken(token)
	if err != nil {
		return nil, err
	}
	response.RefreshToken = refreshToken
	// 判断是否需要设置Cookie
	if j.SendCookie {
		if j.CookieName == "" {

			j.CookieName = CrpcToken
		}
		if j.CookieMaxAge == 0 {
			j.CookieMaxAge = expire.Unix() - j.TimeFunc().Unix()
		}
		ctx.SetCookie(
			j.CookieName,
			signedString,
			"/",
			j.CookieDomain,
			int(j.CookieMaxAge),
			j.CookieSecure,
			j.CookieHttpOnly,
		)
	}
	return response, nil
}

func (j *JwtHandler) usingPublicKeyAlgo() bool {
	switch j.Alg {
	case "RS256", "RS512", "RS238":
		return true
	default:
		return false
	}
}

func (j *JwtHandler) refreshToken(token *jwt.Token) (string, error) {
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = j.TimeFunc().Add(j.RefreshTimeOut).Unix()
	var signedString string
	var err error
	if j.usingPublicKeyAlgo() {
		signedString, err = token.SignedString(j.PrivateKey)
	} else {
		signedString, err = token.SignedString(j.Key)
	}
	if err != nil {
		return "", err
	}
	return signedString, nil
}

// LogoutHandler 登出函数处理
func (j *JwtHandler) LogoutHandler(ctx *crpc.Context) error {
	if j.SendCookie {
		if j.CookieName == "" {
			j.CookieName = CrpcToken
		}
		ctx.SetCookie(
			j.CookieName, "", "/", j.CookieDomain, -1, j.CookieSecure, j.CookieHttpOnly,
		)
	}
	return nil
}

func (j *JwtHandler) RefreshHandler(ctx *crpc.Context) (*JwtResponse, error) {
	refresh, ok := ctx.Get(j.RefreshKey)
	if !ok {
		return nil, errors.New("refresh token is null")
	}
	if j.Alg == "" {
		j.Alg = "HS256"
	}
	// 解析token
	parse, err := jwt.Parse(refresh.(string), func(token *jwt.Token) (interface{}, error) {
		if j.usingPublicKeyAlgo() {
			return j.PrivateKey, nil
		} else {
			return j.Key, nil
		}
	})
	if err != nil {
		return nil, err
	}

	claims := parse.Claims.(jwt.MapClaims)
	if j.TimeFunc == nil {
		j.TimeFunc = func() time.Time {
			return time.Now()
		}
	}
	expire := j.TimeFunc().Add(j.TimeOut)
	// 设置过期时间
	claims["exp"] = expire.Unix()
	// 设置发布时间
	claims["iat"] = j.TimeFunc().Unix()
	// C部分
	var signedString string
	if j.usingPublicKeyAlgo() {
		signedString, err = parse.SignedString(j.PrivateKey)
	} else {
		signedString, err = parse.SignedString(j.Key)
	}
	if err != nil {
		return nil, err
	}
	response := &JwtResponse{
		Toke: signedString,
	}
	// refreshToken
	refreshToken, err := j.refreshToken(parse)
	if err != nil {
		return nil, err
	}
	response.RefreshToken = refreshToken
	// 判断是否需要设置Cookie
	if j.SendCookie {
		if j.CookieName == "" {

			j.CookieName = CrpcToken
		}
		if j.CookieMaxAge == 0 {
			j.CookieMaxAge = expire.Unix() - j.TimeFunc().Unix()
		}
		ctx.SetCookie(
			j.CookieName,
			signedString,
			"/",
			j.CookieDomain,
			int(j.CookieMaxAge),
			j.CookieSecure,
			j.CookieHttpOnly,
		)
	}
	return response, nil
}
