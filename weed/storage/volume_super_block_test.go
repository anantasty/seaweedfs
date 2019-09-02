package storage

import (
	"testing"

	"github.com/joeslay/seaweedfs/weed/storage/needle"
)

func TestSuperBlockReadWrite(t *testing.T) {
	rp, _ := NewReplicaPlacementFromByte(byte(001))
	ttl, _ := needle.ReadTTL("15d")
	s := &SuperBlock{
		version:          needle.CurrentVersion,
		ReplicaPlacement: rp,
		Ttl:              ttl,
	}

	bytes := s.Bytes()

	if !(bytes[2] == 15 && bytes[3] == needle.Day) {
		println("byte[2]:", bytes[2], "byte[3]:", bytes[3])
		t.Fail()
	}

}
