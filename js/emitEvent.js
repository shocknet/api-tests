const program = require("commander");
const SocketIO = require("socket.io-client")
const fetch = require("node-fetch");

program.version("1.0.0")
    .option("-h, --helperHost [alias]","helper host address")
    .option("-a, --user-alias [alias]","user alias")
    .option("-p, --pass [pass]","user pass")
    .option("-i, --api-host [host]","api host address")
    .option("-d, --device-id [deviceid]","device id for socket io connection")
    .option("-t, --token [token]","token socket io connection")
    .option("-s, --sender","whether sender or receiver in a two sided non repeatable action")
    .option("-e, --extra-data [data]","the data")
    .option("-x, --event-name [data]","the event name")
    .parse(process.argv);
//console.log(program)
if(!program.helperHost ||!program.userAlias || !program.pass || !program.apiHost || !program.deviceId || !program.token){
    console.log("missing param: please provide all params ,-h, -a, -p, -i, -d, -t")
}
const socket = SocketIO(`${program.apiHost}`,{
    autoConnect:true,
    reconnectionAttempts:Infinity,
    query:{
        'x-shockwallet-device-id':program.deviceId
    }
})
socket.connect()
socket.on(program.eventName,res => {
    console.log(JSON.stringify(res))
    process.exit(0)
})
socket.on("connect",()=>{
    runTest()
    .catch(e => process.exit(1))
})
socket.on('connect_error', error => {
    process.exit(1)
  })

  socket.on('connect_timeout', timeout => {
    process.exit(1)
  })
const runTest =  async () => {
    const toSend = JSON.parse(program.extraData)
    socket.emit(program.eventName,toSend)
}