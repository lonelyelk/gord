package main

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"math/rand"

	"github.com/icza/mjpeg"
)

const (
	width      = 200
	height     = 200
	cycles     = 10000
	fps        = 12
	frameEvery = 40
	seedn      = 20
	seedr      = 5
)

type FloatNumber interface {
	float32 | float64
}

func laplacian(ambit [][]float32) float32 {
	return 0.2*(ambit[0][1]+ambit[2][1]+ambit[1][0]+ambit[1][2]) +
		0.05*(ambit[0][0]+ambit[0][2]+ambit[2][2]+ambit[2][0]) -
		ambit[1][1]
}

func make2D[T FloatNumber](n, m int) [][]T {
	matrix := make([][]T, n)
	rows := make([]T, n*m)
	for i, startRow := 0, 0; i < n; i, startRow = i+1, startRow+m {
		endRow := startRow + m
		matrix[i] = rows[startRow:endRow:endRow]
	}
	return matrix
}

func f[T FloatNumber](x, y int) T {
	return 0.01 + (0.1-0.01)*T(height-y)/height
	// return 0.01
	// return 0.0367
	// return 0.0545
}

func k[T FloatNumber](x, y int) T {
	return 0.045 + (0.07-0.045)*T(width-x)/width
	// return 0.041
	// return 0.0649
	// return 0.062
}

func toJpeg[T FloatNumber](a, b [][]T) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			if a[i][j]*0.4 >= b[i][j] {
				img.Set(j, i, color.White)
			} else {
				img.Set(j, i, color.Black)
			}
		}
	}

	buf := &bytes.Buffer{}
	err := jpeg.Encode(buf, img, nil)
	if err != nil {
		return []byte{}, err
	}
	return buf.Bytes(), nil
}

func main() {
	aw, err := mjpeg.New("test.avi", int32(width), int32(height), fps)
	if err != nil {
		log.Fatal(err)
	}
	defer aw.Close()

	a1 := make2D[float32](height, width)
	a2 := make2D[float32](height, width)
	b1 := make2D[float32](height, width)
	b2 := make2D[float32](height, width)
	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			a1[i][j] = 1
		}
	}
	for c := 0; c < seedn; c++ {
		x := rand.Intn(width)
		y := rand.Intn(height)
		for i := y - seedr; i < y+seedr; i++ {
			for j := x - seedr; j < x+seedr; j++ {
				if i < 0 || j < 0 || i >= height || j >= width {
					continue
				} else {
					b1[i][j] = 1
					a1[i][j] = 0
				}
			}
		}

	}
	jpeg, err := toJpeg(a1, b1)
	if err != nil {
		log.Fatal(err)
	}
	err = aw.AddFrame(jpeg)
	if err != nil {
		log.Fatal(err)
	}
	for c := 1; c <= cycles; c++ {
		for i := 1; i < height-1; i++ {
			for j := 1; j < width-1; j++ {
				a2[i][j] = a1[i][j] + laplacian([][]float32{
					a1[i-1][j-1 : j+2 : j+2],
					a1[i][j-1 : j+2 : j+2],
					a1[i+1][j-1 : j+2 : j+2],
				}) -
					a1[i][j]*b1[i][j]*b1[i][j] +
					f[float32](j, i)*(1-a1[i][j])
				b2[i][j] = b1[i][j] + 0.5*laplacian([][]float32{
					b1[i-1][j-1 : j+2 : j+2],
					b1[i][j-1 : j+2 : j+2],
					b1[i+1][j-1 : j+2 : j+2],
				}) +
					a1[i][j]*b1[i][j]*b1[i][j] -
					(k[float32](j, i)+f[float32](j, i))*b1[i][j]
			}
		}
		for i := 0; i < height; i++ {
			a2[i][0] = a2[i][1]
			a2[i][width-1] = a2[i][width-2]
			b2[i][0] = b2[i][1]
			b2[i][width-1] = b2[i][width-2]
		}
		for j := 0; j < width; j++ {
			a2[0][j] = a2[1][j]
			a2[height-1][j] = a2[height-2][j]
			b2[0][j] = b2[1][j]
			b2[height-1][j] = b2[height-2][j]
		}
		a1, a2 = a2, a1
		b1, b2 = b2, b1
		if c%frameEvery != 0 {
			continue
		}
		jpeg, err := toJpeg(a1, b1)
		if err != nil {
			log.Fatal(err)
		}
		err = aw.AddFrame(jpeg)
		if err != nil {
			log.Fatal(err)
		}
	}
}
