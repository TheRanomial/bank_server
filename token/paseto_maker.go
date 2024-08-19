package token

import (
	"fmt"
	"time"
	"golang.org/x/crypto/chacha20poly1305"

	"github.com/o1egl/paseto"
)

type PasetoMaker struct {
	paseto 			*paseto.V2
	symmetricKey 	[]byte
}

func NewPasetoMaker(symmetrickey string) (Maker,error){
	if len(symmetrickey)<chacha20poly1305.KeySize{
		return nil,fmt.Errorf("key invalid: should be of minimum %d characters",chacha20poly1305.KeySize)
	}

	maker:=&PasetoMaker{
		paseto: paseto.NewV2(),
		symmetricKey: []byte(symmetrickey),
	}

	return maker,nil

}

func (maker *PasetoMaker) CreateToken(username string,duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(username,duration)
	if err != nil {
		return "", payload, err
	}

	token,err:=maker.paseto.Encrypt(maker.symmetricKey,payload,nil)
	return token, payload, err
}

func (maker *PasetoMaker) VerifyToken(token string) (*Payload, error) {
	p:=&Payload{}
	err:=maker.paseto.Decrypt(token,maker.symmetricKey,p,nil)

	if err!=nil{
		return nil,err
	}

	err=p.Valid()
	if err!=nil{
		return nil,err
	}

	return p,nil
}