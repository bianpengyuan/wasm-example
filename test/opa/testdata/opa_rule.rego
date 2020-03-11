package test

default allow = false

allow {
  input.requestMethod == "GET"
  input.sourcePrincipal == "spiffe://cluster.local/ns/default/sa/client"
  input.destinationService == "server.default.svc.cluster.local"
  glob.match("/echo", [], input.requestUrlPath)
}

allow {
  input.requestMethod == "POST"
  input.sourcePrincipal == "spiffe://cluster.local/ns/default/sa/client"
  input.destinationService == "server.default.svc.cluster.local"
  glob.match("/echo", [], input.requestUrlPath)
}
