#include <iostream>
#include <memory>
#include <string>
#include <thread>
#include <chrono>
#include <csignal>
#include <atomic>
#include <fstream>
#include <sstream>
#include <map>

#ifdef _WIN32
#include <winsock2.h>
#include <ws2tcpip.h>
#pragma comment(lib, "ws2_32.lib")
#else
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <unistd.h>
#endif

namespace backup {
    class AgentService {
    public:
        class StubInterface {
        public:
            virtual ~StubInterface() = default;
            virtual grpc::Status Register(grpc::ClientContext* context,
                const ::backup::AgentRegistrationRequest& request,
                ::backup::AgentRegistrationResponse* response) = 0;
        };
    };
}
