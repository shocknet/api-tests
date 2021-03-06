package main

import (
	"crypto/rsa"
)

type Settings struct {
	Image           string    `json:"image"`
	Nodes           []Node    `json:"nodes"`
	Network         string    `json:"network"`
	NewUserEndpoint string    `json:"newUserEndpoint"`
	TestServerPort  string    `json:"testServerPort"`
	Gunk1           GunK1     `json:"gunk1"`
	Superpeer       Superpeer `json:"superpeer"`
}
type Node struct {
	Name      string `json:"name"`
	Port      string `json:"port"`
	Addr      string `json:"addr"`
	TLS       string `json:"tls"`
	Macaroon  string `json:"macaroon"`
	AliasBase string `json:"aliasBase"`
	Host      string `json:"host"`
}
type GunK1 struct {
	Image string `json:"image"`
	Port  string `json:"port"`
	Name  string `json:"name"`
}

type Superpeer struct {
	Image string `json:"image"`
	Port  string `json:"port"`
	Name  string `json:"name"`
}

type EncryptedData struct {
	Data string `json:"encryptedData"`
	Key  string `json:"encryptedKey"`
	IV   string `json:"iv"`
}

type TestInfo struct {
	DeviceID         string
	RSAPrivKey       *rsa.PrivateKey
	Address          string
	Token            string
	Alias            string
	Pass             string
	APIPubKey        string
	GunPubKey        string
	HandshakeAddress string
	HandshakeReqID   string
}

type APIRoutes struct {
	auth                 string
	exchangeKeys         string
	getDisplayName       string
	getHandshakeRequests string
	setDisplayName       string
	generateHSNode       string
	generateOrderAddress string
	initFeedWall         string
	sendHandshake        string
	getHandshakeAddress  string
	acceptHSRequest      string
	gunRpcPut            string
	gunKeys              GunKeys
}

type GunKeys struct {
	setDisplayName   string
	setHandshakeNode string
	setOrderAddress  string
}
type ErrorsWrapper struct {
	Err          string `json:"err"`
	ErrorMessage string `json:"errorMessage"`
}

type NewUserReq struct {
	Alias string `json:"alias"`
	Pass  string `json:"pass"`
}
type NewUserRes struct {
	Ok  bool   `json:"ok"`
	Err string `json:"err"`
}
type AuthReq struct {
	Alias string `json:"alias"`
	Pass  string `json:"password"`
}
type AuthRes struct {
	Token string      `json:"authorization"`
	User  AuthResUser `json:"user"`
}
type AuthResUser struct {
	Alias     string `json:"alias"`
	APIPubKey string `json:"publicKey"`
}

type ExchangeKeysReq struct {
	PublicKey string `json:"publicKey"`
	DeviceId  string `json:"deviceId"`
}

type ExchangeKeysRes struct {
	APIPublicKey string `json:"APIPublicKey"`
}

type TestServer struct {
	Infos map[string]TestInfo
}

type EncryptedDataHttp struct {
	Alias string `json:"alias"`
	Data  string `json:"encryptedData"`
}

type DecryptedDataHttp struct {
	Alias string `json:"alias"`
	Data  string `json:"decryptedData"`
}

type OkRes struct {
	Ok bool `json:"ok"`
}
type DataRes struct {
	Data string `json:"data"`
}
type DataArrRes struct {
	Data []interface{} `json:"data"`
}
type TokenRes struct {
	Token string `json:"token"`
}

type HandshakeRequestsRes struct {
	ID                   string  `json:"id"`
	RequestorDisplayName string  `json:"requestorDisplayName"`
	RequestorPK          string  `json:"requestorPK"`
	Timestamp            float64 `json:"timestamp"`
}
type ActionSetDisplayName struct {
	Token       string `json:"token"`
	DisplayName string `json:"displayName"`
}

type ActionSendHandshake struct {
	Token           string `json:"token"`
	RecipientPubKey string `json:"recipientPublicKey"`
	UuId            string `json:"uuid"`
}
type reqSendHandshake struct {
	PublicKey string `json:"publicKey"`
}

type ActionAcceptRequest struct {
	Token     string `json:"token"`
	RequestID string `json:"requestID"`
}

type RPCAction struct {
	Path  string `json:"path"`
	Value string `json:"value"`
}
