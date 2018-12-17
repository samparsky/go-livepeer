package pm

import (
	"fmt"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// SigVerifier is an interface which describes an object capable
// of verification of ECDSA signatures produced by ETH addresses
type SigVerifier interface {
	// Verify checks if a provided signature over a message
	// is valid for a given ETH address
	Verify(addr ethcommon.Address, sig []byte, msg []byte) bool
}

// BrokerSigVerifier is an implementation of the SigVerifier interface
// that relies on an implementation of the Broker interface to provide a registry
// mapping ETH addresses to approved signer sets. This implementation will
// recover a ETH address from a signature and check if the recovered address
// is approved
type BrokerSigVerifier struct {
	broker Broker
}

// NewBrokerSigVerifier returns an instance of a broker signature verifier
func NewBrokerSigVerifier(broker Broker) *BrokerSigVerifier {
	return &BrokerSigVerifier{
		broker: broker,
	}
}

// Verify checks if a provided signature over a message
// is valid for a given ETH address
func (bsv *BrokerSigVerifier) Verify(addr ethcommon.Address, sig []byte, msg []byte) (bool, error) {
	personalMsg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", 32, msg)
	personalHash := crypto.Keccak256([]byte(personalMsg))

	pubkey, err := crypto.SigToPub(personalHash, sig)
	if err != nil {
		return false, err
	}

	rec := crypto.PubkeyToAddress(*pubkey)

	if addr == rec {
		// If recovered address matches, return early
		return true, nil
	}

	approved, err := bsv.broker.IsApprovedSigner(addr, rec)
	if err != nil {
		return false, err
	}

	return approved, nil
}
