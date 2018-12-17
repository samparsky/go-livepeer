package pm

import (
	"fmt"
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Validator is an interface which describes an object capable
// of validating tickets
type Validator interface {
	// IsValidTicket checks if a ticket is valid
	IsValidTicket(ticket *Ticket, sig []byte, recipientRand *big.Int) (bool, error)

	// IsWinningTicket checks if a ticket won
	// Note: This method does not check if a ticket is valid which is done using IsValidTicket
	IsWinningTicket(ticket *Ticket, sig []byte, recipientRand *big.Int) bool
}

// BrokerValidator is an implementation of the Validator interface
// that relies on an implementation of the Broker interface to provide
// a set of already used tickets
type BrokerValidator struct {
	addr        ethcommon.Address
	broker      Broker
	sigVerifier SigVerifier
}

// NewBrokerValidator returns an instance of a broker validator
func NewBrokerValidator(addr ethcommon.Address, broker Broker, sigVerifier SigVerifier) *BrokerValidator {
	return &BrokerValidator{
		addr:        addr,
		broker:      broker,
		sigVerifier: sigVerifier,
	}
}

// IsValidTicket checks if a ticket is valid
func (bv *BrokerValidator) IsValidTicket(ticket *Ticket, sig []byte, recipientRand *big.Int) (bool, error) {
	if ticket.Recipient != bv.addr {
		return false, fmt.Errorf("invalid ticket recipient")
	}

	if (ticket.Sender == ethcommon.Address{}) {
		return false, fmt.Errorf("invalid ticket sender")
	}

	if crypto.Keccak256Hash(recipientRand.Bytes()) != ticket.RecipientRandHash {
		return false, fmt.Errorf("invalid preimage provided for hash commitment recipientRandHash")
	}

	used, err := bv.broker.IsUsedTicket(ticket)
	if err != nil {
		return false, err
	}

	if used {
		return false, fmt.Errorf("ticket has already been used")
	}

	if bv.sigVerifier.Verify(ticket.Sender, sig, ticket.Hash().Bytes()) {
		return false, fmt.Errorf("invalid sender signature over ticket hash")
	}

	return true, nil
}

// IsWinningTicket checks if a ticket won
// Note: This method does not check if a ticket is valid which is done using IsValidTicket
// A ticket wins if:
// H(SIG(H(T)), T.RecipientRand) < T.WinProb
func (bv *BrokerValidator) IsWinningTicket(ticket *Ticket, sig []byte, recipientRand *big.Int) bool {
	recipientRandBytes := ethcommon.LeftPadBytes(recipientRand.Bytes(), bytes32Size)
	res := new(big.Int).SetBytes(crypto.Keccak256(sig, recipientRandBytes))

	return res.Cmp(ticket.WinProb) < 0
}
