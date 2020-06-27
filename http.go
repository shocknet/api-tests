package main

/*
import (
	"net/http"

	"github.com/boufni95/goutils"
)

func (s *TestServer) Decrypt(w http.ResponseWriter, r *http.Request) {
	encrypted := EncryptedDataHttp{}
	err := goutils.ExtractReqBody(r, &encrypted)
	if err != nil {
		goutils.SendResError(w, err.Error())
		goutils.PrintError(err)
		return
	}

	instanceData := s.Infos[encrypted.Alias]
	plainText, err := instanceData.Decrypt([]byte(encrypted.Data))
	if err != nil {
		goutils.SendResError(w, err.Error())
		goutils.PrintError(err)
		return
	}
	decrypted := DecryptedDataHttp{
		Alias: encrypted.Alias,
		Data:  string(plainText),
	}
	resBytes, err := goutils.ToJsonBytes(decrypted)
	if err != nil {
		goutils.SendResError(w, err.Error())
		goutils.PrintError(err)
		return
	}
	w.Write(resBytes)
}

func (s *TestServer) Encrypt(w http.ResponseWriter, r *http.Request) {
	plain := DecryptedDataHttp{}
	err := goutils.ExtractReqBody(r, &plain)
	if err != nil {
		goutils.SendResError(w, err.Error())
		goutils.PrintError(err)
		return
	}
	instanceData := s.Infos[plain.Alias]
	Encrypted, err := instanceData.Encrypt([]byte(plain.Data))
	if err != nil {
		goutils.SendResError(w, err.Error())
		goutils.PrintError(err)
		return
	}
	encryptedJString, err := goutils.ToJsonString(Encrypted)
	if err != nil {
		goutils.SendResError(w, err.Error())
		goutils.PrintError(err)
		return
	}
	res := EncryptedDataHttp{
		Alias: plain.Alias,
		Data:  encryptedJString,
	}
	resBytes, err := goutils.ToJsonBytes(res)
	if err != nil {
		goutils.SendResError(w, err.Error())
		goutils.PrintError(err)
		return
	}
	w.Write(resBytes)
}
*/
/*
infoMap := make(map[string]TestInfo)
for _, v := range TestInfos {
	infoMap[v.Alias] = v
}
testServer := TestServer{
	Infos: infoMap,
}

router := mux.NewRouter()
router.HandleFunc("/encrypt", testServer.Encrypt)
router.HandleFunc("/decrypt", testServer.Decrypt)
http.Handle("/", router)

for i, v := range TestInfos {
	isSender := i%2 == 0
	if i == len(TestInfos)-1 {
		isSender = false
	}
	go RunSocketIO(settings.TestServerPort, v, isSender)
}

log.Fatal(http.ListenAndServe(":"+settings.TestServerPort, nil))*/
