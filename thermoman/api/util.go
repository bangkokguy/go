package api

import "strconv"

// Converts a string to an int64 value.
func convertStringToInt64(val string) int64 {
  i64, _ := strconv.ParseInt(val, 10, 64)
  return i64
}

// Converts a string to an int32 value
func convertStringToInt32(val string) int32 {
  return int32(convertStringToInt64(val))
}

// Converts a string to an float64 value
func convertStringToFloat64(val string) float64 {
  f64, _ := strconv.ParseFloat(val, 64)
  return f64
}

// Converts a string to an float32 value
func convertStringToFloat32(val string) float32 {
  return float32(convertStringToFloat64(val))
}