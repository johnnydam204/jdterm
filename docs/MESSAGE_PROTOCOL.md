# JDTerm Message Protocol

## Message Format

``` json
{
  "version":1,
  "id":1,
  "type":"ping",
  "code":0,
  "data":{},
  "error":""
}
```

## Rules

* version: protocol version.
* id: request/response correlation.
* type: message type.
* code: result code (0=OK).
* data: payload.
* error: human-readable error.

## Initial Message Types

* ping / pong
* listPorts
* connect
* disconnect
* send
* receive
