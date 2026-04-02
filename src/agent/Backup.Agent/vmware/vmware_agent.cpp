#include "vmware_agent.h"

namespace backup {

VMwareAgent::VMwareAgent() = default;
VMwareAgent::~VMwareAgent() = default;

bool VMwareAgent::Connect(const std::string& host, const std::string& username, const std::string& password) {
    connected_ = true;
    host_ = host;
    return true;
}

void VMwareAgent::Disconnect() {
    connected_ = false;
}

bool VMwareAgent::IsConnected() const {
    return connected_;
}

std::vector<VMwareVM> VMwareAgent::ListVMs() {
    return cached_vms_;
}

VMwareVM* VMwareAgent::GetVM(const std::string& vm_id) {
    for (auto& vm : cached_vms_) {
        if (vm.vm_id == vm_id) return &vm;
    }
    return nullptr;
}

bool VMwareAgent::CreateSnapshot(const std::string& vm_id, const std::string& name) {
    return true;
}

bool VMwareAgent::DeleteSnapshot(const std::string& vm_id, const std::string& snapshot_id) {
    return true;
}

std::vector<VMwareSnapshot> VMwareAgent::ListSnapshots(const std::string& vm_id) {
    return {};
}

bool VMwareAgent::RevertToSnapshot(const std::string& vm_id, const std::string& snapshot_id) {
    return true;
}

std::string VMwareAgent::GetDiskPath(const std::string& vm_id, int disk_number) {
    return "";
}

uint64_t VMwareAgent::GetChangedBlocks(const std::string& vm_id, const std::string& snapshot_id) {
    return 0;
}

std::unique_ptr<IVMwareAgent> CreateVMwareAgent() {
    return std::make_unique<VMwareAgent>();
}

}
