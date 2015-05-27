package main

import (
	"fmt"
	"math/rand"
	"time"
	"sync"
)

var wg sync.WaitGroup

type Point struct {
	x float64
	y float64
}

func (p Point) String() string {
	return fmt.Sprintf("[%v, %v]\n", p.x, p.y)
}

func (p Point) equals(o Point) bool {
	return p.x == o.x && p.y == o.y
}

func calculateDistance(p1 Point, p2 Point) float64 {
	dx := p1.x - p2.x
	dy := p1.y - p2.y
	return dx*dx + dy*dy
}

func calculateCentroid(points []Point) Point {
	sumX := 0.0
	sumY := 0.0
	for _,v := range points {
		sumX += v.x
		sumY += v.y
	}
	return Point{sumX/float64(len(points)), sumY/float64(len(points))}
}

type Cluster struct {
	centroid Point
	points   []Point
}

type ClusterList []Cluster

func (c Cluster) String() string {
	result := "Cluster:\n"
	for _, v := range c.points {
		result += v.String()
	}
	result += "Centroid: " + c.centroid.String()
	return result
}

func (c Cluster) equals(o Cluster) bool {
	result := len(c.points) == len(o.points)
	if !result {
		// I dont want to loop through all points in this obvious case
		return result
	}
	for _, v := range c.points {
		result = result && o.contains(v)
	}
	return result
}

func (c Cluster) contains(p Point) bool {
	result := false
	for _, v := range c.points {
		result = result || v.equals(p)
	}
	return result
}

func (c ClusterList) remove(o Cluster) ClusterList {
	for i, v := range c {
		if v.equals(o) {
			result := append(c[:i], c[i+1:]...)
			return result
		}
	}
	return nil
}

type ClusterTupel struct {
	c1       Cluster
	c2       Cluster
	distance float64
}

func generatePoints(n int) []Point {
	result := make([]Point, n)
	for i := 0; i < n; i++ {
		result[i] = Point{rand.Float64() * 100, rand.Float64() * 100}
	}
	return result
}

func prepareClustering(points []Point) []Cluster {
	result := make([]Cluster, len(points))
	for i := 0; i < len(points); i++ {
		tempClusterPoint := make([]Point, 0, 1)
		clusterPoint := append(tempClusterPoint, points[i])
		result[i] = Cluster{points[i], clusterPoint}
	}
	return result
}

func findClosestTupel(tupels []ClusterTupel) ClusterTupel {
	closest := tupels[0]
	if len(tupels) >= 1 {
		for _,v := range tupels {
			if v.distance < closest.distance {
				closest = v
			}
		}
		return closest
	}
	return closest
}

func buildTupel(clusterList []Cluster, startIndex int, c chan []ClusterTupel) {
	defer wg.Done()
	result := make([]ClusterTupel, 0, len(clusterList) - startIndex)
	for ;startIndex < len(clusterList) - 1; startIndex++ {
		distance := calculateDistance(clusterList[startIndex].centroid, clusterList[startIndex+1].centroid)
		tupel := ClusterTupel{clusterList[startIndex], clusterList[startIndex+1], distance}
		result = append(result, tupel)
	} 
	c <- result
}

func findClustersToMerge(cluster []Cluster) ClusterTupel {
	// too lazy to precalculate the exact size
	tupels := make([]ClusterTupel, 0, 4950)
	// does the buffer size have to be set?
	channels := make([]chan []ClusterTupel, 0, 99)

	for i := 0; i < len(cluster) - 1; i++ {
		c := make(chan []ClusterTupel, 1)
		channels = append(channels, c)
		wg.Add(1)
		go buildTupel(cluster, i, c)
	}
	wg.Wait()
	for _,v := range channels {
		tupels = append(tupels, <-v...)
	}

	return findClosestTupel(tupels)
}

func merge(tupel ClusterTupel) Cluster {
	points := make([]Point, 0, len(tupel.c1.points) + len(tupel.c2.points))
	points = append(points, tupel.c1.points...)
	points = append(points, tupel.c2.points...)
	centroid := calculateCentroid(points)
	return Cluster{centroid, points}
}

func cluster(cluster ClusterList) ClusterList {
	for len(cluster) > 3 {
		toMerge := findClustersToMerge(cluster)
		merged := merge(toMerge)
		cluster = cluster.remove(toMerge.c1)
		cluster = cluster.remove(toMerge.c2)
		cluster = append(cluster, merged)
	}
	return cluster
}

func prettyPrint(cluster []Cluster, duration time.Duration) {
	result := fmt.Sprintf("%v resulting clusters\n", len(cluster))
	result += fmt.Sprintf("Program took %v\n", duration)
	for _,v := range cluster {
		result += "-------------------\n"
		result += v.String()
	}
	fmt.Println(result)
}

func main() {
	start := time.Now()
	points := generatePoints(100)
	preCluster := prepareClustering(points)
	clustered := cluster(preCluster)
	elapsed := time.Since(start)
	prettyPrint(clustered, elapsed)
}
