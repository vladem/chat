cc_library(
  name = "server",
  srcs = ["server.cpp"],
  hdrs = ["server.h"],
  visibility = ["//visibility:public"],
  deps = [
    "@com_github_grpc_grpc//:grpc++",
     "//:storage_cc_grpc",
     "//storage/poller:poller"
  ]
)