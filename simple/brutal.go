package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Point struct {
	X float64
	Y float64
}

type Polygon struct {
	ID   int
	Ring []Point
	N    int
	xmin float64
	xmax float64
	ymin float64
	ymax float64
}

func NewPolygon(r *bufio.Reader) *Polygon {
	var ring []Point
	var x, y float64
	var xy []byte
	idStr, _ := r.ReadBytes(' ')
	id, _ := strconv.Atoi(string(idStr[:len(idStr)-1]))
	polyStr, _ := r.ReadBytes('\n')
	coords := bytes.Split(polyStr[:len(polyStr)-1], []byte(";"))
	xy = coords[0]
	sep := bytes.IndexByte(xy, ',')
	x, _ = strconv.ParseFloat(string(xy[:sep]), 64)
	y, _ = strconv.ParseFloat(string(xy[sep+1:]), 64)
	xmin, xmax, ymin, ymax := x, x, y, y
	for _, xy = range coords {
		sep = bytes.IndexByte(xy, ',')
		x, _ = strconv.ParseFloat(string(xy[:sep]), 64)
		y, _ = strconv.ParseFloat(string(xy[sep+1:]), 64)
		ring = append(ring, Point{x, y})
		if x < xmin {
			xmin = x
		}
		if x > xmax {
			xmax = x
		}
		if y < ymin {
			ymin = y
		}
		if y > ymax {
			ymax = y
		}
	}
	return &Polygon{id, ring, len(ring), xmin, xmax, ymin, ymax}
}

func (p *Polygon) Contains(pt Point) (inside bool) {
	if pt.X < p.xmin || pt.X > p.xmax || pt.Y < p.ymin || pt.Y > p.ymax {
		return
	}
	for i := 0; i < p.N-1; i++ {
		Pi, Pj := p.Ring[i+1], p.Ring[i]
		if (pt.Y < Pi.Y) != (pt.Y < Pj.Y) && (pt.X < (Pj.X-Pi.X)*(pt.Y-Pi.Y)/(Pj.Y-Pi.Y)+Pi.X) { // tricks here
			inside = !inside
		}
	}
	return
}

type GeoEncoder struct {
	Data []*Polygon
}

func NewGeoEncoder(polygons []*Polygon) *GeoEncoder {
	return &GeoEncoder{polygons}
}

func (e *GeoEncoder) Encode(p Point) int {
	for _, polygon := range e.Data {
		if polygon.Contains(p) {
			return polygon.ID
		}
	}
	return 0
}

func main() {
	Run()
}

func Run() {
	var M, N int
	var x, y float64
	var begin time.Time
	var parsePolygon, parsePoint, buildIndex, encodePoint, writeRes time.Duration
	reader := bufio.NewReaderSize(os.Stdin, 64*4096)
	fmt.Fscanf(reader, "%d %d\n", &M, &N)
	polygons := make([]*Polygon, M)
	points := make([]Point, N)

	begin = time.Now()
	for i := 0; i < M; i++ {
		polygons[i] = NewPolygon(reader)
	}
	parsePolygon = time.Now().Sub(begin)

	begin = time.Now()
	for i := 0; i < N; i++ {
		xy, _ := reader.ReadBytes('\n')
		sep := bytes.IndexByte(xy, ',')
		x, _ = strconv.ParseFloat(string(xy[:sep]), 64)
		y, _ = strconv.ParseFloat(string(xy[sep+1:len(xy)-1]), 64)
		points[i] = Point{x, y}
	}
	parsePoint = time.Now().Sub(begin)

	begin = time.Now()
	encoder := NewGeoEncoder(polygons)
	buildIndex = time.Now().Sub(begin)

	var buf bytes.Buffer
	begin = time.Now()
	for _, p := range points {
		buf.WriteString(strconv.Itoa(encoder.Encode(p)))
		buf.WriteByte('\n')
	}
	encodePoint = time.Now().Sub(begin)

	begin = time.Now()
	os.Stdout.Write(buf.Bytes())
	writeRes = time.Now().Sub(begin)

	// summary
	_, _, _, _, _ = parsePolygon, parsePoint, buildIndex, encodePoint, writeRes
	// fmt.Fprintf(os.Stderr, "parse poly:  \t%v\n", parsePolygon)
	// fmt.Fprintf(os.Stderr, "parse point: \t%v\n", parsePoint)
	// fmt.Fprintf(os.Stderr, "build index: \t%v\n", buildIndex)
	// fmt.Fprintf(os.Stderr, "encode point:\t%v\n", encodePoint)
	// fmt.Fprintf(os.Stderr, "write result:\t%v\n", writeRes)
	// fmt.Fprintf(os.Stderr, "query time  :\t%v\n", time.Duration(float64(encodePoint.Nanoseconds())/float64(N)))
}
