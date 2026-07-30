# Contributing

<!-- markdownlint-disable MD013 -->

## Get the source code

* Clone the Scribli repository.
* Work from the active Scribli development branch unless a maintainer asks for another branch.

## NPM dependencies

Install pnpm: `npm install -g pnpm@11.12.0`

Enter the app folder and execute:

* `pnpm install electron@42.6.1 -D`
* `pnpm run install:electron`
* `pnpm run dev`
* `pnpm run start`

Note: Electron 42 no longer downloads its binary automatically during `pnpm install`. Run `pnpm run install:electron` to fetch the binary before `pnpm run start`.

Note: `pnpm run start` launches the Electron shell, which starts the packaged Scribli kernel from `app/kernel`. If the kernel is missing, run the Windows build script first.

## Kernel

1. Install the latest version of [golang](https://go.dev/)
2. Open CGO support, that is, configure the environment variable `CGO_ENABLED=1`
3. On Windows, add the directory reported by `go env GOBIN` to `PATH`; if it is empty, add the `bin` subdirectory of `go env GOPATH`

### Windows Desktop

* `cd kernel`
* `go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest`
* `goversioninfo -platform-specific=true -icon=resource/icon.ico -manifest=resource/goversioninfo.exe.manifest`
* `go build -tags "fts5 sqlcipher" -o "../app/kernel/Scribli-Kernel.exe"`
* `cd ../app/kernel`
* `./Scribli-Kernel.exe serve --mode=dev`

Native mobile, macOS, and Linux build/package targets have been removed from this repository.

## Issue workflow

* Issues and pull requests that have been closed with no activity for 30 days are locked automatically to keep the tracker focused on open work
* If you run into a problem similar to a locked one, please open a new issue and link back to the original — avoid replying on old, closed threads, as that revives stale context and pings everyone who participated
* A new issue with a clear reproduction and a reference to the closed one is far easier to act on than a comment appended to a months-old thread
