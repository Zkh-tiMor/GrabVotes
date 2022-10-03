package jwt

import (
	"github.com/dgrijalva/jwt-go"
	"log"
	"time"
)

const (
	TokenExpireDuration = time.Hour * 24 * 365
)

var mySecret = []byte("GrabVotes")

func keyFunc(token *jwt.Token) (interface{}, error) {
	return mySecret, nil
}

type MyClaims struct {
	UserID             string `json:"id"`
	jwt.StandardClaims        //内嵌匿名字段，但如果给这个结构体加上名称的话表示内嵌的是一个独立的结构体
	//匿名结构体实现继承的效果，有名字的结构体实现组合效果。如果内嵌多个匿名结构体则可实现类似于多重继承的效果
}

// ParseToken 解析token返回包含用户消息的结构体
func ParseToken(tokenString string) (*MyClaims, bool, error) {
	// 解析token
	var claims = new(MyClaims) //解析好的存放在mc中

	//keyFunc 用来根据token值返回密钥
	token, err := jwt.ParseWithClaims(tokenString, claims, keyFunc)
	if err != nil { //解析过程出错
		log.Fatalln("parse token failed: ", err.Error())
		return nil, false, err
	}
	if token.Valid { // 校验token是否合法
		return claims, true, nil
	}
	return nil, false, nil
}
