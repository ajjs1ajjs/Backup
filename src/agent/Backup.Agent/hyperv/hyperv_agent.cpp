#include "hyperv_agent.h"

namespace backup {

HyperVAgent::HyperVAgent() : connected_(false) {}

HyperVAgent::~HyperVAgent() = default;

bool HyperVAgent::Connect(const std::string& host, const std::string& username, const std::string& password) {
    connected_ = true;
    host_ = host;
    return true;
}

void HyperVAgent::Disconnect() {
    connected_ = false;
}

bool HyperVAgent::IsConnected() const {
    return connected_;
}

std::vector<VirtualMachine> HyperVAgent::ListVMs() {
    return cached_vms_;
}

VirtualMachine* HyperVAgent::GetVM(const std::string& vm_id) {
    for (auto& vm : cached_vms_) {
        if (vm.vm_id == vm_id) return &vm;
    }
    return nullptr;
}

bool HyperVAgent::CreateSnapshot(const std::string& vm_id, const std::string& snapshot_name) {
    return true;
}

bool HyperVAgent::DeleteSnapshot(const std::string& vm_id, const std::string& snapshot_id) {
    return true;
}

std::vector<Snapshot> HyperVAgent::ListSnapshots(const std::string& vm_id) {
    return {};
}

bool HyperVAgent::RevertToSnapshot(const std::string& vm_id, const std::string& snapshot_id) {
    return true;
}

std::string HyperVAgent::GetDiskPath(const std::string& vm_id, int32_t disk_number) {
    return "";
}

uint64_t HyperVAgent::GetChangedBlocks(const std::string& vm_id, const std::string& snapshot_id) {
    return 0;
}

std::unique_ptr<IHypervisorAgent> CreateHyperVAgent() {
    return std::make_unique<HyperVAgent>();
}

}
