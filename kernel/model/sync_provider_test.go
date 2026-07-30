package model

import (
	"testing"

	appconf "github.com/icha-senpai/note/kernel/conf"
)

func TestDisabledSyncProviderIsNeverOnline(t *testing.T) {
	originalConf := Conf
	t.Cleanup(func() { Conf = originalConf })

	Conf = NewAppConf()
	Conf.Sync = &appconf.Sync{Provider: appconf.ProviderDisabled}

	if isProviderOnline(true) {
		t.Fatal("disabled sync provider reported online")
	}
}
