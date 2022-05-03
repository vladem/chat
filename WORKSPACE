load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
http_archive(
    name = "com_github_grpc_grpc",
    urls = [
        "https://github.com/grpc/grpc/archive/b39ffcc425ea990a537f98ec6fe6a1dcb90470d7.tar.gz",
    ],
    strip_prefix = "grpc-b39ffcc425ea990a537f98ec6fe6a1dcb90470d7",
)
load("@com_github_grpc_grpc//bazel:grpc_deps.bzl", "grpc_deps")
grpc_deps()
load("@com_github_grpc_grpc//bazel:grpc_extra_deps.bzl", "grpc_extra_deps")
grpc_extra_deps()


http_archive(
  name = "com_google_googletest",
  urls = ["https://github.com/google/googletest/archive/e2239ee6043f73722e7aa812a459f54a28552929.zip"],
  strip_prefix = "googletest-e2239ee6043f73722e7aa812a459f54a28552929",
)