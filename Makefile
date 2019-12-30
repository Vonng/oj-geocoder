POINTSIZE=200000

build:
	CGO_ENABLED=0 go build  -a -ldflags '-extldflags "-static"'

test:
	GOGC=off go run main.go < case1.in

test2:
	GOGC=off go run main.go < case1.in
	md5 res
	md5 case1.out
	diff res case1.out

check1:
	GOGC=off go run main.go < case/case1.in > res
	md5 res
	md5 case/case1.out

check2:build
	GOGC=off ./geo < case/case2.in > res
	md5 res
	md5 case/case2.out

t:build
	GOGC=off GOMAXPROCS=1 ./geo < case/case2.in > res
	md5 res
	md5 case/case2.out

run:build
	GOGC=off ./geo < case2.in > res

runp:build
	GOGC=off ./geo -p=profile < case/case2.in > /dev/null

prof:
	go tool pprof profile

clean:
	rm -rf  geoencoder res.txt

case1:
	echo "15 20" > case1.in
	psql data -1qAtxc "copy (SELECT id || ' ' ||  trim(replace(replace(((ST_AsGeoJson(fence)::JSONB)#>'{coordinates,0}')::TEXT,' ',''),'],[',';'),'[]') FROM geo.fences WHERE id / 10000 * 10000 = 110000) to stdout" >> case1.in
	psql data -1qAtxc "copy (SELECT point FROM geo.points WHERE adcode / 10000 * 10000 = 110000 ORDER BY id LIMIT 20) TO STDOUT" >>  case1.in
	psql data -1qAtxc "copy (SELECT adcode FROM geo.points WHERE adcode / 10000 * 10000 = 110000 ORDER BY id LIMIT 20) TO STDOUT" > case1.out

case2:
	echo "2819 $(POINTSIZE)" > case2.in
	psql data -1qAtxc "copy (SELECT id || ' ' ||  trim(replace(replace(((ST_AsGeoJson(fence)::JSONB)#>'{coordinates,0}')::TEXT,' ',''),'],[',';'),'[]') FROM geo.fences) to stdout" >> case2.in
	psql data -1qAtxc "copy (SELECT point FROM geo.points ORDER BY id LIMIT $(POINTSIZE)) TO STDOUT" >>  case2.in
	psql data -1qAtxc "copy (SELECT adcode FROM geo.points ORDER BY id LIMIT $(POINTSIZE)) TO STDOUT" > case2.out


.Phony: fence, point, input, input2, clean, run, runp
