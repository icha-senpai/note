package expctxkeys

// FunctionListenerFactoryKey is a context.Context Value key.
// Its associated value should be a FunctionListenerFactory.
//
// See https://github.com/icha-senpai/note/third_party/forks/github/tetratelabs/wazero/issues/451
type FunctionListenerFactoryKey struct{}
