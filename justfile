target := "upsmon"

default: build

build:
    cd cmd && go build -ldflags='-s -w' -o ../{{target}}

clean:
    rm -f {{target}}
