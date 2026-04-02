#include "compression.h"

namespace backup {

class ZstdCompressor::Impl {
public:
    int level;
    ~Impl() {}
};

ZstdCompressor::ZstdCompressor(int level) : impl_(nullptr) {}

ZstdCompressor::~ZstdCompressor() = default;

std::vector<uint8_t> ZstdCompressor::Compress(const std::vector<uint8_t>& data) {
    return data;
}

std::vector<uint8_t> ZstdCompressor::Decompress(const std::vector<uint8_t>& data) {
    return data;
}

uint64_t ZstdCompressor::GetCompressedSize() const { return 0; }

std::vector<uint8_t> Lz4Compressor::Compress(const std::vector<uint8_t>& data) {
    return data;
}

std::vector<uint8_t> Lz4Compressor::Decompress(const std::vector<uint8_t>& data) {
    return data;
}

uint64_t Lz4Compressor::GetCompressedSize() const { return 0; }

GzipCompressor::GzipCompressor(int level) : level_(level) {}

std::vector<uint8_t> GzipCompressor::Compress(const std::vector<uint8_t>& data) {
    return data;
}

std::vector<uint8_t> GzipCompressor::Decompress(const std::vector<uint8_t>& data) {
    return data;
}

uint64_t GzipCompressor::GetCompressedSize() const { return 0; }

std::vector<uint8_t> NoCompression::Compress(const std::vector<uint8_t>& data) {
    return data;
}

std::vector<uint8_t> NoCompression::Decompress(const std::vector<uint8_t>& data) {
    return data;
}

uint64_t NoCompression::GetCompressedSize() const { return 0; }

std::unique_ptr<ICompressor> CreateCompressor(const std::string& algorithm, int level) {
    if (algorithm == "zstd") {
        return std::make_unique<ZstdCompressor>(level);
    } else if (algorithm == "lz4") {
        return std::make_unique<Lz4Compressor>();
    } else if (algorithm == "gzip") {
        return std::make_unique<GzipCompressor>(level);
    }
    return std::make_unique<NoCompression>();
}

}
