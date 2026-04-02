#include "kvm_agent.h"

namespace backup {

KVMAgent::KVMAgent() = default;
KVMAgent::~KVMAgent() = default;

bool KVMAgent::Connect(const std::string& host, const std::string& username, const std::string& password) {
    connected_ = true;
    host_ = host;
    return true;
}

void KVMAgent::Disconnect() {
    connected_ = false;
}

bool KVMAgent::IsConnected() const {
    return connected_;
}

std::vector<KVMVM> KVMAgent::ListVMs() {
    return cached_vms_;
}

KVMVM* KVMAgent::GetVM(const std::string& vm_id) {
    for (auto& vm : cached_vms_) {
        if (vm.vm_id == vm_id) return &vm;
    }
    return nullptr;
}

bool KVMAgent::CreateSnapshot(const std::string& vm_id, const std::string& name) {
    return true;
}

bool KVMAgent::DeleteSnapshot(const std::string& vm_id, const std::string& snapshot_id) {
    return true;
}

std::vector<KVMSnapshot> KVMAgent::ListSnapshots(const std::string& vm_id) {
    return {};
}

bool KVMAgent::RevertToSnapshot(const std::string& vm_id, const std::string& snapshot_id) {
    return true;
}

std::string KVMAgent::GetDiskPath(const std::string& vm_id, int disk_number) {
    return "";
}

uint64_t KVMAgent::GetChangedBlocks(const std::string& vm_id, const std::string& snapshot_id) {
    return 0;
}

std::unique_ptr<IKVMAgent> CreateKVMAgent() {
    return std::make_unique<KVMAgent>();
}

}
