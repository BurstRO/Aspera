package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math/big"

	"github.com/ac0v/aspera/pkg/burstmath"
	jutils "github.com/ac0v/aspera/pkg/json"
	t "github.com/ac0v/aspera/pkg/transaction"

	"github.com/json-iterator/go"
)

var (
	ErrBlockUnexpectedLen = errors.New("block unexpected length in byte serialisation")
)

const (
	// TODO: move constants
	oneBurst = 100000000
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Block struct {
	PayloadLength       uint32           `json:"payloadLength"`
	TotalAmountNQT      int64            `json:"totalAmountNQT"`
	GenerationSignature jutils.HexSlice  `json:"generationSignature,omitempty"`
	GeneratorPublicKey  jutils.HexSlice  `json:"generatorPublicKey,omitempty"`
	PayloadHash         jutils.HexSlice  `json:"payloadHash,omitempty"`
	BlockSignature      jutils.HexSlice  `json:"blockSignature,omitempty"`
	Transactions        []*t.Transaction `json:"transactions"`
	Version             int32            `json:"version,omitempty"`
	Nonce               uint64           `json:"nonce,omitempty,string"`
	TotalFeeNQT         int64            `json:"totalFeeNQT,omitempty"`
	BlockATs            jutils.HexSlice  `json:"blockATs"`
	PreviousBlock       uint64           `json:"previousBlock,omitempty,string"`
	Timestamp           uint32           `json:"timestamp,omitempty"`
	Block               uint64           `json:"block,omitempty,string"`
	Height              int32            `json:"height,omitempty"`
	PreviousBlockHash   jutils.HexSlice  `json:"previousBlockHash,omitempty"` // if version > 1
	isValid             bool             `struct:"-"`
}

func (b *Block) CalcScoop() uint32 {
	return burstmath.CalcScoop(b.Height, b.GenerationSignature)
}

func (b *Block) ToBytes() ([]byte, error) {
	bsCap := 4 + 4 + 8 + 4 + 4 + 32 + 32 + (32 + 32) + 8 + 64
	if b.Version < 3 {
		bsCap += 4 + 4
	} else {
		bsCap += 8 + 8
	}
	bsCap += len(b.BlockATs)

	w := bytes.NewBuffer(nil)

	if err := binary.Write(w, binary.LittleEndian, b.Version); err != nil {
		return nil, err
	}

	if err := binary.Write(w, binary.LittleEndian, b.Timestamp); err != nil {
		return nil, err
	}

	if err := binary.Write(w, binary.LittleEndian, b.PreviousBlock); err != nil {
		return nil, err
	}

	if err := binary.Write(w, binary.LittleEndian, uint32(len(b.Transactions))); err != nil {
		return nil, err
	}

	if b.Version < 3 {
		totalAmountQNT := int32(b.TotalAmountNQT / oneBurst)
		if err := binary.Write(w, binary.LittleEndian, totalAmountQNT); err != nil {
			return nil, err
		}

		totalFeeNQT := int32(b.TotalFeeNQT / oneBurst)
		if err := binary.Write(w, binary.LittleEndian, totalFeeNQT); err != nil {
			return nil, err
		}
	} else {
		if err := binary.Write(w, binary.LittleEndian, b.TotalAmountNQT); err != nil {
			return nil, err
		}

		if err := binary.Write(w, binary.LittleEndian, b.TotalFeeNQT); err != nil {
			return nil, err
		}
	}

	if err := binary.Write(w, binary.LittleEndian, b.PayloadLength); err != nil {
		return nil, err
	}

	if err := binary.Write(w, binary.LittleEndian, b.PayloadHash); err != nil {
		return nil, err
	}

	if err := binary.Write(w, binary.LittleEndian, b.GeneratorPublicKey); err != nil {
		return nil, err
	}

	if err := binary.Write(w, binary.LittleEndian, b.GenerationSignature); err != nil {
		return nil, err
	}

	if b.Version > 1 {
		if err := binary.Write(w, binary.LittleEndian, b.PreviousBlockHash); err != nil {
			return nil, err
		}
	}

	if err := binary.Write(w, binary.LittleEndian, b.Nonce); err != nil {
		return nil, err
	}

	if b.BlockATs != nil {
		if err := binary.Write(w, binary.LittleEndian, b.BlockATs); err != nil {
			return nil, err
		}
	}

	if err := binary.Write(w, binary.LittleEndian, b.BlockSignature); err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

func (b *Block) CalculateHash() (*[32]byte, error) {
	if bs, err := b.ToBytes(); err == nil {
		bs := sha256.Sum256(bs)
		return &bs, nil
	} else {
		return nil, err
	}
}

func (b *Block) CalculateID() (uint64, error) {
	if hash, err := b.CalculateHash(); err == nil {
		bigInt := big.NewInt(0)
		bigInt.SetBytes([]byte{
			hash[7], hash[6], hash[5], hash[4],
			hash[3], hash[2], hash[1], hash[0]})
		return bigInt.Uint64(), nil
	} else {
		return 0, err
	}
}

func (b *Block) toError(message string) error {
	b.isValid = false
	if v, err := json.Marshal(b); err == nil {
		return errors.New(message + " << " + string(v))
	} else {
		return nil
	}
}

func (b *Block) Validate(previous *Block) error {
	if b.Version != 3 {
		return b.toError("invalid block version")
	}
	b.isValid = true
	return nil
}

func (b *Block) IsValid() bool {
	return b.isValid
}
