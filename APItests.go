package main

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/boufni95/goutils"
)

func CreateGunUser(r *rand.Rand, settings Settings, TestInfos []TestInfo, i int, v Node) bool {

	signIn, err := GenerateAuthPair(v.AliasBase, r)
	signInBytes, err := goutils.ToJsonBytes(signIn)
	res, err := http.Post(settings.NewUserEndpoint, "application/json", bytes.NewBuffer(signInBytes))
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	resNewUser := NewUserRes{}
	err = goutils.JsonBytesToType(bodyBytes, &resNewUser)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	if !resNewUser.Ok {
		goutils.PrintError(errors.New(resNewUser.Err))
		return false
	}
	goutils.Log("Successfully created user " + signIn.Alias + " with pass " + signIn.Pass + " *not with api")
	dID, err := generateDeviceID(r)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	TestInfos[i].DeviceID = dID
	TestInfos[i].Alias = signIn.Alias
	TestInfos[i].Pass = signIn.Pass
	return true
}

func ExchangeKeys(r *rand.Rand, settings Settings, TestInfos []TestInfo, routes APIRoutes, i int, v Node) bool {
	priv, err := rsa.GenerateKey(r, 2048)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	TestInfos[i].RSAPrivKey = priv

	pubASN := x509.MarshalPKCS1PublicKey(&TestInfos[i].RSAPrivKey.PublicKey)
	pemdata := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pubASN,
		},
	)
	bodyEx := ExchangeKeysReq{
		DeviceId:  TestInfos[i].DeviceID,
		PublicKey: string(pemdata),
	}
	bodyExBytes, err := goutils.ToJsonBytes(bodyEx)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", TestInfos[i].Address+routes.exchangeKeys, bytes.NewBuffer(bodyExBytes))
	// ...
	//req.Header.Add("authorization", TestInfos[i].Token)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	ExRes := ExchangeKeysRes{}
	err = goutils.JsonBytesToType(bodyBytes, &ExRes)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	if ExRes.APIPublicKey == "" {
		goutils.PrintError(errors.New("Error in exchange key, APIPublicKey is empty"))
		return false
	}
	goutils.Log("Successfully exchanged keys for user " + TestInfos[i].Alias)
	TestInfos[i].APIPubKey = ExRes.APIPublicKey
	return true
}
func APIAuth(r *rand.Rand, settings Settings, TestInfos []TestInfo, routes APIRoutes, i int, v Node) bool {
	bodyAuth := AuthReq{
		Alias: TestInfos[i].Alias,
		Pass:  TestInfos[i].Pass,
	}
	authBytes, err := goutils.ToJsonBytes(bodyAuth)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	client := &http.Client{
		Timeout: 20 * time.Second,
	}
	req, err := http.NewRequest("POST", v.Host+":"+v.Port+routes.auth, bytes.NewBuffer(authBytes))
	// ...
	req.Header.Add("x-shockwallet-device-id", TestInfos[i].DeviceID)
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)

	if err != nil {
		goutils.PrintError(err)
		return false
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	bodyBytes, err = TestInfos[i].Decrypt(bodyBytes)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	resAuth := AuthRes{}
	err = goutils.JsonBytesToType(bodyBytes, &resAuth)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	if resAuth.User.APIPubKey == "" || resAuth.Token == "" {
		goutils.PrintError(errors.New("error in auth empty token or API pub"))
		return false
	}
	goutils.Log("Successfully auth user " + TestInfos[i].Alias + " with pass " + TestInfos[i].Pass)
	TestInfos[i].GunPubKey = resAuth.User.APIPubKey
	TestInfos[i].Token = resAuth.Token
	return true
}

func GetDisplayName(r *rand.Rand, settings Settings, TestInfos []TestInfo, routes APIRoutes, i int, v Node) bool {
	client := &http.Client{
		Timeout: 20 * time.Second,
	}
	req, err := http.NewRequest("GET", v.Host+":"+v.Port+routes.getDisplayName, nil)
	// ...
	req.Header.Add("x-shockwallet-device-id", TestInfos[i].DeviceID)
	req.Header.Add("Authorization", TestInfos[i].Token)
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)

	if err != nil {
		goutils.PrintError(err)
		return false
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	plain, err := TestInfos[i].Decrypt(bodyBytes)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	dataRes := DataRes{}
	goutils.JsonBytesToType(plain, &dataRes)
	if dataRes.Data == "" {
		return false
	}
	goutils.Log("Successfully got " + dataRes.Data + " as alias for " + TestInfos[i].Alias + " alias")
	return true
}

func SetDisplayName(r *rand.Rand, settings Settings, TestInfos []TestInfo, routes APIRoutes, i int, v Node) bool {
	data := ActionSetDisplayName{
		Token:       TestInfos[i].Token,
		DisplayName: TestInfos[i].Alias,
	}
	dataBytes, err := goutils.ToJsonBytes(data)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	encBytes, err := TestInfos[i].Encrypt(dataBytes)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	dataS, err := goutils.ToJsonString(encBytes)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	dataRes, err := RunSocketIO(settings.TestServerPort, "js/emitEvent", TestInfos[i], false, dataS, routes.setDisplayName)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	decData, err := TestInfos[i].Decrypt(dataRes)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	okRes := OkRes{}
	goutils.JsonBytesToType(decData, &okRes)
	if !okRes.Ok {
		goutils.Log(string(decData))
		return false
	}
	goutils.Log("Successfully set " + TestInfos[i].Alias + " alias")
	return true
}

