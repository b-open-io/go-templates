package ordlock

import (
	"errors"

	ec "github.com/bsv-blockchain/go-sdk/primitives/ec"
	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/bsv-blockchain/go-sdk/transaction"
	sighash "github.com/bsv-blockchain/go-sdk/transaction/sighash"

	"github.com/bitcoin-sv/go-templates/template/p2pkh"
)

var (
	ErrNoPrivateKey = errors.New("private key not supplied")
	ErrMalformedTx  = errors.New("malformed transaction: requires at least 2 outputs")
)

// CancelUnlocker spends an existing listing via the seller cancel path.
type CancelUnlocker struct {
	PrivateKey  *ec.PrivateKey
	SigHashFlag *sighash.Flag
}

// CancelListing unlocks an existing listing for the seller. Create stays off.
func CancelListing(key *ec.PrivateKey, sigHashFlag *sighash.Flag) (*CancelUnlocker, error) {
	if key == nil {
		return nil, ErrNoPrivateKey
	}
	if sigHashFlag == nil {
		shf := sighash.AllForkID
		sigHashFlag = &shf
	}
	return &CancelUnlocker{
		PrivateKey:  key,
		SigHashFlag: sigHashFlag,
	}, nil
}

func (c *CancelUnlocker) Sign(tx *transaction.Transaction, inputIndex uint32) (*script.Script, error) {
	s, err := (&p2pkh.P2PKH{
		PrivateKey:  c.PrivateKey,
		SigHashFlag: c.SigHashFlag,
	}).Sign(tx, inputIndex)
	if err != nil {
		return nil, err
	}
	if err = s.AppendOpcodes(script.Op1); err != nil {
		return nil, err
	}
	return s, nil
}

func (c *CancelUnlocker) EstimateLength(_ *transaction.Transaction, _ uint32) uint32 {
	return 107
}

// PurchaseUnlocker spends an existing listing via the buyer purchase path.
type PurchaseUnlocker struct{}

// PurchaseListing unlocks an existing listing for a buyer. Create stays off.
func PurchaseListing() *PurchaseUnlocker {
	return &PurchaseUnlocker{}
}

func (p *PurchaseUnlocker) Sign(tx *transaction.Transaction, inputIndex uint32) (*script.Script, error) {
	if len(tx.Outputs) < 2 {
		return nil, ErrMalformedTx
	}
	s := &script.Script{}
	if err := s.AppendPushData(tx.Outputs[0].Bytes()); err != nil {
		return nil, err
	}
	if len(tx.Outputs) > 2 {
		var extra []byte
		for _, out := range tx.Outputs[2:] {
			extra = append(extra, out.Bytes()...)
		}
		if err := s.AppendPushData(extra); err != nil {
			return nil, err
		}
	} else if err := s.AppendOpcodes(script.Op0); err != nil {
		return nil, err
	}
	preimage, err := tx.CalcInputPreimage(inputIndex, sighash.All|sighash.AnyOneCanPayForkID)
	if err != nil {
		return nil, err
	}
	if err = s.AppendPushData(preimage); err != nil {
		return nil, err
	}
	if err = s.AppendOpcodes(script.Op0); err != nil {
		return nil, err
	}
	return s, nil
}

func (p *PurchaseUnlocker) EstimateLength(tx *transaction.Transaction, inputIndex uint32) uint32 {
	s, err := p.Sign(tx, inputIndex)
	if err != nil {
		return 0
	}
	return uint32(len(*s))
}
