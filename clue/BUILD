load("@rules_cc//cc:defs.bzl", "go_binary", "go_library", "go_test")

cc_library(
    name = "hello-greet",
    srcs = ["hello-greet.cc"],
)

cc_binary(
    name = "hello-world",
    srcs = ["hello-world.cc"],
    deps = [
        ":hello-greet",
        "//lib:hello-time",
    ],
)