package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Point struct {
	x float64
	y float64
}

type PointList []Point

func (p Point) String() string {
	return fmt.Sprintf("[%v, %v]\n", p.x, p.y)
}
func (p Point) equals(o Point) bool {
	return p.x == o.x && p.y == o.y
}
func (pList PointList) remove(p Point) PointList {
	for i, v := range pList {
		if v.equals(p) {
			result := append(pList[:i], pList[i+1:]...)
			return result
		}
	}
	return nil
}
func calculateDistance(p1 Point, p2 Point) float64 {
	dx := p1.x - p2.x
	dy := p1.y - p2.y
	return dx*dx + dy*dy
}
func calculateCentroid(points []Point) Point {
	sumX := 0.0
	sumY := 0.0
	for _, v := range points {
		sumX += v.x
		sumY += v.y
	}
	return Point{sumX / float64(len(points)), sumY / float64(len(points))}
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

type ClusterPointDistance struct {
	c        Cluster
	p        Point
	distance float64
}

func generatePoints(n int) []Point {
	result := make([]Point, n)
	for i := 0; i < n; i++ {
		result[i] = Point{rand.Float64() * 100, rand.Float64() * 100}
	}
	return result
}
func hier_prepareClustering(points []Point) []Cluster {
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
		for _, v := range tupels {
			if v.distance < closest.distance {
				closest = v
			}
		}
		return closest
	}
	return closest
}
func buildTupel(clusterList []Cluster, startIndex int, c chan []ClusterTupel) {
	result := make([]ClusterTupel, 0, len(clusterList)-startIndex)
	for ; startIndex < len(clusterList)-1; startIndex++ {
		distance := calculateDistance(clusterList[startIndex].centroid, clusterList[startIndex+1].centroid)
		tupel := ClusterTupel{clusterList[startIndex], clusterList[startIndex+1], distance}
		result = append(result, tupel)
	}
	c <- result
}
func calculateCombinationAmount(size int) int {
	if size <= 1 {
		return 0
	} else {
		return calculateCombinationAmount(size-1) + size - 1
	}
}
func findClustersToMerge(cluster []Cluster) ClusterTupel {
	tupels := make([]ClusterTupel, 0, calculateCombinationAmount(len(cluster)))
	// does the buffer size have to be set?
	channels := make([]chan []ClusterTupel, 0, 99)
	for i := 0; i < len(cluster)-1; i++ {
		c := make(chan []ClusterTupel, 1)
		channels = append(channels, c)
		go buildTupel(cluster, i, c)
	}
	for _, v := range channels {
		tupels = append(tupels, <-v...)
	}
	return findClosestTupel(tupels)
}
func mergeClusterInCluster(tupel ClusterTupel) Cluster {
	points := make([]Point, 0, len(tupel.c1.points)+len(tupel.c2.points))
	points = append(points, tupel.c1.points...)
	points = append(points, tupel.c2.points...)
	centroid := calculateCentroid(points)
	return Cluster{centroid, points}
}
func hier_cluster(points []Point, targetAmount int) ClusterList {
	cluster := ClusterList(hier_prepareClustering(points))
	for len(cluster) > targetAmount {
		toMerge := findClustersToMerge(cluster)
		merged := mergeClusterInCluster(toMerge)
		cluster = cluster.remove(toMerge.c1)
		cluster = cluster.remove(toMerge.c2)
		cluster = append(cluster, merged)
	}
	return cluster
}
func prettyPrint(cluster []Cluster, duration time.Duration) {
	result := fmt.Sprintf("%v resulting clusters\n", len(cluster))
	result += fmt.Sprintf("Program took %v\n", duration)
	for _, v := range cluster {
		result += "-------------------\n"
		result += v.String()
	}
	fmt.Println(result)
}
func k_preCluster(points PointList, k int) (ClusterList, PointList) {
	// we are preclustering via random-picking
	selected := make([]Point, 0, k)
	result := make([]Cluster, 0, k)
	if k == 0 {
		return result, points
	}
	start := points[rand.Int()%len(points)]
	points = points.remove(start)
	selected = append(selected, start)
	for len(selected) < k {
		maxDistance := 0.0
		var p Point
		for _, v := range points {
			for _, x := range selected {
				distance := calculateDistance(x, v)
				if distance > maxDistance {
					maxDistance = distance
					p = v
				}
			}

		}
		selected = append(selected, p)
		points = points.remove(p)
	}
	for _, v := range selected {
		clusterPoints := make([]Point, 0, 5)
		clusterPoints = append(clusterPoints, v)
		cluster := Cluster{v, clusterPoints}
		result = append(result, cluster)
	}
	return result, points
}
func k_calculateDistancesToClusters(cluster []Cluster, points []Point) []ClusterPointDistance {
	result := make([]ClusterPointDistance, 0, len(cluster)*len(points))
	for i := 0; i < len(cluster); i++ {
		for _, v := range points {
			distance := calculateDistance(cluster[i].centroid, v)
			cpd := ClusterPointDistance{cluster[i], v, distance}
			result = append(result, cpd)
		}
	}
	return result
}
func k_cluster(points PointList, k int) []Cluster {
	cluster, points := k_preCluster(points, k)
	for len(points) > 0 {
		distances := k_calculateDistancesToClusters(cluster, points)
		var minDistance ClusterPointDistance
		if len(distances) >= 1 {
			minDistance = distances[0]
		}
		for _, v := range distances {
			if minDistance.distance > v.distance {
				minDistance = v
			}
		}
		
		merged := mergePointInCluster(minDistance.c, minDistance.p)
		cluster = cluster.remove(minDistance.c)
		cluster = append(cluster, merged)
		points = points.remove(minDistance.p)
	}
	return cluster
}
func mergePointInCluster(c Cluster, p Point) Cluster {
	newPoints := append(c.points, p)
	newCentroid := calculateCentroid(c.points)
	return Cluster{newCentroid, newPoints}
}
func main() {
	points := generatePoints(100)
	start := time.Now()
	hier_clustered := hier_cluster(points, 3)
	elapsed := time.Since(start)
	prettyPrint(hier_clustered, elapsed)
	start = time.Now()
	k_clustered := k_cluster(PointList(points), 3)
	elapsed = time.Since(start)
	prettyPrint(k_clustered, elapsed)
}
