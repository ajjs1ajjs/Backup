#include "vss_agent.h"
#include <iostream>

VssAgent::VssAgent() { CoInitializeEx(NULL, COINIT_MULTITHREADED); }
VssAgent::~VssAgent() { Cleanup(); CoUninitialize(); }

bool VssAgent::Initialize() {
    if (FAILED(CreateVssBackupComponents(&backupComponents))) return false;
    if (FAILED(backupComponents->InitializeForBackup())) return false;
    if (FAILED(backupComponents->StartSnapshotSet(&snapshotSetId))) return false;
    return true;
}

bool VssAgent::CreateSnapshot(const std::wstring& volumeName) {
    VSS_ID snapshotId;
    if (FAILED(backupComponents->AddToSnapshotSet(const_cast<PWSTR>(volumeName.c_str()), GUID_NULL, &snapshotId))) return false;
    
    if (FAILED(backupComponents->PrepareForBackup(&async))) return false;
    async->Wait();

    if (FAILED(backupComponents->DoSnapshotSet(&async))) return false;
    async->Wait();

    // Отримання шляху до снапшоту (наприклад, \\?\GLOBALROOT\Device\HarddiskVolumeShadowCopy1)
    VSS_SNAPSHOT_PROP prop;
    if (SUCCEEDED(backupComponents->GetSnapshotProperties(snapshotId, &prop))) {
        std::wcout << L"Snapshot created: " << prop.m_pwszSnapshotDeviceObject << std::endl;
        VssFreeSnapshotProperties(&prop);
        return true;
    }
    return false;
}

void VssAgent::Cleanup() {
    if (backupComponents) {
        backupComponents->DeleteSnapshots(snapshotSetId, VSS_OBJECT_SNAPSHOT_SET, TRUE, NULL, NULL);
        backupComponents->Release();
        backupComponents = nullptr;
    }
}
