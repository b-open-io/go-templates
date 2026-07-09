package shrug

import (
	"math/big"

	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/bsv-blockchain/go-sdk/script/interpreter"
	"github.com/bsv-blockchain/go-sdk/transaction"
)

const SHRUG_TAG = "¯\\_(ツ)_/¯"

type Shrug struct {
	Id           *transaction.Outpoint // nil = deploy; the token id is this output's outpoint
	Amount       *big.Int              // 0 = mint authority, >0 = token value; arbitrary precision
	ScriptSuffix []byte
}

func Decode(s *script.Script) *Shrug {
	shrug := &Shrug{}
	pos := 0

	if op, err := s.ReadOp(&pos); err != nil {
		return nil
	} else if string(op.Data) != SHRUG_TAG {
		return nil
	}

	if op, err := s.ReadOp(&pos); err != nil || op.Op > script.OpPUSHDATA4 {
		return nil
	} else if len(op.Data) == 36 {
		shrug.Id = transaction.NewOutpointFromBytes(op.Data)
	} else if len(op.Data) != 0 {
		return nil
	}

	if op, err := s.ReadOp(&pos); err != nil || op.Op != script.Op2DROP {
		return nil
	}

	if op, err := s.ReadOp(&pos); err != nil || op.Op > script.OpPUSHDATA4 {
		return nil
	} else if number, err := interpreter.MakeScriptNumber(op.Data, len(op.Data), true, true); err != nil {
		return nil
	} else if number.Val.Sign() < 0 {
		return nil
	} else {
		shrug.Amount = number.Val
	}

	if op, err := s.ReadOp(&pos); err != nil || op.Op != script.OpDROP {
		return nil
	}

	shrug.ScriptSuffix = (*s)[pos:]
	return shrug
}

func (i *Shrug) Lock() *script.Script {
	s := &script.Script{}
	_ = s.AppendPushData([]byte(SHRUG_TAG))
	if i.Id != nil {
		_ = s.AppendPushData(i.Id.Bytes())
	} else {
		_ = s.AppendOpcodes(script.Op0)
	}
	_ = s.AppendOpcodes(script.Op2DROP)
	if i.Amount != nil && i.Amount.Sign() > 0 {
		_ = s.AppendPushData((&interpreter.ScriptNumber{
			Val:          i.Amount,
			AfterGenesis: true,
		}).Bytes())
	} else {
		_ = s.AppendOpcodes(script.Op0)
	}
	_ = s.AppendOpcodes(script.OpDROP)
	return script.NewFromBytes(append(*s, i.ScriptSuffix...))
}
