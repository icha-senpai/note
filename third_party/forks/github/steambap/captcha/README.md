> Package captcha provides an easy to use, unopinionated API for captcha generation.

<div>

[![PkgGoDev](https://pkg.go.dev/badge/github.com/icha-senpai/note/third_party/forks/github/steambap/captcha)](https://pkg.go.dev/github.com/icha-senpai/note/third_party/forks/github/steambap/captcha)
[![Build Status](https://github.com/icha-senpai/note/third_party/forks/github/steambap/captcha/workflows/CI/badge.svg)](https://github.com/icha-senpai/note/third_party/forks/github/steambap/captcha/actions?workflow=CI)
[![codecov](https://codecov.io/gh/steambap/captcha/branch/main/graph/badge.svg)](https://codecov.io/gh/steambap/captcha)
[![Go Report Card](https://goreportcard.com/badge/github.com/icha-senpai/note/third_party/forks/github/steambap/captcha)](https://goreportcard.com/report/github.com/icha-senpai/note/third_party/forks/github/steambap/captcha)

</div>

## Why another captcha generator?
I want a simple and framework-independent way to generate captcha. It also should be flexible, at least allow me to pick my favorite font.

## install
```
go get github.com/icha-senpai/note/third_party/forks/github/steambap/captcha
```

## usage
```Go
func handle(w http.ResponseWriter, r *http.Request) {
	// create a captcha of 150x50px
	data, _ := captcha.New(150, 50)

	// session come from other library such as gorilla/sessions
	session.Values["captcha"] = data.Text
	session.Save(r, w)
	// send image data to client
	data.WriteImage(w)
}

```

[documentation](https://pkg.go.dev/github.com/icha-senpai/note/third_party/forks/github/steambap/captcha) |
[example](example/basic/main.go)

## sample image
![image](example/captcha.png)

![image](example/captcha-math.png)

## Compatibility

This package uses embed package from Go 1.16. If for some reasons you have to use pre 1.16 version of Go, reference pre 1.4 version of this module in your go.mod.

## Contributing
If your found a bug, please contribute!
see [contributing.md](contributing.md) for more detail.

## License
[MIT](LICENSE)
