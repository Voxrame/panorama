package world_test

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/lord-server/panorama/pkg/geom"
	"github.com/lord-server/panorama/pkg/world"
	"github.com/stretchr/testify/require"
)

func TestBlock_DecodeMapBlock(t *testing.T) {
	t.Parallel()

	var data = hexToBytes(t, `
		1D28B52FFD00587D0C0002C8101AB027CB010CC3300CC322DD0840586F6269B0
		01868E5014D9320503FE5ED7BD5EAFEA55AB7ADD6DBB5B5BDB52CE8110CEC21C
		82C441610CA710380231C696074F12238140A8F1F6B2FF1D91074992C6128011
		E028982129120219661832081943042232331488049224CCB635030020B45F02
		5C428F1F1323C9E16A3DB7AF4F2DAC25F013EA5ED582E5C64D7CCFEA0279FB15
		1C6B3568843711B80EACAED4D1937124755F462A631BA646A3216850071E1684
		59D8B29511DAD50B0D9CEF43FDDB631CBEF349B9697067DF3DCE7838FF9AE360
		D5535B4C2D4FDA6783E7A9DC687BC72D370326D5ACE777144005ADBED5416C43
		06891BF6A05C08167C187FCF03273EE6CD101268FDCF677EED9D0F699FD2447F
		D0EBBF16AF71DBF7B62D88C2333F00CA7F734A2A619B1DEA4D7B1E7132CF048F
		FFB8B131FF596296DC2F8644971E689223B108BAB9325CA97D6AC96CB710FBCF
		97BAD6CFD0A70A52B23A9D3B9C44137E4439D1D1AD67D62FDB286962D0CDB983
		46CE627F78DA7BED75FC7DFC2F61A17280B721644A524118AA`)

	block, err := world.DecodeMapBlock(data)
	require.NoError(t, err)

	require.Equal(t, block.ResolveName(0), "default:cave_ice")
	require.Equal(t, block.ResolveName(1), "air")

	validNodeIDs := []uint16{0, 1}

	for z := range geom.BlockSize {
		for y := range geom.BlockSize {
			for x := range geom.BlockSize {
				node := block.GetNode(geom.NodePosition{
					X: x,
					Y: y,
					Z: z,
				})

				require.Contains(t, validNodeIDs, node.ID)
			}
		}
	}
}

func hexToBytes(t *testing.T, data string) []byte {
	t.Helper()

	data = strings.ReplaceAll(data, " ", "")
	data = strings.ReplaceAll(data, "\n", "")
	data = strings.ReplaceAll(data, "\t", "")

	out, err := hex.DecodeString(data)
	if err != nil {
		panic(err)
	}

	return out
}

func TestBlock_DecodeMapBlock_InvalidData(t *testing.T) {
	t.Parallel()

	invalidData := []byte{0xFF, 0xFF, 0xFF, 0xFF}
	_, err := world.DecodeMapBlock(invalidData)
	require.Error(t, err)
}
