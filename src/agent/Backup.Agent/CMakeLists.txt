cmake_minimum_required(VERSION 3.20)
project(Backup.Agent VERSION 1.0.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

option(BUILD_TESTS "Build tests" OFF)
option(BUILD_SHARED_LIBS "Build shared libraries" OFF)

if(NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE Release)
endif()

find_package(Threads REQUIRED)
find_package(Zstd QUIET)

include_directories(
    ${CMAKE_CURRENT_SOURCE_DIR}
    ${CMAKE_CURRENT_SOURCE_DIR}/core
    ${CMAKE_CURRENT_SOURCE_DIR}/hyperv
    ${CMAKE_CURRENT_SOURCE_DIR}/vmware
    ${CMAKE_CURRENT_SOURCE_DIR}/kvm
    ${CMAKE_CURRENT_SOURCE_DIR}/database
)

set(AGENT_SOURCES
    main.cpp
    hyperv/hyperv_agent.cpp
    vmware/vmware_agent.cpp
    kvm/kvm_agent.cpp
    database/database_agent.cpp
    core/data_mover.cpp
    core/compression.cpp
    core/cbt.cpp
)

add_executable(backup-agent ${AGENT_SOURCES})

target_link_libraries(backup-agent PRIVATE
    Threads::Threads
)

if(Zstd_FOUND)
    target_link_libraries(backup-agent PRIVATE Zstd::libzstd)
endif()

target_compile_options(backup-agent PRIVATE
    $<$<CXX_COMPILER_ID:MSVC>:/W4 /EHsc>
    $<$<CXX_COMPILER_ID:GNU>:-Wall -Wextra>
    $<$<CXX_COMPILER_ID:Clang>:-Wall -Wextra>
)

install(TARGETS backup-agent DESTINATION bin)
