const express = require('express')
const Http = require('http')
const Gun = require('gun')
require('gun/nts.js')
var bodyParser = require('body-parser')
const gun = new Gun({
    axe:false,
      //peers: ["http://localhost:8765/gun"]
    //peers: ["http://gun.shock.network:8765/gun"]
    peers: ["http://guntest.shock.network:8765/gun"]
      //peers: ["http://167.88.11.206:8765/gun"]
      //peers: ["http://guntest.herokuapp.com/gun"]
    })
user = gun.user()
const app = express()
const port = 3000
app.use(bodyParser.urlencoded({ extended: false }))
app.use(bodyParser.json())
app.use(async (req, res, next) => {
    console.log('Route:', req.path)
    next()
    
  })
app.get('/', (req, res) => res.send('Hello World!'))
app.post('/create', (req, res) => {
    console.log(req.body)
    const {alias,pass} = req.body
    user.create(alias, pass, ackC => {
        if(ackC.err) {
            let err = ackC.err
            if(typeof ackC.err === 'number'){
                err =  err.toString()
            }
            res.json({ok:false,err:err})
            return
        }
        user.auth(alias, pass, ackA => {
            if(ackA.err){
                let err = ackA.err
                if(typeof ackA.err === 'number'){
                    err =  err.toString()
                }
                res.json({ok:false,err:ackA.err})
            }
            res.json({ok:true,err:""})
        })
    })
})
const httpServer = Http.Server(app)
httpServer.listen(port, '0.0.0.0')
console.log(httpServer.address())