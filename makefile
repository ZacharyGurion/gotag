UNAME_S := $(shell uname -s)

ifeq ($(OS),Windows_NT)  # For cmd.exe
    EXT = .exe
    PKG_CFLAGS = -IC:/path/to/taglib/include
else
    EXT =
    PKG_CFLAGS = `pkg-config --cflags taglib`
endif

CXX = g++
CPP_WRAPPER = wrapper.cpp
H_WRAPPER = wrapper.h
O_WRAPPER = wrapper.o
EXEC = gotag$(EXT)
STD = c++23

all: build-cpp build-go

build-cpp:
	$(CXX) -c $(CPP_WRAPPER) -o $(O_WRAPPER) -std=$(STD) $(PKG_CFLAGS)

build-go:
	go build -o $(EXEC) main.go

run:
	./$(EXEC) test.mp3