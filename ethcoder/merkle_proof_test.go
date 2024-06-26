package ethcoder

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"

	"github.com/0xsequence/ethkit/go-ethereum/accounts/abi"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestMerkleProofKnown(t *testing.T) {
	testAddr := common.HexToAddress("0x1e946c284bdBb05Fb6EF41016C524E8681e3d05E")
	leaves := [][]byte{
		testAddr.Bytes(),
		common.HexToAddress("0x1D74B866598B339006160d704642459B04ba890B").Bytes(),
		common.HexToAddress("0x37e948435E916069D3a1431Ddf508421073fF3E7").Bytes(),
		common.HexToAddress("0x29c34A7d23B8BCBE7c5Ec94C6525b78bb5cbAf36").Bytes(),
	}
	mt := NewMerkleTree(leaves, nil, nil)

	expectedRoot := common.Hex2Bytes("2620d31912c95198ebbf40473b7b069e98587ec49d0cd46aacef8c746c682334")
	root := mt.GetRoot()
	assert.Equal(t, expectedRoot, root)
	fmt.Printf("Root: %x\n", root)

	expectedProof := [][]byte{
		common.Hex2Bytes("1d74b866598b339006160d704642459b04ba890b"),
		common.Hex2Bytes("39ceb165765d969b9bfbbab524649adc484bab29db86b6c0df8635feebf0154e"),
	}
	proof, err := mt.GetProof(testAddr.Bytes())
	assert.Nil(t, err)
	for i, p := range proof {
		fmt.Printf("Proof part %d: IsLeft=%v, Data=%x\n", i, p.IsLeft, p.Data)
		assert.Equal(t, expectedProof[i], []byte(p.Data))
	}

	isValid, err := mt.Verify(proof, testAddr.Bytes(), root)
	assert.Nil(t, err)
	assert.True(t, isValid)
}

func TestMerkleProofLarge(t *testing.T) {
	addrCount := 100
	leaves := make([][]byte, addrCount)
	for i := 0; i < addrCount; i++ {
		leaf := make([]byte, 20)
		rand.Read(leaf)
		leaves[i] = leaf
	}

	mt := NewMerkleTree(leaves, nil, nil)

	root := mt.GetRoot()
	assert.NotNil(t, root)

	proof, err := mt.GetProof(leaves[69])
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(proof), 1)

	isValid, err := mt.Verify(proof, leaves[69], root)
	assert.Nil(t, err)
	assert.True(t, isValid)
}

func TestMerkleInvalidLeaf(t *testing.T) {
	invalidLeaf := common.HexToAddress("0x1e946c284bdBb05Fb6EF41016C524E8681e3d05E").Bytes()
	leaves := [][]byte{
		common.HexToAddress("0x1D74B866598B339006160d704642459B04ba890B").Bytes(),
		common.HexToAddress("0x37e948435E916069D3a1431Ddf508421073fF3E7").Bytes(),
		common.HexToAddress("0x29c34A7d23B8BCBE7c5Ec94C6525b78bb5cbAf36").Bytes(),
	}

	mt := NewMerkleTree(leaves, nil, nil)

	root := mt.GetRoot()
	assert.NotNil(t, root)

	// Invalid leaf
	_, err := mt.GetProof(invalidLeaf)
	assert.Error(t, err)

	// Valid proof
	proof, err := mt.GetProof(leaves[0])
	assert.Nil(t, err)

	// Invalid leaf
	isValid, _ := mt.Verify(proof, invalidLeaf, root)
	assert.False(t, isValid)
}

func TestMerkleSingleLeaf(t *testing.T) {
	leaf := common.HexToAddress("0x1e946c284bdBb05Fb6EF41016C524E8681e3d05E").Bytes()
	leaves := [][]byte{
		leaf,
	}

	mt := NewMerkleTree(leaves, nil, nil)

	root := mt.GetRoot()
	assert.NotNil(t, root)

	proof, err := mt.GetProof(leaf)
	assert.Nil(t, err)

	isValid, err := mt.Verify(proof, leaf, root)
	assert.Nil(t, err)
	assert.True(t, isValid)
}

type TLeaf struct {
	Addr    common.Address
	TokenId *big.Int
}

func TestMerkleProofHashFn(t *testing.T) {
	addressTy, err := abi.NewType("address", "address", nil)
	assert.Nil(t, err)
	uintTy, err := abi.NewType("uint256", "uint256", nil)
	assert.Nil(t, err)
	arguments := []abi.Argument{
		{Name: "addr", Type: addressTy},
		{Name: "tokenId", Type: uintTy},
	}

	hashFn := func(leaf TLeaf) ([]byte, error) {
		packed, err := abi.Arguments(arguments).Pack(leaf.Addr, leaf.TokenId)
		if err != nil {
			return nil, err
		}
		return Keccak256(packed), nil
	}

	leaves := make([]TLeaf, 4)
	leaves[0] = TLeaf{Addr: common.HexToAddress("0x1e946c284bdBb05Fb6EF41016C524E8681e3d05E"), TokenId: big.NewInt(1)}
	leaves[1] = TLeaf{Addr: common.HexToAddress("0x1D74B866598B339006160d704642459B04ba890B"), TokenId: big.NewInt(1)}
	leaves[2] = TLeaf{Addr: common.HexToAddress("0x37e948435E916069D3a1431Ddf508421073fF3E7"), TokenId: big.NewInt(1)}
	leaves[3] = TLeaf{Addr: common.HexToAddress("0x29c34A7d23B8BCBE7c5Ec94C6525b78bb5cbAf36"), TokenId: big.NewInt(1)}

	mt := NewMerkleTree(leaves, &hashFn, nil)

	root := mt.GetRoot()
	assert.NotNil(t, root)

	proof, err := mt.GetProof(leaves[0])
	assert.Nil(t, err)

	isValid, err := mt.Verify(proof, leaves[0], root)
	assert.Nil(t, err)
	assert.True(t, isValid)
}
