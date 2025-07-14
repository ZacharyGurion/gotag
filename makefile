CXX=g++
CPP_WRAPPER=wrapper.cpp
H_WRAPPER=wrapper.h
O_WRAPPER=wrapper.o
EXEC=gotag
STD=c++23

all: build-cpp build-go

build-cpp:
	@$(CXX) -c $(CPP_WRAPPER) -o $(O_WRAPPER) -std=$(STD) `pkg-config --cflags taglib`

build-go:
	go build -o $(EXEC) main.go

run:
	./$(EXEC) test.mp3
