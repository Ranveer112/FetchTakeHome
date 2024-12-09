Receipt Processor
A a webservice that fulfils the documented API. Following pre-requistes are required to host this server
1) Go compiler should be installed.

To run this server, on unix, first build the executable using
`go build server.go`

and then run the executable with `-port` flag. For example, to host this server on port 50, run the following command
on the same directory that the executable lies in:
`./server -port 50` 