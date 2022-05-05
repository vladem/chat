#pragma once
#include <coroutine>
#include <iostream>
#include <memory>
#include <mutex>
#include <atomic>
#include <queue>
#include <string>

namespace NStorage {
class TPoller {
public:
    struct TCoroutine {
        struct TPromise {
            TCoroutine get_return_object() { return {}; }
            std::suspend_never initial_suspend() { return {}; }
            std::suspend_never final_suspend() noexcept { return {}; }
            void return_void() {}
            void unhandled_exception() {}
        };
        using promise_type = TPromise;
    };

    struct TAwaitable {
        TAwaitable(TPoller& parent) : Parent(parent) {}
        bool await_ready() { return false; }
        void await_suspend(std::coroutine_handle<> h) {
            const auto guard = std::lock_guard<std::mutex>(Parent.Lock);
            Parent.PendingCoros.push(h);
        }
        void await_resume() {}

        TPoller& Parent;
    };

public:
    void Run();
    void Stop();
    TAwaitable Read();
    TAwaitable Write(std::string&&);
    ~TPoller() {
        std::cout << "~TPoller" << std::endl;
        while (!PendingCoros.empty()) {
            PendingCoros.front().destroy();
            PendingCoros.pop();
        }
    }

private:
    std::queue<std::coroutine_handle<>> PendingCoros;
    std::mutex Lock;
    std::atomic<bool> Running = false;
};

std::unique_ptr<TPoller> CreatePoller();
} // namespace NStorage