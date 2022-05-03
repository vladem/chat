#include "poller.h"

// TODO: remove me
#include <chrono>
#include <thread>

namespace NStorage {
void TPoller::Run() {
    {
        const auto guard = std::lock_guard<std::mutex>(Lock);
        Running = true;
    }
    while (true) {
        std::cout << "TPoller sleeping.." << std::endl;
        std::this_thread::sleep_for(std::chrono::milliseconds(1000));
        const auto guard = std::lock_guard<std::mutex>(Lock);
        if (!Running) {
            break;
        }
        if (PendingCoros.empty()) {
            continue;
        }
        auto coro = PendingCoros.front();
        PendingCoros.pop();
        std::cout << "TPoller calling coro.." << std::endl;
        coro();
    }
}

void TPoller::Stop() {
    const auto guard = std::lock_guard<std::mutex>(Lock);
    Running = false;
}

// TAwaitable Read() {
//     return TAwaitable(*this);
// }

TPoller::TAwaitable TPoller::Write(std::string&&) {
    return TAwaitable(*this);
}

std::unique_ptr<TPoller> CreatePoller() {
    return std::make_unique<TPoller>();
}
} // namespace NStorage