#include "client.h"
#include <memory>

using grpc::ClientContext;
using grpc::Status;
using storage::Greeter;
using storage::HelloReply;
using storage::HelloRequest;

namespace NStorage {
class TClient : public IClient {
public:
    TClient(std::shared_ptr<grpc::Channel> channel)
        : Stub_(Greeter::NewStub(channel)) {}

    virtual std::string Write(const std::string& data) override {
        HelloRequest request;
        request.set_name(data);
        HelloReply reply;
        ClientContext context;
        Status status = Stub_->SayHello(&context, request, &reply);
        if (status.ok()) {
            return reply.message();
        } else {
            std::cout << status.error_code() << ": " << status.error_message() << std::endl;
            return "RPC failed";
        }
    }

private:
    std::unique_ptr<Greeter::Stub> Stub_;
};

std::unique_ptr<IClient> CreateClient(std::shared_ptr<grpc::Channel> channel) {
    return std::make_unique<TClient>(channel);
}
} // namespace NStorage