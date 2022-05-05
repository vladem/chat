#include "server.h"
#include <iostream>
#include <memory>
#include <string>
#include <thread>

#include <grpc/support/log.h>
#include <grpcpp/grpcpp.h>

#include "proto/storage.grpc.pb.h"

namespace NStorage {
using grpc::Server;
using grpc::ServerAsyncResponseWriter;
using grpc::ServerBuilder;
using grpc::ServerCompletionQueue;
using grpc::ServerContext;
using grpc::Status;
using storage::Greeter;
using storage::HelloReply;
using storage::HelloRequest;

class TServerImpl final : public IServer {
public:
    TServerImpl(std::unique_ptr<TPoller> poller, const std::string& address) : Poller(std::move(poller)), Address(address) {}

    ~TServerImpl() {
        Stop();
    }

    virtual void Run() override {
        Running = true;
        ServerBuilder builder;
        builder.AddListeningPort(Address, grpc::InsecureServerCredentials());
        builder.RegisterService(&Service_);
        Cq_ = builder.AddCompletionQueue();
        Server_ = builder.BuildAndStart();
        std::cout << "Server listening on " << Address << std::endl;
        HandleRpcs();
    }

    virtual void Stop() override {
        if (!Running) {
            return;
        }
        Running = false;
        Poller->Stop();  // TODO: weird to stop poller here, if it's started outside + TPoller may be destructed right after this method
        Server_->Shutdown();
        Cq_->Shutdown();
    }

private:
    class TCallData : public std::enable_shared_from_this<TCallData> {
    public:
        TCallData(Greeter::AsyncService* service, ServerCompletionQueue* cq, TPoller* poller)
            : Service_(service), Cq_(cq), Responder_(&Ctx_), Poller(poller) {
            Service_->RequestSayHello(&Ctx_, &Request_, &Responder_, Cq_, Cq_, this);
        }

        TPoller::TCoroutine Proceed() {
            const auto this_ = shared_from_this(); // holds 'this' reference until the end of method 
            std::cout << "Write started" << std::endl;
            co_await Poller->Write(std::move(*Request_.mutable_name())); // TODO: co_await returning result (ok/error)
            std::cout << "Write done" << std::endl;
            {
                // TODO: fix this
                // TCalldata may be destroyed only after tag passed to the 'Responder_.Finish' is processed by the Cq
                Responder_.Finish(Reply_, Status::OK, nullptr);
                std::this_thread::sleep_for(std::chrono::milliseconds(1000));
            }
            // Responder_.Finish(Reply_, Status::OK, this);
            // co_await Poller->WaitFinished();  // TODO: implement me
            co_return;
        }

        // TODO: remove me
        ~TCallData() {
            std::cout << "~TCallData" << std::endl;
        }

    private:
        Greeter::AsyncService* Service_;
        ServerCompletionQueue* Cq_;
        ServerContext Ctx_;
        HelloRequest Request_;
        HelloReply Reply_;
        ServerAsyncResponseWriter<HelloReply> Responder_;
        TPoller* Poller;
    };

    void HandleRpcs() {
        auto callData = std::make_shared<TCallData>(&Service_, Cq_.get(), Poller.get());
        void* tag;
        bool ok;
        while (true) {
            GPR_ASSERT(Cq_->Next(&tag, &ok));
            std::cout << "got event from cq" << std::endl;
            if (!ok) {
                break;
            }
            if (!tag) {
                continue;
            }
            static_cast<TCallData*>(tag)->Proceed();
            callData = std::make_shared<TCallData>(&Service_, Cq_.get(), Poller.get()); // TODO: limit pending calls?
        }
        std::cout << "Server main loop finished" << std::endl;
    }

    std::unique_ptr<ServerCompletionQueue> Cq_;
    Greeter::AsyncService Service_;
    std::unique_ptr<Server> Server_;
    std::unique_ptr<TPoller> Poller;
    std::string Address;
    bool Running = false;
};

std::unique_ptr<IServer> CreateServer(std::unique_ptr<TPoller> poller, const std::string& address) {
    return std::make_unique<TServerImpl>(std::move(poller), address);
}
} // namespace NStorage