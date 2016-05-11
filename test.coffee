net = require 'net'


connectionIdCount = 0


server = net.createServer (socket) ->
  connectionId = connectionIdCount++
  console.log "NEW CONNECTION: #{connectionId}"
  
  socket.on 'data', (data) ->
    console.log "CONNECTION #{connectionId} DATA: ", data.toString()
    socket.write "GOT: #{data.toString()}"
    
  socket.on 'disconnect', -> console.log "CONNECTION #{connectionId} DISCONNECT"


server.listen 9000
