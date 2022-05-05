#include "poller.h"

// TODO: remove me
#include <chrono>
#include <thread>

namespace NStorage {
void TPoller::Run() {
    Running.store(true);
    while (true) {
        std::cout << "TPoller sleeping.." << std::endl;
        std::this_thread::sleep_for(std::chrono::milliseconds(1000));
        if (!Running.load()) {
            break;
        }
        std::coroutine_handle<> coro;
        {
            const auto guard = std::lock_guard<std::mutex>(Lock);
            if (PendingCoros.empty()) {
                continue;
            }
            coro = PendingCoros.front();
            PendingCoros.pop();
        }
        std::cout << "TPoller calling coro.." << std::endl;
        coro();
    }
    std::cout << "Poller main loop finished" << std::endl;
}

void TPoller::Stop() {
    std::cout << "Stop poller" << std::endl;
    Running.store(false);
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