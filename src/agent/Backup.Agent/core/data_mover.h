#pragma once

#include <string>
#include <vector>
#include <memory>
#include <functional>
#include <atomic>
#include <thread>
#include <mutex>
#include <queue>
#include <cstdint>
#include <optional>
#include <future>

namespace backup {

enum class TransferStatus {
    Pending,
    InProgress,
    Completed,
    Failed,
    Cancelled,
    Paused
};

enum class TransferDirection {
    Upload,
    Download
};

struct TransferOptions {
    uint64_t chunk_size = 64 * 1024; // 64KB default
    uint32_t max_concurrent_transfers = 4;
    uint32_t max_retry_attempts = 3;
    uint32_t retry_delay_ms = 1000;
    bool compression_enabled = true;
    bool deduplication_enabled = false;
    std::string checksum_algorithm = "xxhash";
};

struct TransferProgress {
    std::string transfer_id;
    uint64_t bytes_transferred = 0;
    uint64_t total_bytes = 0;
    double speed_mbps = 0.0;
    TransferStatus status = TransferStatus::Pending;
    std::string error_message;
};

struct Chunk {
    std::string transfer_id;
    uint64_t chunk_index = 0;
    std::vector<uint8_t> data;
    std::string checksum;
    bool is_last = false;
};

using ProgressCallback = std::function<void(const TransferProgress&)>;
using ChunkCallback = std::function<void(const Chunk&)>;

class IDataSource {
public:
    virtual ~IDataSource() = default;
    virtual bool Open() = 0;
    virtual void Close() = 0;
    virtual bool ReadChunk(uint64_t offset, std::vector<uint8_t>& buffer) = 0;
    virtual uint64_t GetSize() const = 0;
    virtual std::string GetChecksum() const = 0;
};

class IDataDestination {
public:
    virtual ~IDataDestination() = default;
    virtual bool Open() = 0;
    virtual void Close() = 0;
    virtual bool WriteChunk(uint64_t offset, const std::vector<uint8_t>& buffer) = 0;
    virtual bool Flush() = 0;
};

class ITransferTask {
public:
    virtual ~ITransferTask() = default;
    virtual void Start() = 0;
    virtual void Pause() = 0;
    virtual void Resume() = 0;
    virtual void Cancel() = 0;
    virtual TransferProgress GetProgress() const = 0;
    virtual bool WaitForCompletion(std::chrono::milliseconds timeout) = 0;
};

}
