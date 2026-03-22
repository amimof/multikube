package keys

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

const (
	UUIDLen     = 16 // Bytes
	TagUID  Tag = 0x01
	TagName Tag = 0x02
	TagIdx  Tag = 0x03
)

var (
	ErrEmptyUID    = errors.New("empty uid id")
	ErrEmptyName   = errors.New("empty name id")
	ErrEmptyIndex  = errors.New("empty index id")
	ErrBadEncoding = errors.New("bad id encoding")
	ErrBadTag      = errors.New("unknown id tag")
	ErrBadUIDLen   = errors.New("uuid payload must be 16 bytes")
	ErrBadPrefix   = errors.New("bad or unknown prefix")
	ErrBadUUIDLen  = errors.New("uuid payload must be 16 bytes")
)

type (
	Tag byte
	ID  struct {
		tag Tag
		raw []byte
	}
)

func Index(s string) (ID, error) {
	if s == "" {
		return ID{}, ErrEmptyName
	}
	return ID{tag: TagIdx, raw: []byte(s)}, nil
}

func Name(s string) (ID, error) {
	if s == "" {
		return ID{}, ErrEmptyName
	}
	return ID{tag: TagName, raw: []byte(s)}, nil
}

func UUID(u uuid.UUID) (ID, error) {
	cp := make([]byte, UUIDLen)
	copy(cp, u[:])
	return ID{tag: TagUID, raw: cp}, nil
}

// AsUUID returns the UUID if tag is TagUID.
func (id ID) AsUUID() (uuid.UUID, error) {
	if id.tag != TagUID {
		return uuid.UUID{}, errors.New("id is not a uuid")
	}
	if len(id.raw) != UUIDLen {
		return uuid.UUID{}, ErrBadUIDLen
	}
	var u uuid.UUID
	copy(u[:], id.raw)
	return u, nil
}

func (id ID) Tag() Tag {
	return id.tag
}

func (id ID) Raw() []byte {
	return id.raw
}

func (id ID) Encode() []byte {
	out := make([]byte, 0, 1+len(id.raw))
	out = append(out, byte(id.tag))
	out = append(out, id.raw...)
	return out
}

func (id ID) EncodePrefixed(prefix []byte) []byte {
	enc := id.Encode()
	out := make([]byte, 0, len(prefix)+len(enc))
	out = append(out, prefix...)
	out = append(out, enc...)
	return out
}

func (id ID) String() string {
	switch id.Tag() {
	case TagUID:
		u, err := id.AsUUID()
		if err != nil {
			return ""
		}
		return u.String()
	case TagName, TagIdx:
		return string(id.Raw())
	}
	return ""
}

func (id ID) IdxStr() string {
	return id.NameStr()
}

func (id ID) NameStr() string {
	if id.Tag() == TagName || id.Tag() == TagIdx {
		return string(id.Raw())
	}
	return ""
}

func (id ID) UUIDStr() string {
	if id.Tag() == TagUID {
		u, err := id.AsUUID()
		if err != nil {
			return ""
		}
		return u.String()
	}
	return ""
}

// ParseBytes lets you decode ID from a full key (useful in iterators/scans).
func ParseBytes(prefix, fullKey []byte) (ID, error) {
	if !bytes.HasPrefix(fullKey, prefix) {
		return ID{}, ErrBadPrefix
	}
	return Decode(fullKey[len(prefix):])
}

func Decode(b []byte) (ID, error) {
	if len(b) < 2 {
		return ID{}, ErrBadEncoding
	}

	tag := Tag(b[0])
	payload := b[1:]

	switch tag {
	case TagName:
		if len(payload) == 0 {
			return ID{}, ErrEmptyName
		}
		raw := make([]byte, len(payload))
		copy(raw, payload)
		return ID{tag: tag, raw: raw}, nil

	case TagUID:
		if len(payload) != UUIDLen {
			return ID{}, ErrBadUUIDLen
		}
		raw := make([]byte, UUIDLen)
		copy(raw, payload)
		return ID{tag: tag, raw: raw}, nil
	case TagIdx:
		if len(payload) == 0 {
			return ID{}, ErrEmptyIndex
		}
		raw := make([]byte, len(payload))
		copy(raw, payload)
		return ID{tag: tag, raw: raw}, nil
	default:
		return ID{}, ErrBadTag
	}
}

// ParseStr returns a UID from a string. Prioritizes UUID's
func ParseStr(in string) (ID, error) {
	if uid, err := FromUIDOrName(in, ""); err == nil {
		return uid, nil
	}
	if uid, err := FromUIDOrName("", in); err == nil {
		return uid, nil
	}
	return ID{}, fmt.Errorf("error parsing %q as uid", in)
}

// FromUIDOrName creates a ID from the provded strings. Prioritizes UIDs
func FromUIDOrName(uid, name string) (ID, error) {
	if uid != "" {
		u, err := uuid.Parse(uid)
		if err != nil {
			return ID{}, err
		}
		res, err := UUID(u)
		if err != nil {
			return ID{}, err
		}
		return res, nil
	}
	if name != "" {
		res, err := Name(name)
		if err != nil {
			return ID{}, err
		}
		return res, nil
	}
	return ID{}, fmt.Errorf("one of uid or name is required")
}
