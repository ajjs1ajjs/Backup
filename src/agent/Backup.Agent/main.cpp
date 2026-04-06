#include <iostream>
#include <memory>
#include <string>
#include <vector>
#include <fstream>
#include <cstring>
#include <thread>
#include <chrono>

#include "hyperv/hyperv_agent.h"
#include "vmware/vmware_agent.h"
#include "kvm/kvm_agent.h"

namespace backup {

class AgentFactory {
public:
    static std::unique_ptr<IHypervisorAgent> Create(const std::string& type) {
        if (type == "hyperv") {
            return CreateHyperVAgent();
        } else if (type == "vmware") {
            return CreateVMwareAgent();
        } else if (type == "kvm") {
            return CreateKVMAgent();
        }
        return nullptr;
    }
};

}

std::string config_file;
std::string command;
std::string agent_type = "hyperv"; // Значення за замовчуванням

void print_help() {
    std::cout << "Backup Agent v1.0.0" << std::endl;
    std::cout << "Usage: backup-agent <command> [options]" << std::endl;
    std::cout << "Options:" << std::endl;
    std::cout << "  --config, -c <file>   - Path to config file" << std::endl;
    std::cout << "  --type, -t <type>     - Hypervisor type (hyperv, vmware, kvm)" << std::endl;
    std::cout << "Commands:" << std::endl;
    std::cout << "  daemon    - Run as daemon (default)" << std::endl;
    std::cout << "  backup    - Start backup operation" << std::endl;
    std::cout << "  restore   - Start restore operation" << std::endl;
    std::cout << "  list      - List VMs" << std::endl;
    std::cout << "  version   - Show version" << std::endl;
    std::cout << "  help      - Show help" << std::endl;
}

void parse_args(int argc, char* argv[]) {
    for (int i = 1; i < argc; i++) {
        std::string arg = argv[i];
        if (arg == "--config" || arg == "-c") {
            if (i + 1 < argc) {
                config_file = argv[i + 1];
                i++;
            }
        } else if (arg == "--type" || arg == "-t") {
            if (i + 1 < argc) {
                agent_type = argv[i + 1];
                i++;
            }
        } else if (command.empty()) {
            command = arg;
        }
    }
}

int main(int argc, char* argv[]) {
    parse_args(argc, argv);
    
    auto agent = backup::AgentFactory::Create(agent_type);
    if (!agent && command != "version" && command != "help") {
        std::cerr << "Error: Unsupported agent type: " << agent_type << std::endl;
        return 1;
    }

    if (command.empty() || command == "daemon") {
        std::cout << "Agent (" << agent_type << ") running in daemon mode..." << std::endl;
        // Тут повинна бути логіка gRPC сервера та серцебиття
        while (true) {
            std::this_thread::sleep_for(std::chrono::seconds(30));
            std::cout << "Heartbeat sent" << std::endl;
        }
        return 0;
    }

    if (command == "version" || command == "--version" || command == "-v") {
        std::cout << "Backup Agent v1.0.0" << std::endl;
        return 0;
    }

    if (command == "help" || command == "--help" || command == "-h") {
        print_help();
        return 0;
    }

    if (command == "list") {
        std::cout << "Listing VMs for " << agent_type << "..." << std::endl;
        auto vms = agent->ListVMs();
        for (const auto& vm : vms) {
            std::cout << " - " << vm.name << " (" << vm.vm_id << ") [" << vm.status << "]" << std::endl;
        }
        return 0;
    }

    if (command == "backup") {
        std::cout << "Starting backup for " << agent_type << "..." << std::endl;
        // Спрощена логіка бекапу (створення снапшоту)
        if (agent->CreateSnapshot("test-vm", "manual-backup")) {
            std::cout << "Snapshot created successfully" << std::endl;
            return 0;
        }
        return 1;
    }

    std::cout << "Unknown command: " << command << std::endl;
    print_help();
    return 1;
}
