cmake_minimum_required(VERSION 3.16)
project(eupf C)

set(CMAKE_C_COMPILER clang)

add_library(eupf OBJECT src/xdp/ip_entrypoint.c src/xdp/gtp_entrypoint.c)

target_compile_options(eupf PRIVATE -O2 -target bpf)