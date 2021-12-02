package api

import (
  
  "net/http"
  
)

// ServerInterface represents all server handlers.
type DeviceStub interface {
GetIP(w http.ResponseWriter, r *http.Request/*, params types.GetIPParams*/)

}

type DeviceWrapper struct {
  DeviceDelegate DeviceStub
}


func(stub *DeviceWrapper) GetIP(w http.ResponseWriter, r *http.Request) {


  /*getipParams := types.GetIPParams {
    Id: convertStringToInt32(r.URL.Query().Get("Id")),
                              }  */

  stub.DeviceDelegate.GetIP(w, r/*, getipParams*/)
}
