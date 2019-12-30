package main

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"
	//. "github.com/dhconnelly/rtreego"
)

/************************************************
* Geometry
************************************************/
type Geometry interface {
	Bounds() *Rect
}
type Rect struct {
	p, q Point // p:min q:max
}
type Point [2]float64
type Polygon struct {
	ID                     int
	N                      int
	xmin, xmax, ymin, ymax float64
	Ring                   []Point
}

func (p Point) Bounds() *Rect {
	return &Rect{p, p}
}
func (p *Polygon) Bounds() *Rect {
	return &Rect{[2]float64{p.xmin, p.ymin}, [2]float64{p.xmax, p.ymax}}
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
	return &Polygon{id, len(ring), xmin, xmax, ymin, ymax, ring}
}

func (p *Polygon) Contains(pt Point) (inside bool) {
	if pt[0] < p.xmin || pt[0] > p.xmax || pt[1] < p.ymin || pt[1] > p.ymax {
		return
	}
	for i := 0; i < p.N-1; i++ {
		Pi, Pj := p.Ring[i+1], p.Ring[i]
		if (pt[1] < Pi[1]) != (pt[1] < Pj[1]) && (pt[0] < (Pj[0]-Pi[0])*(pt[1]-Pi[1])/(Pj[1]-Pi[1])+Pi[0]) { // core tricks here
			inside = !inside
		}
	}
	return
}

/************************************************
* RTree (Optional)
************************************************/
type Rtree struct {
	MinChildren int
	MaxChildren int
	root        *node
	height      int
}

type node struct {
	parent  *node
	leaf    bool
	entries []entry
	level   int // node depth in the Rtree
}

type entry struct {
	bb    *Rect // bounding-box of all children
	child *node
	obj   Geometry
}

func NewTree(min, max int) *Rtree {
	rt := &Rtree{
		MinChildren: min,
		MaxChildren: max,
		height:      1,
		root: &node{
			entries: []entry{},
			leaf:    true,
			level:   1,
		},
	}
	return rt
}

func (tree *Rtree) Insert(obj Geometry) {
	e := entry{obj.Bounds(), nil, obj}
	tree.insert(e, 1)
}

func (tree *Rtree) Search(obj Geometry) []Geometry {
	return tree.search([]Geometry{}, tree.root, obj.Bounds())
}

func (tree *Rtree) search(results []Geometry, n *node, bb *Rect) []Geometry {
	for _, e := range n.entries {
		if !intersect(e.bb, bb) {
			continue
		}
		if !n.leaf {
			results = tree.search(results, e.child, bb)
			continue
		}
		results = append(results, e.obj)
	}
	return results
}

// insert adds the specified entry to the tree at the specified level.
func (tree *Rtree) insert(e entry, level int) {
	leaf := tree.chooseNode(tree.root, e, level)
	leaf.entries = append(leaf.entries, e)

	// update parent pointer if necessary
	if e.child != nil {
		e.child.parent = leaf
	}

	// split leaf if overflows
	var split *node
	if len(leaf.entries) > tree.MaxChildren {
		leaf, split = leaf.split(tree.MinChildren)
	}
	root, splitRoot := tree.adjustTree(leaf, split)
	if splitRoot != nil {
		oldRoot := root
		tree.height++
		tree.root = &node{
			parent: nil,
			level:  tree.height,
			entries: []entry{
				{bb: oldRoot.computeBoundingBox(), child: oldRoot},
				{bb: splitRoot.computeBoundingBox(), child: splitRoot},
			},
		}
		oldRoot.parent = tree.root
		splitRoot.parent = tree.root
	}
}

// product of its side lengths
func (r *Rect) size() float64 {
	size := 1.0
	for i, a := range r.p {
		b := r.q[i]
		size *= b - a
	}
	return size
}

func intersect(r1, r2 *Rect) bool {
	for i := range r1.p {
		a1, b1, a2, b2 := r1.p[i], r1.q[i], r2.p[i], r2.q[i]
		if b2 <= a1 || b1 <= a2 {
			return false
		}
	}
	return true
}

// smallest rectangle containing both r1 and r2.
func boundingBox(r1, r2 *Rect) (bb *Rect) {
	bb = new(Rect)
	// x
	if r1.p[0] <= r2.p[0] {
		bb.p[0] = r1.p[0]
	} else {
		bb.p[0] = r2.p[0]
	}
	if r1.q[0] <= r2.q[0] {
		bb.q[0] = r2.q[0]
	} else {
		bb.q[0] = r1.q[0]
	}

	// y
	if r1.p[1] <= r2.p[1] {
		bb.p[1] = r1.p[1]
	} else {
		bb.p[1] = r2.p[1]
	}
	if r1.q[1] <= r2.q[1] {
		bb.q[1] = r2.q[1]
	} else {
		bb.q[1] = r1.q[1]
	}

	return
}

func (tree *Rtree) chooseNode(n *node, e entry, level int) *node {
	if n.leaf || n.level == level {
		return n
	}
	diff := math.MaxFloat64
	var chosen entry
	for _, en := range n.entries {
		bb := boundingBox(en.bb, e.bb)
		d := bb.size() - en.bb.size()
		if d < diff || (d == diff && en.bb.size() < chosen.bb.size()) {
			diff = d
			chosen = en
		}
	}
	return tree.chooseNode(chosen.child, e, level)
}

func (tree *Rtree) adjustTree(n, nn *node) (*node, *node) {
	if n == tree.root {
		return n, nn
	}
	var ent *entry
	for i := range n.parent.entries {
		if n.parent.entries[i].child == n {
			ent = &n.parent.entries[i]
			break
		}
	}
	ent.bb = n.computeBoundingBox()
	if nn == nil {
		return tree.adjustTree(n.parent, nil)
	}
	enn := entry{nn.computeBoundingBox(), nn, nil}
	n.parent.entries = append(n.parent.entries, enn)
	if len(n.parent.entries) > tree.MaxChildren {
		return tree.adjustTree(n.parent.split(tree.MinChildren))
	}
	return tree.adjustTree(n.parent, nil)
}

