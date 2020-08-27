package main

import (
	"errors"
	"flag"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/boufni95/goutils"
)

func main() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	goutils.InitLogger("debug")
	var settings Settings
	err := goutils.JsonFileToType("config.json", &settings)
	if err != nil {
		goutils.PrintError(err)
		return
	}
	//parse flags
	var force, delete, test, gunauth, single, extra, gunk1, all bool
	flag.BoolVar(&delete, "d", false, "delete all containers")
	flag.BoolVar(&force, "f", false, "force the creation of the container, will run the delete first")
	flag.BoolVar(&test, "t", false, "also run all tests")
	flag.BoolVar(&single, "s", false, "only perform single tests")
	flag.BoolVar(&extra, "x", false, "debug")
	flag.BoolVar(&gunk1, "k", false, "start or delete gunk1 (only)")
	flag.BoolVar(&all, "a", false, "works only with -d and -k to kill everything")
	//flag.BoolVar(&gunauth, "g", false, "run gun auth CAUTION it leaves the node process running after exit")
	flag.Parse()
	//goutils.Dump(settings)
	if gunk1 {
		if delete {
			_, err := execute("docker", "stop", settings.Gunk1.Name)
			if err != nil {
				goutils.PrintError(err)
			}
			_, err = execute("docker", "rm", settings.Gunk1.Name)
			if err != nil {
				goutils.PrintError(err)
			}
			if all {
				removeContainers(settings.Nodes)
			}
			return
		}
		_, err = execute("docker", "create", "--name", settings.Gunk1.Name, "-p", settings.Gunk1.Port+":8080", settings.Gunk1.Image)
		if err != nil {
			goutils.PrintError(err)
		}
		_, err = execute("docker", "network", "connect", settings.Network, settings.Gunk1.Name)
		if err != nil {
			goutils.PrintError(err)
		}
		_, err = execute("docker", "start", settings.Gunk1.Name)
		if err != nil {
			goutils.PrintError(err)
		}
		return
	}
	if delete {
		removeContainers(settings.Nodes)
		return
	}
	if force {
		removeContainers(settings.Nodes)
	}

	//check image exists
	err = checkImage(settings.Image)
	if err != nil {
		goutils.PrintError(err)
		return
	}
	if len(settings.Nodes) == 0 {
		goutils.PrintError(errors.New("please provide at least one node to make the tests"))
		return
	}
	for _, v := range settings.Nodes {
		res := []byte("")
		res, err = execute("docker", "inspect", "--format={{.Name}}", v.Name)
		if err != nil {
			goutils.Log("looks like " + v.Name + " container does not exist, will create it now")
			err = createContainer(settings.Image, v)
			if err != nil {
				goutils.PrintError(err)
				continue
			}
		}
		res, err = execute("docker", "inspect", "--format={{.State.Running}}", v.Name)
		trimRes := strings.TrimSuffix(string(res), "\n")
		if err != nil || string(trimRes) != "true" {
			goutils.Log("looks like " + v.Name + " container exists,but is not running will run it now")
			err = runContainer(settings.Network, v)
			if err != nil {
				goutils.PrintError(err)
				continue
			}
		}
		goutils.Log("looks like " + v.Name + " container exists and is running, no more actions")

	}

	if !test {
		goutils.Log("Operation complete, use the flag -t to also run tests")
		return
	}
	if gunauth {
		goutils.Log("starting gunauth server and waiting 1 sec")
		go executeAsync("node", "gunauth/main")
		time.Sleep(1 * time.Second)
	}
	routes := APIRoutes{
		exchangeKeys:         "/api/security/exchangeKeys",
		auth:                 "/api/lnd/auth",
		getDisplayName:       "/api/gun/ON_DISPLAY_NAME",
		getHandshakeRequests: "/api/gun/ON_RECEIVED_REQUESTS",
		setDisplayName:       "SET_DISPLAY_NAME",
		generateHSNode:       "GENERATE_NEW_HANDSHAKE_NODE",
		initFeedWall:         "INIT_FEED_WALL",
		generateOrderAddress: "GENERATE_ORDER_ADDRESS",
		sendHandshake:        "SEND_HANDSHAKE_REQUEST",
		getHandshakeAddress:  "/api/gun/ON_HANDSHAKE_ADDRESS",
		acceptHSRequest:      "ACCEPT_REQUEST",
	}
	TestInfos := make([]TestInfo, len(settings.Nodes))

	for i, v := range settings.Nodes {
		TestInfos[i].Address = "http://localhost:" + v.Port
		goutils.Log("Handling user: " + v.AliasBase)
		//create new user in a separate runtime cuz Polar is already unlocked, and its not possible to create a new user with the wallet already unlocked
		ok := CreateGunUser(r, settings, TestInfos, i, v)
		if !ok {
			return
		}
		//exchange keys

		ok = ExchangeKeys(r, settings, TestInfos, routes, i, v)
		if !ok {
			return
		}
		//auth with the nely created user
		cont := 1
		for cont <= 10 {
			ok = APIAuth(r, settings, TestInfos, routes, i, v)
			if ok {
				goutils.Log("Successful auth after " + strconv.Itoa(cont) + " attempts")
				break
			}
			if cont == 10 {
				goutils.PrintError(errors.New("cant auth even after 10 attempts"))
				return
			}
			goutils.Log("Auth attempt n" + strconv.Itoa(cont) + " failed, sleeping 7 sec before new attempt")
			cont++
			time.Sleep(7 * time.Second)

		}
		ok = SetDisplayName(r, settings, TestInfos, routes, i, v)
		if !ok {
			return
		}
		ok = GetDisplayName(r, settings, TestInfos, routes, i, v)
		if !ok {
			return
		}
		ok = GenerateHandshakeNode(r, settings, TestInfos, routes, i, v)
		if !ok {
			return
		}
		ok = GenerateOrderAddress(r, settings, TestInfos, routes, i, v)
		if !ok {
			return
		}
		ok = GenerateWall(r, settings, TestInfos, routes, i, v)
		if !ok {
			return
		}
		ok = GetHandshakeAddress(r, settings, TestInfos, routes, i, v)
		if !ok {
			return
		}
		//TODO generate order address -> unimplemented socket
		//TODO SET_BIO

	}
	if single {
		goutils.Log("All single test were successful no more actions")
		return
	}
	if len(settings.Nodes) == 1 {
		goutils.PrintError(errors.New("single tests complete, please provide at least two node to make the other tests"))
		return
	}
	if len(settings.Nodes) < 2 {
		goutils.PrintError(errors.New("more than 2 nodes not implemented yet"))
		return
	}
	for i, v := range settings.Nodes {
		isSender := i%2 == 0
		if i == len(TestInfos)-1 {
			isSender = false
		}
		if isSender {
			ok := SendHandshake(r, settings, TestInfos, routes, i, v, TestInfos[i+1].GunPubKey)
			if !ok {
				return
			}
		} else {
			cont := 0
			for cont <= 3 {
				ok := GetHandshakeRequests(r, settings, TestInfos, routes, i, v)
				if ok {
					goutils.Log("Successful got requests after " + strconv.Itoa(cont) + " attempts")
					break
				}
				if cont == 3 {
					goutils.PrintError(errors.New("cant get requests even after 3 attempts"))
					return
				}
				goutils.Log("requests fetch attempt n" + strconv.Itoa(cont) + " failed, sleeping 5 sec before new attempt")
				cont++
				time.Sleep(5 * time.Second)
			}
			//TODO accept request
			ok := AcceptHandshake(r, settings, TestInfos, routes, i, v)
			if !ok {
				return
			}
		}
	}
	goutils.Log("All test were successful no more actions")

}