func GenerateHandshakeNode(r *rand.Rand, settings Settings, TestInfos []TestInfo, routes APIRoutes, i int, v Node) bool {
	data := TokenRes{
		Token: TestInfos[i].Token,
	}
	dataBytes, err := goutils.ToJsonBytes(data)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	encBytes, err := TestInfos[i].Encrypt(dataBytes)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	dataS, err := goutils.ToJsonString(encBytes)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	dataRes, err := RunSocketIO(settings.TestServerPort, "js/emitEvent", TestInfos[i], false, dataS, routes.generateHSNode)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	decData, err := TestInfos[i].Decrypt(dataRes)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	okRes := OkRes{}
	goutils.JsonBytesToType(decData, &okRes)
	if !okRes.Ok {
		goutils.Log(string(decData))
		return false
	}
	goutils.Log("Successfully generated handshake node for " + TestInfos[i].Alias)
	return true
}

func GetHandshakeAddress(r *rand.Rand, settings Settings, TestInfos []TestInfo, routes APIRoutes, i int, v Node) bool {
	client := &http.Client{
		Timeout: 20 * time.Second,
	}
	req, err := http.NewRequest("GET", v.Host+":"+v.Port+routes.getHandshakeAddress, nil)
	// ...
	req.Header.Add("x-shockwallet-device-id", TestInfos[i].DeviceID)
	req.Header.Add("Authorization", TestInfos[i].Token)
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)

	if err != nil {
		goutils.PrintError(err)
		return false
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	plain, err := TestInfos[i].Decrypt(bodyBytes)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	dataRes := DataRes{}
	goutils.JsonBytesToType(plain, &dataRes)
	if dataRes.Data == "" {
		return false
	}
	TestInfos[i].HandshakeAddress = dataRes.Data
	goutils.Log("Successfully got " + dataRes.Data + " as handshake address")
	return true
}
func GetHandshakeRequests(r *rand.Rand, settings Settings, TestInfos []TestInfo, routes APIRoutes, i int, v Node) bool {
	client := &http.Client{
		Timeout: 20 * time.Second,
	}
	req, err := http.NewRequest("GET", v.Host+":"+v.Port+routes.getHandshakeRequests, nil)
	// ...
	req.Header.Add("x-shockwallet-device-id", TestInfos[i].DeviceID)
	req.Header.Add("Authorization", TestInfos[i].Token)
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)

	if err != nil {
		goutils.PrintError(err)
		return false
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	plain, err := TestInfos[i].Decrypt(bodyBytes)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	dataRes := DataArrRes{}
	goutils.JsonBytesToType(plain, &dataRes)
	if len(dataRes.Data) == 0 {
		goutils.Log(string(plain))
		return false
	}
	goutils.Log("Successfully got " + strconv.Itoa(len(dataRes.Data)) + " handshake requests")
	goutils.Dump(dataRes.Data)
	return true
}

func SendHandshake(r *rand.Rand, settings Settings, TestInfos []TestInfo, routes APIRoutes, i int, v Node, to string) bool {
	uuid := strconv.FormatInt(time.Now().Unix(), 10)
	data := ActionSendHandshake{
		Token:           TestInfos[i].Token,
		RecipientPubKey: to,
		UuId:            uuid,
	}
	dataBytes, err := goutils.ToJsonBytes(data)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	encBytes, err := TestInfos[i].Encrypt(dataBytes)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	dataS, err := goutils.ToJsonString(encBytes)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	dataRes, err := RunSocketIO(settings.TestServerPort, "js/emitEvent", TestInfos[i], false, dataS, routes.sendHandshake)

	if err != nil {
		goutils.PrintError(err)
		return false
	}
	decData, err := TestInfos[i].Decrypt(dataRes)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	okRes := OkRes{}
	goutils.JsonBytesToType(decData, &okRes)
	if !okRes.Ok {
		goutils.Log(string(decData))
		return false
	}
	goutils.Log("Successfully sent request to " + to)
	return true
}

func RunSocketIO(port string, file string, info TestInfo, sender bool, data string, event string) ([]byte, error) {
	if sender {
		return execute("node",
			file,
			"-h", "http://localhost:"+port,
			"-a", info.Alias,
			"-p", info.Pass,
			"-i", info.Address,
			"-d", info.DeviceID,
			"-t", info.Token,
			"-e", data,
			"-x", event,
			"-s",
		)
	} else {
		return execute("node",
			file,
			"-h", "http://localhost:"+port,
			"-a", info.Alias,
			"-p", info.Pass,
			"-i", info.Address,
			"-d", info.DeviceID,
			"-t", info.Token,
			"-e", data,
			"-x", event,
		)
	}
}
