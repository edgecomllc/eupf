cmake_minimum_required(VERSION 3.16)
project(eupf C)

set(CMAKE_C_COMPILER clang)

add_library(eupf OBJECT src/xdp/ip_entrypoint.c 
                        src/xdp/gtp_entrypoint.c
                        src/xdp/qer_program.c
                        src/xdp/far_program.c
                        src/xdp/upf_program.c)

target_compile_options(eupf PRIVATE -O2 -target bpf)

# some tricks for dropbox/ebpf library
target_include_directories(eupf PRIVATE src)
target_compile_definitions(eupf PRIVATE __linux__)
target_include_directories(eupf PRIVATE ~/go/pkg/mod/github.com/dropbox/goebpf@v0.0.0-20221212193212-e568275f8431)