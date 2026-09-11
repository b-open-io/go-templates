package ordlock

import (
	"bytes"
	"errors"

	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/bsv-blockchain/go-sdk/transaction"
)

// ErrListingCreate is returned by Lock/Create. Listing create is off (OPL-4692);
// Decode, purchase, and cancel of existing listings stay on.
var ErrListingCreate = errors.New("ordlock listing create is disabled")

type OrdLock struct {
	Seller   *script.Address `json:"seller"`
	Price    uint64          `json:"price"`
	PricePer float64         `json:"pricePer"`
	PayOut   []byte          `json:"payout"`
}

// Lock refuses new listing scripts. Buy and cancel of existing listings stay on.
func Lock(_ *script.Address, _ *script.Address, _ uint64) (*script.Script, error) {
	return nil, ErrListingCreate
}

// Create is Lock. Listing create is off.
func Create(seller *script.Address, payout *script.Address, price uint64) (*script.Script, error) {
	return Lock(seller, payout, price)
}

// Lock refuses new listing scripts. Buy and cancel of existing listings stay on.
func (o *OrdLock) Lock() (*script.Script, error) {
	return nil, ErrListingCreate
}

// IsOrdLock reports whether scr contains an OrdLock prefix/suffix pair.
func IsOrdLock(scr *script.Script) bool {
	if scr == nil {
		return false
	}
	prefix := bytes.Index(*scr, OrdLockPrefix)
	if prefix == -1 {
		return false
	}
	return bytes.Index((*scr)[prefix+len(OrdLockPrefix):], OrdLockSuffix) != -1
}

// IsPurchase reports whether an unlocking script is a purchase (vs cancel).
func IsPurchase(unlock *script.Script) bool {
	if unlock == nil {
		return false
	}
	return bytes.Contains(*unlock, OrdLockSuffix)
}

func Decode(scr *script.Script) *OrdLock {
	if sCryptPrefixIndex := bytes.Index(*scr, OrdLockPrefix); sCryptPrefixIndex == -1 {
		return nil
	} else if ordLockSuffixIndex := bytes.Index(*scr, OrdLockSuffix); ordLockSuffixIndex == -1 {
		return nil
	} else if ordLockOps, err := script.DecodeScript((*scr)[sCryptPrefixIndex+len(OrdLockPrefix) : ordLockSuffixIndex]); err != nil || len(ordLockOps) == 0 {
		return nil
	} else {
		// pkhash := lib.PKHash(ordLockOps[0].Data)
		payOutput := &transaction.TransactionOutput{}
		if _, err = payOutput.ReadFrom(bytes.NewReader(ordLockOps[1].Data)); err != nil {
			return nil
		}
		ordLock := &OrdLock{
			Price:  payOutput.Satoshis,
			PayOut: payOutput.Bytes(),
		}
		if ordLock.Seller, err = script.NewAddressFromPublicKeyHash(ordLockOps[0].Data, true); err != nil {
			return nil
		}

		return ordLock
	}
}
