#include "storage/poller/poller.h"
#include "storage/server/server.h"
#include "storage/tests/client.h"
#include <gtest/gtest.h>
#include <thread>

TEST(TestWrite, TestSimple) {
    // server
    auto poller = NStorage::CreatePoller();
    std::thread pollerThread(&NStorage::TPoller::Run, poller.get());
    size_t port = 50052; // TODO: pick unused
    std::ostringstream s;
    s << "0.0.0.0:" << port;
    auto server = NStorage::CreateServer(std::move(poller), s.str());
    std::thread serverThread(&NStorage::IServer::Run, server.get());
    // std::this_thread::sleep_for(std::chrono::milliseconds(1000));

    // client
    auto client = NStorage::CreateClient(
        grpc::CreateChannel(s.str(), grpc::InsecureChannelCredentials()));
    std::string data("some data");
    std::string reply = client->Write(data);
    std::cout << "client received: " << reply << std::endl;
    // EXPECT_STRNE("hello", "world");
    // EXPECT_EQ(7 * 6, 42);
    server->Stop();
    serverThread.join();
    pollerThread.join();
}
