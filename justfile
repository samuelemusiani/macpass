all: build_dir macpass macpassd

macpass:
  go build -o ./build/macpass ./macpass

macpassd:
  go build -o ./build/macpassd ./macpassd

build_dir:
  mkdir -p build
