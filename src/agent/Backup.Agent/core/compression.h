#pragma once

#include <string>
#include <vector>
#include <cstdint>
#include <memory>

namespace backup {

struct CompressionConfig {
    std::string algorithm = "zstd";
    int level = 3;
    uint32_t block_size = 64 * 1024;
};

class ICompressor {
public:
    virtual ~ICompressor() = default;
    virtual std::vector<uint8_t> Compress(const std::vector<uint8_t>& data) = 0;
    virtual std::vector<uint8_t> Decompress(const std::vector<uint8_t>& data) = 0;
    virtual uint64_t GetCompressedSize() const = 0;
};

class ZstdCompressor : public ICompressor {
public:
    explicit ZstdCompressor(int level = 3);
    ~ZstdCompressor() override;
    
    std::vector<uint8_t> Compress(const std::vector<uint8_t>& data) override;
    std::vector<uint8_t> Decompress(const std::vector<uint8_t>& data) override;
    uint64_t GetCompressedSize() const override;

private:
    struct Impl;
    std::unique_ptr<Impl> impl_;
};

class Lz4Compressor : public ICompressor {
public:
    std::vector<uint8_t> Compress(const std::vector<uint8_t>& data) override;
    std::vector<uint8_t> Decompress(const std::vector<uint8_t>& data) override;
    uint64_t GetCompressedSize() const override;
};

class GzipCompressor : public ICompressor {
public:
    explicit GzipCompressor(int level = 6);
    std::vector<uint8_t> Compress(const std::vector<uint8_t>& data) override;
    std::vector<uint8_t> Decompress(const std::vector<uint8_t>& data) override;
    uint64_t GetCompressedSize() const override;

private:
    int level_;
};

class NoCompression : public ICompressor {
public:
    std::vector<uint8_t> Compress(const std::vector<uint8_t>& data) override;
    std::vector<uint8_t> Decompress(const std::vector<uint8_t>& data) override;
    uint64_t GetCompressedSize() const override;
};

std::unique_ptr<ICompressor> CreateCompressor(const std::string& algorithm, int level = 3);

}
