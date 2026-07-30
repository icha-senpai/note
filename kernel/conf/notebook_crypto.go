// Scribli - Refactor your thinking
// Copyright (c) 2020-present, Scribli
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package conf

import "github.com/icha-senpai/note/kernel/util"

type NotebookCrypto struct {
	Enabled         bool              `json:"enabled"`
	MasterSalt      []byte            `json:"masterSalt"`
	KDFParams       util.Argon2Params `json:"kdfParams"`
	KEKVerifier     []byte            `json:"kekVerifier"`
	VerifierNonce   []byte            `json:"verifierNonce"`
	AutoLockMinutes int               `json:"autoLockMinutes"`

	Spec      int    `json:"spec,omitempty"`
	BackupID  string `json:"backupID,omitempty"`
	CreatedAt int64  `json:"createdAt,omitempty"`
	Checksum  string `json:"checksum,omitempty"`
	KEKMAC    []byte `json:"kekMAC,omitempty"`
}

func NewNotebookCrypto() *NotebookCrypto {
	return &NotebookCrypto{
		KDFParams:       util.DefaultArgon2Params(),
		AutoLockMinutes: 5,
		Spec:            CurrentNotebookCryptoSpec,
	}
}

const CurrentNotebookCryptoSpec = 1

func UpgradeSpec(nc *NotebookCrypto) (upgraded bool) {

	// if nc.Spec == 1 {

	// 	nc.Spec = 2
	// 	upgraded = true
	// }
	return
}
