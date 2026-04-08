#include "data_mover.h"
#include <algorithm>
#include <chrono>
#include <fstream>
#include <sstream>
#include <iostream>
#include <thread>
#include <future>

namespace backup {

class FileDataSource : public IDataSource {
public:
    explicit FileDataSource(const std::string& path) : path_(path) {}

    bool Open() override {
        file_.open(path_, std::ios::binary | std::ios::ate);
        if (!file_.is_open()) return false;
        size_ = file_.tellg();
        file_.seekg(0, std::ios::beg);
        return true;
    }

    void Close() override {
        if (file_.is_open()) file_.close();
    }

    bool ReadChunk(uint64_t offset, std::vector<uint8_t>& buffer) override {
        if (!file_.is_open()) return false;
        file_.seekg(offset, std::ios::beg);
        if (!file_.good()) return false;
        
        buffer.resize(chunk_size_);
        file_.read(reinterpret_cast<char*>(buffer.data()), chunk_size_);
        buffer.resize(file_.gcount());
        return buffer.size() > 0;
    }

    uint64_t GetSize() const override { return size_; }
    std::string GetChecksum() const override { return checksum_; }
    void SetChunkSize(uint64_t size) { 
        const uint64_t MAX_CHUNK_SIZE = 64 * 1024 * 1024; // 64 MB
        chunk_size_ = std::min(size, MAX_CHUNK_SIZE);
        if (chunk_size_ == 0) chunk_size_ = 64 * 1024; // Default to 64 KB if 0
    }

private:
    std::string path_;
    std::ifstream file_;
    uint64_t size_ = 0;
    uint64_t chunk_size_ = 64 * 1024;
    std::string checksum_;
};

class FileDataDestination : public IDataDestination {
public:
    explicit FileDataDestination(const std::string& path) : path_(path) {}

    bool Open() override {
        file_.open(path_, std::ios::binary | std::ios::trunc);
        return file_.is_open();
    }

    void Close() override {
        if (file_.is_open()) {
            file_.flush();
            file_.close();
        }
    }

    bool WriteChunk(uint64_t offset, const std::vector<uint8_t>& buffer) override {
        if (!file_.is_open()) return false;
        file_.seekp(offset, std::ios::beg);
        if (!file_.good()) return false;
        file_.write(reinterpret_cast<const char*>(buffer.data()), buffer.size());
        return file_.good();
    }

    bool Flush() override {
        if (file_.is_open()) {
            file_.flush();
            return file_.good();
        }
        return false;
    }

private:
    std::string path_;
    std::ofstream file_;
};

class TransferTask : public ITransferTask {
public:
    TransferTask(
        std::unique_ptr<IDataSource> source,
        std::unique_ptr<IDataDestination> destination,
        const TransferOptions& options)
        : source_(std::move(source))
        , destination_(std::move(destination))
        , options_(options)
        , status_(TransferStatus::Pending)
        , bytes_transferred_(0)
        , total_bytes_(0)
        , speed_mbps_(0.0)
    {}

    void Start() override {
        if (!source_->Open() || !destination_->Open()) {
            progress_.status = TransferStatus::Failed;
            progress_.error_message = "Failed to open source or destination";
            return;
        }

        total_bytes_ = source_->GetSize();
        progress_.total_bytes = total_bytes_;
        progress_.status = TransferStatus::InProgress;
        status_ = TransferStatus::InProgress;

        auto start_time = std::chrono::steady_clock::now();
        uint64_t offset = 0;
        uint64_t chunk_index = 0;
        std::vector<uint8_t> buffer;

        while (status_ == TransferStatus::InProgress && source_->ReadChunk(offset, buffer)) {
            if (!destination_->WriteChunk(offset, buffer)) {
                progress_.status = TransferStatus::Failed;
                progress_.error_message = "Failed to write chunk";
                break;
            }

            bytes_transferred_ += buffer.size();
            offset += buffer.size();
            chunk_index++;

            progress_.bytes_transferred = bytes_transferred_;
            UpdateSpeed(start_time);

            if (progress_callback_) {
                progress_callback_(progress_);
            }

            if (buffer.size() < options_.chunk_size) break;
        }

        if (status_ == TransferStatus::InProgress) {
            destination_->Flush();
            progress_.status = TransferStatus::Completed;
            status_ = TransferStatus::Completed;
        }
    }

    void Pause() override {
        if (status_ == TransferStatus::InProgress) {
            status_ = TransferStatus::Paused;
            progress_.status = TransferStatus::Paused;
        }
    }

    void Resume() override {
        if (status_ == TransferStatus::Paused) {
            status_ = TransferStatus::InProgress;
            progress_.status = TransferStatus::InProgress;
        }
    }

    void Cancel() override {
        status_ = TransferStatus::Cancelled;
        progress_.status = TransferStatus::Cancelled;
    }

    TransferProgress GetProgress() const override { return progress_; }

    bool WaitForCompletion(std::chrono::milliseconds timeout) override {
        std::this_thread::sleep_for(timeout);
        return status_ == TransferStatus::Completed || status_ == TransferStatus::Failed;
    }

    void SetProgressCallback(ProgressCallback callback) {
        progress_callback_ = std::move(callback);
    }

private:
    void UpdateSpeed(const std::chrono::steady_clock::time_point& start_time) {
        auto elapsed = std::chrono::duration_cast<std::chrono::milliseconds>(
            std::chrono::steady_clock::now() - start_time).count();
        
        if (elapsed > 0) {
            double seconds = elapsed / 1000.0;
            speed_mbps_ = (bytes_transferred_ / (1024.0 * 1024.0)) / seconds;
            progress_.speed_mbps = speed_mbps_;
        }
    }

    std::unique_ptr<IDataSource> source_;
    std::unique_ptr<IDataDestination> destination_;
    TransferOptions options_;
    
    std::atomic<TransferStatus> status_;
    std::atomic<uint64_t> bytes_transferred_;
    uint64_t total_bytes_;
    double speed_mbps_;
    
    TransferProgress progress_;
    ProgressCallback progress_callback_;
};

class DataMover {
public:
    DataMover() = default;
    ~DataMover() {
        WaitAll();
    }

    std::shared_ptr<ITransferTask> CreateTransfer(
        const std::string& source_path,
        const std::string& destination_path,
        const TransferOptions& options = {})
    {
        auto source = std::make_unique<FileDataSource>(source_path);
        auto destination = std::make_unique<FileDataDestination>(destination_path);
        
        auto task = std::make_shared<TransferTask>(
            std::move(source),
            std::move(destination),
            options
        );
        
        tasks_.push_back(task);
        return task;
    }

    void StartTransfer(std::shared_ptr<ITransferTask> task) {
        std::thread([task]() {
            task->Start();
        }).detach();
    }

    void WaitAll() {
        for (auto& task : tasks_) {
            task->WaitForCompletion(std::chrono::seconds(30));
        }
    }

    size_t GetActiveCount() const {
        return std::count_if(tasks_.begin(), tasks_.end(), [](const auto& task) {
            return task->GetProgress().status == TransferStatus::InProgress;
        });
    }

private:
    std::vector<std::shared_ptr<ITransferTask>> tasks_;
};

}
