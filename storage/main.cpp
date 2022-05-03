#include "poller/poller.h"
#include "server/server.h"

#include <thread>
#include <sstream>

int main(int argc, char** argv) {
    auto poller = NStorage::CreatePoller();
    std::thread pollerThread(&NStorage::TPoller::Run, poller.get());
    std::ostringstream s;
    s << "0.0.0.0:" << 50051;
    auto server = NStorage::CreateServer(std::move(poller), s.str());
    std::thread serverThread(&NStorage::IServer::Run, server.get());
    //server->Stop();
    //poller->Stop();
    serverThread.join();
    pollerThread.join();
    return 0;
}