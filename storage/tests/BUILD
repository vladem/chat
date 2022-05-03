cc_library(
  name = "client",
  srcs = ["client.cpp"],
  hdrs = ["client.h"],
  visibility = ["//visibility:public"],
  deps = [
    "@com_github_grpc_grpc//:grpc++",
     "//:storage_cc_grpc"
  ]
)

cc_test(
  name = "basic_test",
  size = "small",
  srcs = ["basic_test.cpp"],
  deps = ["@com_google_googletest//:gtest_main", "//storage/server:server", "//storage/tests:client"],
)