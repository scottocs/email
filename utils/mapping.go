package utils

import (
	"email/compile/contract"
	"github.com/fentec-project/bn256"
	"math/big"
)

func G1ArrToPoints(points []bn256.G1) []contract.EmailG1Point {
	arr := make([]contract.EmailG1Point, len(points))
	for i := 0; i < len(points); i++ {
		arr[i] = G1ToPoint(&points[i])
	}
	return arr
}
func G2ArrToPoints(points []bn256.G2) []contract.EmailG2Point {
	arr := make([]contract.EmailG2Point, len(points))
	for i := 0; i < len(points); i++ {
		arr[i] = G2ToPoint(&points[i])
	}
	return arr
}
func G1ToPoint(point *bn256.G1) contract.EmailG1Point {
	// Marshal the G1 point to get the X and Y coordinates as bytes
	pointBytes := point.Marshal()
	x := new(big.Int).SetBytes(pointBytes[:32])
	y := new(big.Int).SetBytes(pointBytes[32:64])

	g1Point := contract.EmailG1Point{
		X: x,
		Y: y,
	}
	return g1Point
}

func G2ToPoint(point *bn256.G2) contract.EmailG2Point {
	// Marshal the G1 point to get the X and Y coordinates as bytes
	pointBytes := point.Marshal()
	//fmt.Println(point.Marshal())

	// Create big.Int for X and Y coordinates
	a1 := new(big.Int).SetBytes(pointBytes[:32])
	a2 := new(big.Int).SetBytes(pointBytes[32:64])
	b1 := new(big.Int).SetBytes(pointBytes[64:96])
	b2 := new(big.Int).SetBytes(pointBytes[96:128])

	g2Point := contract.EmailG2Point{
		X: [2]*big.Int{a1, a2},
		Y: [2]*big.Int{b1, b2},
	}
	return g2Point
}
func PointsToG1(points []contract.EmailG1Point) []bn256.G1 {
	arr := make([]bn256.G1, len(points))
	for i := 0; i < len(points); i++ {
		arr[i] = *PointToG1(points[i])
	}
	return arr
}
func PointsToG2(points []contract.EmailG2Point) []bn256.G2 {
	arr := make([]bn256.G2, len(points))
	for i := 0; i < len(points); i++ {
		arr[i] = *PointToG2(points[i])
	}
	return arr
}
func PointToG1(point contract.EmailG1Point) *bn256.G1 {
	combinedByteArray := make([]byte, 64)
	point.X.FillBytes(combinedByteArray[:32])
	point.Y.FillBytes(combinedByteArray[32:])

	g1 := new(bn256.G1)
	g1.Unmarshal(combinedByteArray)
	return g1

}
func PointToG2(point contract.EmailG2Point) *bn256.G2 {
	combinedByteArray := make([]byte, 128)
	point.X[0].FillBytes(combinedByteArray[:32])
	point.X[1].FillBytes(combinedByteArray[32:64])
	point.Y[0].FillBytes(combinedByteArray[64:96])
	point.Y[1].FillBytes(combinedByteArray[96:128])
	g2 := new(bn256.G2)
	g2.Unmarshal(combinedByteArray)
	return g2

}
