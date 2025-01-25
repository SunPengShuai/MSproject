package gid

import (
	snowflak "github.com/bwmarrin/snowflake"
)

type GID interface {
	GetInt64() (int64, error)
	GetBase64() (string, error)
}

type SnowFlakeGID struct {
	ID snowflak.ID
}

func (s *SnowFlakeGID) GetInt64() (int64, error) {
	return s.ID.Int64(), nil
}
func (s *SnowFlakeGID) GetBase64() (string, error) {
	return s.ID.Base64(), nil
}
