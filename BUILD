package(default_visibility = ["//visibility:public"])

load("@rules_proto//proto:defs.bzl", "proto_library")
load("@rules_cc//cc:defs.bzl", "cc_binary", "cc_proto_library")
load("@com_github_grpc_grpc//bazel:cc_grpc_library.bzl", "cc_grpc_library")

proto_library(
    name = "storage_proto",
    srcs = ["proto/storage.proto"],
)

cc_proto_library(
    name = "storage_cc_proto",
    deps = [":storage_proto"],
)

cc_grpc_library(
    name = "storage_cc_grpc",
    srcs = [":storage_proto"],
    grpc_only = True,
    deps = [":storage_cc_proto"],
)