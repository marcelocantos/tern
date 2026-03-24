// swift-tools-version: 5.9

import PackageDescription

let package = Package(
    name: "Tern",
    platforms: [.iOS(.v16), .macOS(.v13)],
    products: [
        .library(name: "Tern", targets: ["Tern"]),
    ],
    targets: [
        .target(name: "Tern"),
        .testTarget(name: "TernTests", dependencies: ["Tern"]),
        .executableTarget(
            name: "tern-e2e-swift",
            dependencies: ["Tern"],
            path: "e2e/swift"
        ),
    ]
)
