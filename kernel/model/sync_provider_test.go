package model

import (
	"testing"

	appconf "github.com/siyuan-note/siyuan/kernel/conf"
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
