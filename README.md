WebSocket to UDP bridge (for quakejs)
====================================

On the server you are running Quake 3 (or desktop if you are running non-dedicated server) run

    go get github.com/ibuildthecloud/wsudp && wsudp

Then whatever your IP is (for example 192.168.1.143) clients load quakejs from the following URL

https://www.quakejs.com/play?connect%20192.168.1.143:27960
