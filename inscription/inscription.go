package inscription

import (
	"bytes"
	"crypto/sha256"
	"unicode/utf8"

	"github.com/bsv-blockchain/go-sdk/overlay"
	"github.com/bsv-blockchain/go-sdk/script"
)

type File struct {
	Hash    []byte `json:"hash"`
	Size    uint32 `json:"size"`
	Type    string `json:"type"`
	Content []byte `json:"-"`
}

type Inscription struct {
	File         File              `json:"file,omitempty"`
	Parent       *overlay.Outpoint `json:"parent,omitempty"`
	Bitcom       map[string][]byte `json:"bitcom,omitempty"`
	ScriptPrefix []byte            `json:"prefix,omitempty"`
	ScriptSuffix []byte            `json:"suffix,omitempty"`
}

func Decode(scr *script.Script) *Inscription {
	for pos := 0; pos < len(*scr); {
		startI := pos
		if op, err := scr.ReadOp(&pos); err != nil {
			break
		} else if pos > 2 && op.Op == script.OpDATA3 && bytes.Equal(op.Data, []byte("ord")) && (*scr)[startI-2] == 0 && (*scr)[startI-1] == script.OpIF {
			insc := &Inscription{
				ScriptPrefix: (*scr)[:startI-2],
			}

		ordLoop:
			for {
				var field int
				var err error
				var op, op2 *script.ScriptChunk
				if op, err = scr.ReadOp(&pos); err != nil || op.Op > script.Op16 {
					return insc
				} else if op2, err = scr.ReadOp(&pos); err != nil || op2.Op > script.Op16 {
					return insc
				} else if op.Op > script.OpPUSHDATA4 && op.Op <= script.Op16 {
					field = int(op.Op) - 80
				} else if len(op.Data) == 1 {
					field = int(op.Data[0])
				} else if len(op.Data) > 1 {
					if add, err := script.NewAddressFromString(string(op.Data)); err == nil {
						if insc.Bitcom == nil {
							insc.Bitcom = make(map[string][]byte)
						}
						insc.Bitcom[add.AddressString] = op2.Data
					}
					continue
				}
				switch field {
				case 0:
					insc.File.Content = op2.Data
					insc.File.Size = uint32(len(insc.File.Content))
					hash := sha256.Sum256(insc.File.Content)
					insc.File.Hash = hash[:]
					break ordLoop
				case 1:
					if len(op2.Data) < 256 && utf8.Valid(op2.Data) {
						insc.File.Type = string(op2.Data)
					}
				case 3:
					if len(op2.Data) == 36 {
						insc.Parent = overlay.NewOutpointFromBytes([36]byte(op2.Data))
					}
				}

			}
			op, err := scr.ReadOp(&pos)
			if err != nil || op.Op != script.OpENDIF {
				insc.ScriptSuffix = (*scr)[pos:]
				return insc
			}
		}
	}
	return nil
}

func (i *Inscription) Lock() *script.Script {
	s := script.NewFromBytes(i.ScriptPrefix)
	s.AppendOpcodes(script.Op0, script.OpIF)
	s.AppendPushData([]byte("ord"))
	s.AppendOpcodes(script.Op1)
	s.AppendPushData(i.File.Content)
	if i.Parent != nil {
		s.AppendOpcodes(script.Op3)
		s.AppendPushData(i.Parent.Bytes())
	}
	if i.Bitcom != nil {
		for k, v := range i.Bitcom {
			s.AppendPushData([]byte(k))
			s.AppendPushData(v)
		}
	}
	s.AppendOpcodes(script.Op0)
	s.AppendPushData(i.File.Content)
	s.AppendOpcodes(script.OpENDIF)
	return script.NewFromBytes(append(*s, i.ScriptSuffix...))
}
