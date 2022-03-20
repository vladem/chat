# CLI messaging app 

## Overview
Simple client-server application to practice concurrency in go, and to try building grpc/protobuf-based web app from scratch. 

## Local usage
```bash
make run_server  # terminal 1
make run_client me=bob them=alice  # terminal 2
make run_client me=alice them=bob  # terminal 3
# type and send (\n-delimeted) messages in terminals 2,3 :)
```

## Testing
```bash
make test
```

## Nice to have in feature
- Nice logging (with timestamps and request tracing)
- Nice CLI interface (don't duplicate sent messages, first of all)
- Perfomance test (how much concurrent chats is it possible to have on one machine, at what RPS responses are noticeably slow)
- Dockerize
- External storage
- "Message seen" markers
