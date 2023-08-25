package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type rleResponse struct {
	initialiseData []coords
	sizeX          int
	sizeY          int
	info           string
}

func rleDecoder(seed string) rleResponse {
	fileName := "./seeds/" + seed + ".rle"
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	var fileLines []string

	for scanner.Scan() {
		fileLines = append(fileLines, scanner.Text())
	}

	f.Close()

	var paramLineIndex int
	var endLineIndex int
	for n, line := range fileLines {
		if line[0] == 'x' {
			paramLineIndex = n
		}
		endLineIndex = n
	}
	var sizeX int
	var sizeY int
	b := false
	x0 := false
	y0 := false
	t0 := 0
	y0 = !!y0
	for _, char := range fileLines[paramLineIndex] {
		if char == 'x' {
			x0 = true
		}
		if char == 'y' {
			y0 = true
		}
		if char == '=' {
			b = true
		}
		if char >= '0' && char <= '9' && b {
			t0 = t0*10 + int(char-48)
		}
		if char == ',' {
			if x0 {
				sizeX = t0
			} else {
				sizeY = t0
			}
			x0 = false
			y0 = false
			b = false
			t0 = 0
		}
	}
	var initialiseArray []coords
	x := 0
	y := 0
	t := 0
	for l := paramLineIndex + 1; l < endLineIndex+1; l++ {
		for _, char := range fileLines[l] {
			if char >= '0' && char <= '9' {
				if t == 0 {
					t = int(char - 48)
				} else {
					t = (t * 10) + int(char-48)
				}
			}
			if char == 'b' {
				if t == 0 {
					x++
				} else {
					for i := 0; i < t; i++ {
						x++
					}
					t = 0
				}
			}
			if char == 'o' {
				if t == 0 {
					initialiseArray = append(initialiseArray, coords{x, y})
					x++
				} else {
					for i := 0; i < t; i++ {
						initialiseArray = append(initialiseArray, coords{x, y})
						x++
					}
					t = 0
				}
			}
			if char == '$' {
				if t == 0 {
					x = 0
					y++
				} else {
					x = 0
					for i := 0; i < t; i++ {
						y++
					}
					t = 0
				}
			}
		}
	}
	f1, err := os.Open("seedInfo.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f1.Close()

	var seedInfo map[string]string

	decoder := json.NewDecoder(f1)
	if err := decoder.Decode(&seedInfo); err != nil {
		log.Fatal(err)
	}

	fmt.Println(seedInfo[seed])

	return rleResponse{initialiseArray, sizeX, sizeY, seedInfo[seed]}
}
