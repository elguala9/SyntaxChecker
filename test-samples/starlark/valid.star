DEPS = [
    "//pkg/a",
    "//pkg/b",
]

def library(name, srcs, deps = []):
    return struct(
        name = name,
        srcs = srcs,
        deps = deps + DEPS,
    )

lib = library("core", ["main.star"], deps = ["//pkg/c"])
print(lib.name)
