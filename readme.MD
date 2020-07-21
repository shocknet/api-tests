# API tests

This tool creates containers with an image of shocknet api and tests it.

## Requirements
* go runtime or a binary of the tests
* nodejs runtime
* Polar or other pre-configured lnd containers
* user must able to call the docker command

## Config

Example in `config.json`
* `image` the docker image of the api
* `network` the docker network that the lnd containers live in ³
* `newUserEndpoint` the endpoint where `gunauth` is listening to create a new gun user
* `nodes` an array of containers configuration
* * `host` the http address used to reach the api container
* * `aliasBase` the initial characters in the new gun user for this node
* * `name` the name that will be given to the api container ¹ 
* * `port` the port to assign the api container
* * `addr` the address of the lnd node ²
* * `tls` the absolute path to the tls certificate 
* * `macaroon` the absolute path to the macaroon file

¹ `name` can be used to get the api logs via `docker logs <name>`
² `addr` the address can be build with `<lnd-container-name>:<internal-port>`, assuming lnd was started correctly with the right flags (`--tlsextradomain=<lnd-container-name>`). Using the internal port (10009) since this is a connection from container to container. To find the container name when using Polar, run `docker ps` the container name will look like this `polar-n2-alice` and to make sure its possible to very its binding the same ports as the show in Polar
³ `network` Polar will automatically create a network for the nodes, to find the name run `docker inspect --format={{.NetworkSettings.Networks}} <lnd-container-name>` the response will look like this `map[2_default:0xc00067e000]` in this case the network name is `2_default`

## Execution flags
* `-d` delete all created containers and exit
* `-f` delete all containers before creating them
* `-t` run tests

## Build the API image
This test does not build the api image (not yet) the image needs to be built before and provided in the `config.json` file

To build the image move to the API directory and run `docker build -t <image-name> .` the name used here must be passed in the `config.json` file to build the containers

## Actions performed

* delete existing api containers
* create new containers and connect them to the network
* create new gun user via gunauth
* Create RSA 2048 -> MarshalPKCS1 -> pem encode
* Exchange with API for each container (new keys each instance)
* Decrypt/Encrypt RSA 2048 + AES 256 (CBC) PKCS5Padding
* Auth -All API containers
* Set display name -All API containers
* Get display name -All API containers
* Generate handshake node -All API containers
* Get handshake node -All API containers
* Send handshake request -Node #0
* Get handshake requests -Node #1

More actions still need to be implemented

## Execution steps

* Start and Configure Polar. basic setup: 1Bitcoind, 2LND
* Build API image
* Edit `config.json` with the right params
* Start gun auth with `node gunauth/main` It will bind Port 3000 on the local machine
* `yarn install` 
* `go run -f -t .` or `./test_bin -f -t`to start the tests