func checkImage(name string) error {
	res, err := execute("docker", "inspect", name)
	if err != nil {
		return err
	}
	var resInterface []interface{}
	err = goutils.JsonBytesToType(res, &resInterface)
	if err != nil {
		return err
	}
	if len(resInterface) == 0 {
		return err
	}
	if len(resInterface) != 1 {
		return err
	}
	return nil
}

func createContainer(image string, node Node) error {
	//docker create --name shocknet_api -e LND_ADDR="$ipAddr" -p 9835:9835 shocknet_api_img:1
	_, err := execute("docker", "create", "--name", node.Name, "-e", "LND_ADDR="+node.Addr, "-p", node.Port+":9835", "-P", image)
	return err
}

func runContainer(network string, node Node) error {
	//docker cp /home/bitcoin/.lnd/tls.cert shocknet_api:/usr/src/app/tls.cert
	_, err := execute("docker", "cp", node.TLS, node.Name+":/usr/src/app/tls.cert")
	if err != nil {
		return err
	}
	_, err = execute("docker", "cp", node.Macaroon, node.Name+":/usr/src/app/admin.macaroon")
	if err != nil {
		return err
	}
	//docker network connect testnetw alice-container
	_, err = execute("docker", "network", "connect", network, node.Name)
	if err != nil {
		return err
	}
	//docker start shocknet_api
	_, err = execute("docker", "start", node.Name)
	if err != nil {
		return err
	}
	return nil
}

func removeContainers(nodes []Node) {
	for _, v := range nodes {
		_, err := execute("docker", "stop", v.Name)
		if err != nil {
			goutils.PrintError(err)
		}
		_, err = execute("docker", "rm", v.Name)
		if err != nil {
			goutils.PrintError(err)
		}
	}
}
