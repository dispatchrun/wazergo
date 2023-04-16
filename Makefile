.PHONY: clean test testdata

testdata.wat = $(wildcard testdata/*.wat)
testdata.wasm = $(testdata.wat:.wat=.wasm)

test: testdata
	go test -v ./...

testdata: $(testdata.wasm)

clean:
	rm -f $(testdata.wasm)

%.wasm: %.wat
	wat2wasm -o $@ $<
