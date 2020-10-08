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
	"github.com/google/uuid"
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
	resAuth := AuthRes{}
	plain, err := TestInfos[i].Decrypt(bodyBytes)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	err = goutils.JsonBytesToType(plain, &resAuth)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	if resAuth.User.APIPubKey == "" || resAuth.Token == "" {
		goutils.PrintError(errors.New("error in auth empty token or API pub"))
		checkResError(plain)
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
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	dataRes := DataRes{}
	plain, err := TestInfos[i].Decrypt(bodyBytes)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	err = goutils.JsonBytesToType(plain, &dataRes)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	if dataRes.Data == "" {
		return false
	}
	goutils.Log("Successfully got " + dataRes.Data + " as alias for " + TestInfos[i].Alias + " alias")
	return true
}
func SetDisplayName(r *rand.Rand, settings Settings, TestInfos []TestInfo, routes APIRoutes, i int, v Node) bool {
	data := RPCAction{
		Path:  routes.gunKeys.setDisplayName,
		Value: TestInfos[i].Alias,
	}
	okRes := OkRes{}
	ok := postRPC(data, TestInfos[i], v.Host+":"+v.Port+routes.gunRpcPut, &okRes)
	if !ok || !okRes.Ok {
		return false
	}
	goutils.Log("Successfully set " + TestInfos[i].Alias + " alias")
	return true
}

func GenerateHandshakeNode(r *rand.Rand, settings Settings, TestInfos []TestInfo, routes APIRoutes, i int, v Node) bool {
	addr := uuid.New()
	data := RPCAction{
		Path:  routes.gunKeys.setHandshakeNode,
		Value: addr.String(),
	}
	okRes := OkRes{}
	ok := postRPC(data, TestInfos[i], v.Host+":"+v.Port+routes.gunRpcPut, &okRes)
	if !ok || !okRes.Ok {
		return false
	}
	goutils.Log("Successfully generated handshake node for " + TestInfos[i].Alias)
	return true
}
func GenerateOrderAddress(r *rand.Rand, settings Settings, TestInfos []TestInfo, routes APIRoutes, i int, v Node) bool {
	addr := uuid.New()
	data := RPCAction{
		Path:  routes.gunKeys.setOrderAddress,
		Value: addr.String(),
	}
	okRes := OkRes{}
	ok := postRPC(data, TestInfos[i], v.Host+":"+v.Port+routes.gunRpcPut, &okRes)
	if !ok || !okRes.Ok {
		return false
	}
	goutils.Log("Successfully generated order address for " + TestInfos[i].Alias)
	return true
}

func GenerateWall(r *rand.Rand, settings Settings, TestInfos []TestInfo, routes APIRoutes, i int, v Node) bool {
	okRes := OkRes{}
	ok := httpGet(TestInfos[i], v.Host+":"+v.Port+routes.initFeedWall, &okRes)
	if !ok || !okRes.Ok {
		return false
	}
	goutils.Log("Successfully generated feed wall for " + TestInfos[i].Alias)
	return true
}

func GetHandshakeAddress(r *rand.Rand, settings Settings, TestInfos []TestInfo, routes APIRoutes, i int, v Node) bool {
	dataRes := DataRes{}
	ok := httpGet(TestInfos[i], v.Host+":"+v.Port+routes.getHandshakeAddress, &dataRes)
	if !ok {
		return false
	}
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
	reqInterface, ok := dataRes.Data[0].(map[string]interface{})
	if !ok {
		goutils.PrintError(errors.New("Data is not in the right format"))
		goutils.Dump(dataRes.Data)
		goutils.Dump(reqInterface)
		return false
	}
	reqID, ok := reqInterface["id"].(string)
	if !ok {
		goutils.PrintError(errors.New("Data is not in the right format"))
		goutils.Dump(dataRes.Data)
		goutils.Dump(reqID)
		return false
	}

	TestInfos[i].HandshakeReqID = reqID
	goutils.Log("Successfully got " + strconv.Itoa(len(dataRes.Data)) + " handshake requests")
	goutils.Dump(dataRes.Data)
	return true
}

func SendHandshake(r *rand.Rand, settings Settings, TestInfos []TestInfo, routes APIRoutes, i int, v Node, to string) bool {

	data := reqSendHandshake{
		PublicKey: to,
	}
	okRes := OkRes{}
	ok := postRPC(data, TestInfos[i], v.Host+":"+v.Port+routes.sendHandshake, &okRes)
	if !ok || !okRes.Ok {
		return false
	}
	goutils.Log("Successfully sent request to " + to)
	return true
}

func AcceptHandshake(r *rand.Rand, settings Settings, TestInfos []TestInfo, routes APIRoutes, i int, v Node) bool {
	okRes := OkRes{}
	ok := httpPut(TestInfos[i], v.Host+":"+v.Port+routes.acceptHSRequest+TestInfos[i].HandshakeReqID, &okRes)
	if !ok || !okRes.Ok {
		return false
	}
	goutils.Log("Successfully accepted request with ID:" + TestInfos[i].HandshakeReqID)
	return true
}

func httpGet(info TestInfo, url string, response interface{}) bool {
	client := &http.Client{
		Timeout: 20 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	// ...
	req.Header.Add("x-shockwallet-device-id", info.DeviceID)
	req.Header.Add("Authorization", info.Token)
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)

	if err != nil {
		goutils.PrintError(err)
		return false
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	plain, err := info.Decrypt(bodyBytes)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	err = goutils.JsonBytesToType(plain, &response)
	if err != nil {
		return false
	}
	return true
}
func httpPut(info TestInfo, url string, response interface{}) bool {
	client := &http.Client{
		Timeout: 20 * time.Second,
	}
	req, err := http.NewRequest("PUT", url, nil)
	// ...
	req.Header.Add("x-shockwallet-device-id", info.DeviceID)
	req.Header.Add("Authorization", info.Token)
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)

	if err != nil {
		goutils.PrintError(err)
		return false
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	plain, err := info.Decrypt(bodyBytes)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	err = goutils.JsonBytesToType(plain, &response)
	if err != nil {
		return false
	}
	return true
}

func postRPC(data interface{}, info TestInfo, url string, response interface{}) bool {
	client := &http.Client{
		Timeout: 20 * time.Second,
	}
	dataB, err := encryptTypeToBytes(info, data)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(dataB))
	// ...
	req.Header.Add("x-shockwallet-device-id", info.DeviceID)
	req.Header.Add("Authorization", info.Token)
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
	plain, err := info.Decrypt(bodyBytes)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	//fmt.Println(string(plain))
	err = goutils.JsonBytesToType(plain, &response)
	if err != nil {
		goutils.PrintError(err)
		return false
	}
	return true
}

func RunSocketIO(port string, file string, info TestInfo, sender bool, data string, event string) ([]byte, error) {
	if sender {
		return execute("node",
			file,
			"-h", "http://localhost:"+port,
			"-a", info.Alias,
			"-p", info.Pass,
			"-i", info.Address+"/default",
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
			"-i", info.Address+"/default",
			"-d", info.DeviceID,
			"-t", info.Token,
			"-e", data,
			"-x", event,
		)
	}
}
