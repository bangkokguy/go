package impl

import (
  "net/http"
)

type DeviceImpl struct {}

// Get /device
func (device *DeviceImpl) GetIP(w http.ResponseWriter, r *http.Request){ //, params types.GetIPParams) {
  // Implement your logic here
  w.Header().Set("Content-Type", "application/json")

  w.WriteHeader(200)                                                                 
}
