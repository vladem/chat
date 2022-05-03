#pragma once

#include "storage/poller/poller.h"
#include <memory>
#include <string>

namespace NStorage {
class IServer {
public:
    virtual void Run() = 0;

    virtual void Stop() = 0;
};

std::unique_ptr<IServer> CreateServer(std::unique_ptr<TPoller>, const std::string& address);
} // namespace NStorage
