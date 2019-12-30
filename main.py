# brutal force solution in python
#coding: utf-8
import sys
M,N = map(int, sys.stdin.readline().strip().split())
class Polygon(object):
    def __init__(self, polygon):
        ring = [ (float(i[0]),float(i[1])) for i in [xy.split(',') for xy in polygon.strip().split(';')]]
        xmin, xmax, ymin, ymax = ring[0][0], ring[0][0], ring[0][1], ring[0][1]
        for x , y in ring:
            xmin = min(x, xmin)
            xmax = max(x, xmax)
            ymin = min(y, ymin)
            ymax = max(y, ymax)
        self.ring = ring
        self.xmin =  xmin
        self.xmax =  xmax
        self.ymin =  ymin
        self.ymax =  ymax
        self.N =  len(ring)

    def contain(p, pt):
        if pt[0] < p.xmin or pt[0] > p.xmax or pt[1] < p.ymin or pt[1] > p.ymax:
            return False
        inside = False
        for i in range(p.N-1):
            Pi, Pj = p.ring[i+1], p.ring[i]
            if (pt[1] < Pi[1]) != (pt[1] < Pj[1]) and (pt[0] < (Pj[0]-Pi[0])*(pt[1]-Pi[1])/(Pj[1]-Pi[1])+Pi[0]):
                inside = not inside
        return inside

polygons = {}
for i in range(M):
    id, polygon = sys.stdin.readline().split(' ')
    polygons[int(id)] = Polygon(polygon)

def find(point):
    for id, polygon in polygons.items():
        if polygon.contain(point):
            return id
    return 0

for i in range(N):
    x, y = sys.stdin.readline().split(',')
    point = (float(x),float(y))
    print(find(point))
