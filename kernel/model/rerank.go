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

package model

const defaultRerankCandidateCount = 30

func isRerankEnabled() bool {
	return nil != Conf.AI.Rerank && Conf.AI.Rerank.Enabled && len(Conf.AI.Rerank.APIKey) > 0
}

func rerankKey() string {
	if nil != Conf.AI.Rerank && Conf.AI.Rerank.Enabled && "" != Conf.AI.Rerank.APIKey {
		return Conf.AI.Rerank.APIKey
	}
	return ""
}

func rerankEndpoint() string {
	if nil != Conf.AI.Rerank && Conf.AI.Rerank.Enabled && "" != Conf.AI.Rerank.Endpoint {
		return Conf.AI.Rerank.Endpoint
	}
	return ""
}

func rerankModel() string {
	if nil != Conf.AI.Rerank && Conf.AI.Rerank.Enabled && "" != Conf.AI.Rerank.Name {
		return Conf.AI.Rerank.Name
	}
	return ""
}

func rerankTimeout() int {
	if nil != Conf.AI.Rerank && Conf.AI.Rerank.Enabled && 0 < Conf.AI.Rerank.Timeout {
		return Conf.AI.Rerank.Timeout
	}
	return 30
}

func rerankCandidateCount() int {
	if nil != Conf.AI.Rerank && Conf.AI.Rerank.Enabled && 0 < Conf.AI.Rerank.CandidateCount {
		return Conf.AI.Rerank.CandidateCount
	}
	return defaultRerankCandidateCount
}
