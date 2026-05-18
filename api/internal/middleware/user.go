package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/fs185085781/v9os/internal/ioc/uioc"
	"github.com/fs185085781/v9os/internal/model/user"

	"github.com/fs185085781/v9os/internal/config"
	"github.com/fs185085781/v9os/internal/logger"
	"github.com/fs185085781/v9os/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/cast"
)

// 传递用户信息 中间件
type User struct {
	cfg *config.AuthConfig
}

// NewCORS 构造函数
func NewUser(cfg *config.AuthConfig, log logger.Logger) *User {
	res := &User{cfg: cfg}
	log.Println("[用户信息中间件]已初始化")
	return res
}

// Middleware 生成Gin中间件
func (a *User) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := a.getClaims(c)
		if claims == nil {
			c.Next()
			return
		}
		c.Set("userID", claims["userID"])
		c.Set("deptID", claims["deptID"])
		c.Set("tokenID", claims["tokenID"])
		c.Next()
	}
}
func (a *User) getClaims(c *gin.Context) jwt.MapClaims {
	claims := a.getClaimsHeader(c)
	if claims == nil {
		claims = a.getClaimsQuery(c)
	}
	return claims
}
func (a *User) getClaimsHeader(c *gin.Context) jwt.MapClaims {
	return a.getClaimsByToken(c, a.extractToken(c))
}
func (a *User) getClaimsQuery(c *gin.Context) jwt.MapClaims {
	return a.getClaimsByToken(c, c.Query("token"))
}

func (a *User) getClaimsByToken(c *gin.Context, token string) jwt.MapClaims {
	if token == "" {
		return nil
	}
	claims, isLast, err := a.parseToken(token)
	if err != nil {
		return nil
	}
	if isLast {
		c.Header("TokenLast", "true")
	}
	return claims
}

// GenerateToken 生成JWT令牌
func (a *User) GenerateToken(userID string, deptID string) (string, error) {
	claims := jwt.MapClaims{
		"tokenID": a.getTokenIdByUserId(userID),
		"userID":  userID,
		"deptID":  deptID,
		"exp":     time.Now().Add(a.cfg.ExpireDuration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(a.cfg.Secret))
	if err != nil {
		return "", err
	}
	cc := uioc.Cache()
	uid := util.UUID()
	err = cc.SetValue("login:rtoken:"+uid, []byte(userID+"|"+deptID), a.cfg.RefreshExpireDuration)
	if err != nil {
		return "", err
	}
	a.AddResetRemoveRToken(userID, 1, uid)
	return tokenStr + "|" + uid, nil
}
func (a *User) AddResetRemoveRToken(userID string, dType int, rToken string) {
	//dType 1:Add 2:Reset 3:Remove
	cc := uioc.Cache()
	obj := cc.CreateLock("middleware:user:arrrt:" + userID)
	obj.Lock()
	defer obj.UnLock()
	var userRTS []string
	cc.GetObjectRetry("login:utoken:"+userID, &userRTS)
	if userRTS == nil {
		userRTS = make([]string, 0)
	}
	if dType == 1 {
		userRTS = append(userRTS, rToken)
	} else if dType == 2 {
		for _, v := range userRTS {
			if v == rToken {
				continue
			}
			cc.RemoveValue("login:rtoken:" + v)
		}
		userRTS = []string{rToken}
	} else if dType == 3 {
		tmpStrs := make([]string, 0)
		for _, v := range userRTS {
			if v == rToken {
				cc.RemoveValue("login:rtoken:" + v)
			} else {
				tmpStrs = append(tmpStrs, v)
			}
		}
		userRTS = tmpStrs
	} else {
		return
	}
	cc.SetObjectRetry("login:utoken:"+userID, userRTS, a.cfg.RefreshExpireDuration)
}

func (a *User) getTokenIdByUserId(userID string) string {
	var md5str string
	cc := uioc.Cache()
	cc.MemCacheObject("login:usercache:"+userID, &md5str, func() interface{} {
		var u user.User
		uioc.Database().GetByID(cast.ToUint(userID), &u)
		var str string
		if u.ID > 0 {
			str = util.MD5Lower(u.Password + u.Otp + u.Phone + u.Email)
		} else {
			str = util.MD5Lower(userID)
		}
		return str
	}, 2*time.Minute)
	return md5str
}
func (a *User) RefreshToken(c *gin.Context, token string) (string, error) {
	cc := uioc.Cache()
	data, err := cc.GetValue("login:rtoken:" + token)
	if err != nil {
		return "", fmt.Errorf("refresh token invalid")
	}
	data2 := strings.Split(string(data), "|")
	userID := data2[0]
	deptID := data2[1]
	if userID == "" {
		return "", fmt.Errorf("user id invalid")
	}
	a.AddResetRemoveRToken(userID, 3, token)
	return a.GenerateToken(userID, deptID)
}

// parseToken 解析JWT令牌
func (a *User) parseToken(tokenString string) (jwt.MapClaims, bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.cfg.Secret), nil
	})
	//兼容15天内的密钥
	isLast := false
	if err != nil && a.cfg.LastSecret != "" {
		token, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(a.cfg.LastSecret), nil
		})
		isLast = true
	}
	var claims jwt.MapClaims
	if token != nil && token.Claims != nil {
		claims = token.Claims.(jwt.MapClaims)
	}
	if err == nil && token.Valid {
		if claims["tokenID"].(string) != a.getTokenIdByUserId(claims["userID"].(string)) {
			return nil, false, fmt.Errorf("token id invalid")
		}
		if int64(claims["exp"].(float64))-util.UnixSeconds() < 1800 {
			isLast = true
		}
		return claims, isLast, nil
	}
	if err == nil {
		err = fmt.Errorf("token valid false")
	}
	return claims, false, err
}

// extractToken 从请求头提取Token
func (a *User) extractToken(c *gin.Context) string {
	bearerToken := c.GetHeader("Authorization")
	if bearerToken == "" {
		return ""
	}
	if len(bearerToken) > 7 && strings.ToUpper(bearerToken[0:6]) == "BEARER" {
		return bearerToken[7:]
	}
	return bearerToken
}
