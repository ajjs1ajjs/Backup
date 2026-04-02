cmake_minimum_required(VERSION 3.20)
project(Backup.Agent VERSION 1.0.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

option(BUILD_TESTS "Build tests" ON)
option(BUILD_SHARED_LIBS "Build shared libraries" OFF)

if(NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE Release)
endif()

find_package(Threads REQUIRED)

set(COMMON_SOURCES
    main.cpp
    config.cpp
)

set(AGENT_SOURCES
    ${COMMON_SOURCES}
    hyperv_agent.cpp
    vmware_agent.cpp
    kvm_agent.cpp
    database_agent.cpp
    transfer.cpp
    compression.cpp
)

add_executable(backup-agent ${AGENT_SOURCES})

target_compile_options(backup-agent PRIVATE
    $<$<CXX_COMPILER_ID:MSVC>:/W4 /EHsc>
    $<$<CXX_COMPILER_ID:GNU>:-Wall -Wextra -pedantic>
    $<$<CXX_COMPILER_ID:Clang>:-Wall -Wextra -pedantic>
)

if(BUILD_TESTS)
    enable_testing()
    add_subdirectory(tests)
endif()

install(TARGETS backup-agent DESTINATION bin)

set(CPACK_GENERATOR "NSIS;DEB;RPM")
set(CPACK_PACKAGE_VERSION_MAJOR ${PROJECT_VERSION_MAJOR})
set(CPACK_PACKAGE_VERSION_MINOR ${PROJECT_VERSION_MINOR})
set(CPACK_PACKAGE_VERSION_PATCH ${PROJECT_VERSION_PATCH})
include(CPack)
