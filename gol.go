package main

import (
	"sync"
)

var threads int = 8

type params struct {
	height int
	width  int
}

type coords struct {
	X int
	Y int
}

func calculateNextState(p params, start int, end int, world [][]int, out chan<- [][]int, quickOut chan<- coords) {
	splitHeight := end - start + 1
	newWorld := make([][]int, splitHeight)
	for y := 0; y < splitHeight; y++ {
		newWorld[y] = make([]int, p.width)
		for x := 0; x < p.width; x++ {
			newWorld[y][x] = updateCell(p, world, x, y+start, quickOut)
		}
	}
	out <- newWorld
}

func updateCell(p params, world [][]int, x int, y int, quickOut chan<- coords) int {
	aliveCells := countAliveCellsAdjacent(p, world, x, y)
	result := 0
	if world[y][x] == 1 {
		if aliveCells < 2 {
			result = 0
			quickOut <- coords{x, y}
		} else if aliveCells == 3 || aliveCells == 2 {
			result = 1
		} else {
			result = 0
			quickOut <- coords{x, y}
		}
	} else {
		if aliveCells == 3 {
			result = 1
			quickOut <- coords{x, y}
		} else {
			result = 0
		}
	}
	return int(result)
}

func countAliveCellsAdjacent(p params, world [][]int, x int, y int) int {
	left, right, up, down := x-1, x+1, y-1, y+1
	count := 0
	if x == 0 {
		left = p.width - 1
	}
	if x == p.width-1 {
		right = 0
	}
	if y == 0 {
		up = p.height - 1
	}
	if y == p.height-1 {
		down = 0
	}
	count += int(world[up][left]) + int(world[up][x]) + int(world[up][right]) +
		int(world[y][left]) + int(world[y][right]) +
		int(world[down][left]) + int(world[down][x]) + int(world[down][right])
	return count
}

func quickManager(GOLRunnerChannel chan<- []coords, quickOutChannel <-chan coords, wg *sync.WaitGroup) {
	defer wg.Done()
	var quickData []coords
	for c := range quickOutChannel {
		quickData = append(quickData, c)
	}
	GOLRunnerChannel <- quickData
}

func GOLRunner(world [][]int, turnStart int, turns int) data {
	var p params
	worldData := make(map[int][][]int)
	quickData := make(map[int][]coords)
	golRunnerChannel := make(chan []coords, 1)
	p.height = len(world)
	p.width = len(world[0])

	workerChannels := make([]chan [][]int, threads)
	for i := range workerChannels {
		workerChannels[i] = make(chan [][]int)
	}

	splitSize := p.height / threads

	for turn := 0; turn < turns; turn++ {
		quickOut := make(chan coords, 100)
		var wg sync.WaitGroup
		wg.Add(1)
		go quickManager(golRunnerChannel, quickOut, &wg)
		diff := p.height % threads
		pos := 0
		for i := 0; i < threads; i++ {
			channel := workerChannels[i]
			start := pos
			pos += splitSize - 1
			if diff != 0 {
				pos += 1
				diff--
			}
			end := pos
			pos++
			go calculateNextState(p, start, end, world, channel, quickOut)
		}

		newWorld := make([][]int, 0)
		for _, w := range workerChannels {
			splitWorld := <-w
			for _, row := range splitWorld {
				newWorld = append(newWorld, row)
			}
		}

		close(quickOut)
		wg.Wait()

		worldData[turn+turnStart] = newWorld
		world = newWorld
		quickData[turn+turnStart] = <-golRunnerChannel
	}

	for _, w := range workerChannels {
		close(w)
	}
	close(golRunnerChannel)
	return data{worldData, quickData}
}
