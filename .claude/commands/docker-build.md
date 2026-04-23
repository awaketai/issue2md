---
allowed-tools: Bash(docker:*), Bash(make:docker-build)
---

运行 `make docker-build` 构建 Docker 镜像。

- 如果用户提供了参数 `$ARGUMENTS`，则将其作为镜像 tag，执行 `make docker-build IMAGE_TAG=issue2md:$ARGUMENTS`。
- 如果用户没有提供参数（`$ARGUMENTS` 为空），则使用默认 tag `latest`，执行 `make docker-build`。

构建完成后，输出镜像大小信息。如果构建失败，自动分析错误原因并给出修复建议。
