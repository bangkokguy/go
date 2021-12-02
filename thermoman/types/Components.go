package types

type Error struct {
  Code int32  `json:"code"`
  Message string  `json:"message"`
}

type User struct {
  Id int64  `json:"id"`
  Name string  `json:"name"`
  Tag string  `json:"tag"`
}

