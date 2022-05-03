#pragma once
#include "proto/storage.grpc.pb.h"
#include <grpcpp/grpcpp.h>
#include <memory>
#include <string>

namespace NStorage {
class IClient {
public:
    virtual std::string Write(const std::string& data) = 0;
};

std::unique_ptr<IClient> CreateClient(std::shared_ptr<grpc::Channel> channel);
} // namespace NStorage