// computeBoundingBox finds the MBR of the children of n.
func (n *node) computeBoundingBox() (bb *Rect) {
	childBoxes := make([]*Rect, len(n.entries))
	for i, e := range n.entries {
		childBoxes[i] = e.bb
	}
	if len(childBoxes) == 1 {
		return childBoxes[0]

	}
	bb = boundingBox(childBoxes[0], childBoxes[1])
	for _, rect := range childBoxes[2:] {
		bb = boundingBox(bb, rect)
	}
	return
}

func (n *node) split(minGroupSize int) (left, right *node) {
	l, r := n.pickSeeds()
	leftSeed, rightSeed := n.entries[l], n.entries[r]
	remaining := append(n.entries[:l], n.entries[l+1:r]...)
	remaining = append(remaining, n.entries[r+1:]...)

	left = n
	left.entries = []entry{leftSeed}
	right = &node{
		parent:  n.parent,
		leaf:    n.leaf,
		level:   n.level,
		entries: []entry{rightSeed},
	}

	if rightSeed.child != nil {
		rightSeed.child.parent = right
	}
	if leftSeed.child != nil {
		leftSeed.child.parent = left
	}

	for len(remaining) > 0 {
		next := pickNext(left, right, remaining)
		e := remaining[next]

		if len(remaining)+len(left.entries) <= minGroupSize {
			assign(e, left)
		} else if len(remaining)+len(right.entries) <= minGroupSize {
			assign(e, right)
		} else {
			assignGroup(e, left, right)
		}
		remaining = append(remaining[:next], remaining[next+1:]...)
	}
	return
}

func assign(e entry, group *node) {
	if e.child != nil {
		e.child.parent = group
	}
	group.entries = append(group.entries, e)
}

// assignGroup chooses one of two groups to which a node should be added.
func assignGroup(e entry, left, right *node) {
	leftBB := left.computeBoundingBox()
	rightBB := right.computeBoundingBox()
	leftEnlarged := boundingBox(leftBB, e.bb)
	rightEnlarged := boundingBox(rightBB, e.bb)

	leftDiff := leftEnlarged.size() - leftBB.size()
	rightDiff := rightEnlarged.size() - rightBB.size()
	if diff := leftDiff - rightDiff; diff < 0 {
		assign(e, left)
		return
	} else if diff > 0 {
		assign(e, right)
		return
	}
	if diff := leftBB.size() - rightBB.size(); diff < 0 {
		assign(e, left)
		return
	} else if diff > 0 {
		assign(e, right)
		return
	}
	if diff := len(left.entries) - len(right.entries); diff <= 0 {
		assign(e, left)
		return
	}
	assign(e, right)
}

func (n *node) pickSeeds() (int, int) {
	left, right := 0, 1
	maxWastedSpace := -1.0
	for i, e1 := range n.entries {
		for j, e2 := range n.entries[i+1:] {
			d := boundingBox(e1.bb, e2.bb).size() - e1.bb.size() - e2.bb.size()
			if d > maxWastedSpace {
				maxWastedSpace = d
				left, right = i, j+i+1
			}
		}
	}
	return left, right
}

func pickNext(left, right *node, entries []entry) (next int) {
	maxDiff := -1.0
	leftBB := left.computeBoundingBox()
	rightBB := right.computeBoundingBox()
	for i, e := range entries {
		d1 := boundingBox(leftBB, e.bb).size() - leftBB.size()
		d2 := boundingBox(rightBB, e.bb).size() - rightBB.size()
		d := math.Abs(d1 - d2)
		if d > maxDiff {
			maxDiff = d
			next = i
		}
	}
	return
}

/************************************************
* GeoEncoder
************************************************/
type GeoEncoder struct {
	Tree *Rtree
}

func NewGeoEncoder(polygons []*Polygon) *GeoEncoder {
	tree := NewTree(25, 50)
	for _, p := range polygons {
		tree.Insert(p)
	}
	return &GeoEncoder{tree}
}

func (e *GeoEncoder) Encode(p Point) int {
	res := e.Tree.Search(p)
	if len(res) == 0 {
		return 0
	}
	for _, item := range res {
		if polygon := item.(*Polygon); polygon.Contains(p) {
			return polygon.ID
		}
	}
	return 0
}

/************************************************
* Run
************************************************/
func Run() {
	var M, N int
	var x, y float64
	var launch, begin time.Time
	var parsePolygon, parsePoint, buildIndex, encodePoint, writeRes, totalTime time.Duration
	launch = time.Now()

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
	totalTime = time.Now().Sub(launch)

	// summary
	_, _, _, _, _, _ = parsePolygon, parsePoint, buildIndex, encodePoint, writeRes, totalTime
	fmt.Fprintf(os.Stderr, "total time:  \t%v\n", totalTime)
	fmt.Fprintf(os.Stderr, "parse poly:  \t%v\n", parsePolygon)
	fmt.Fprintf(os.Stderr, "parse point: \t%v\n", parsePoint)
	fmt.Fprintf(os.Stderr, "build index: \t%v\n", buildIndex)
	fmt.Fprintf(os.Stderr, "encode point:\t%v\n", encodePoint)
	fmt.Fprintf(os.Stderr, "write result:\t%v\n", writeRes)
	fmt.Fprintf(os.Stderr, "query time  :\t%v\n", time.Duration(float64(encodePoint.Nanoseconds())/float64(N)))
}

func main() {
	Run()
}
