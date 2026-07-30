# DejaVu

DejaVu is Scribli's local fork of the data snapshot and sync component used by the kernel.

## Features

- Git-like version control.
- File deduplication in chunks.
- Data compression.
- AES encryption.
- User-controlled sync and backup backends.

## Notes

- Folders are not supported.
- Permission attributes are not supported.
- Symbolic links are not supported.
- Official cloud behavior must stay disabled in Scribli; keep sync providers user-controlled.

## License

DejaVu uses the GNU Affero General Public License, Version 3.

## Acknowledgement

- [dustin/go-humanize](https://github.com/icha-senpai/note/third_party/forks/github/dustin/go-humanize) `MIT license`
- [klauspost/compress](https://github.com/icha-senpai/note/third_party/forks/github/klauspost/compress) `BSD-3-Clause license`
- [panjf2000/ants](https://github.com/panjf2000/ants) `MIT license`
- [InfuseAI/ArtiVC](https://github.com/InfuseAI/ArtiVC) `Apache-2.0 license`
- [restic/restic](https://github.com/restic/restic) `BSD-2-Clause license`
- [sabhiram/go-gitignore](https://github.com/icha-senpai/note/third_party/forks/github/sabhiram/go-gitignore) `MIT license`
