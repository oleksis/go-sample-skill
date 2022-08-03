/*
 * Copyright © 2022 Atomist, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"reflect"
	"testing"

	"github.com/atomist-skills/go-skill"
	"olympos.io/encoding/edn"
)

func TestProcessCommit(t *testing.T) {
	commit := OnCommit{
		Sha: "d2c6724307f007755fc770944fd7bc5ff55933b0",
		Repo: struct {
			Name          string `edn:"git.repo/name"`
			DefaultBranch string `edn:"git.repo/default-branch"`
			Org           struct {
				Name              string `edn:"git.org/name"`
				InstallationToken string `edn:"github.org/installation-token"`
				Url               string `edn:"git.provider/url"`
			} `edn:"git.repo/org"`
			SourceId string `edn:"git.repo/source-id"`
		}{
			Name:          "go-sample-skill",
			DefaultBranch: "main",
			Org: struct {
				Name              string `edn:"git.org/name"`
				InstallationToken string `edn:"github.org/installation-token"`
				Url               string `edn:"git.provider/url"`
			}{
				Name: "atomist-skills",
				Url:  "https://github.com/",
			},
			SourceId: "123456",
		},
	}
	req := skill.RequestContext{
		Log: skill.Logger{
			Debug:  func(msg string) { t.Log(msg) },
			Debugf: t.Logf,
			Info:   func(msg string) { t.Log(msg) },
			Infof:  t.Logf,
		},
		Transact: func(entities interface{}) error {
			switch reflect.TypeOf(entities).Kind() {
			case reflect.Slice:
				s := reflect.ValueOf(entities)
				if s.Len() != 3 {
					t.Errorf("Expected 3 entities, got %d", s.Len())
				}

				repoEntity := s.Index(0).Interface().(GitRepoEntity)
				if !reflect.DeepEqual(repoEntity, GitRepoEntity{
					EntityType: edn.Keyword("git/repo"),
					Entity:     "$repo",
					SourceId:   "123456",
					Url:        "https://github.com/",
				}) {
					t.Errorf("Repo entity not expected")
				}
				commitEntity := s.Index(1).Interface().(GitCommitEntity)
				if !reflect.DeepEqual(commitEntity, GitCommitEntity{
					EntityType: edn.Keyword("git/commit"),
					Entity:     "$commit",
					Repo:       "$repo",
					Sha:        commit.Sha,
					Url:        "https://github.com/",
				}) {
					t.Errorf("Commit entity not expected")
				}
				commitSignatureEntity := s.Index(2).Interface().(GitCommitSignatureEntity)
				if !(commitSignatureEntity.Signature != "" && commitSignatureEntity.Reason == "valid") {
					t.Errorf("Signature entity not expected")
				}

				return nil
			}
			t.Error("Expected slice of entities")
			return nil
		},
	}
	ctx := context.Background()
	gitCommit, err := getCommit(ctx, req, &commit)
	if err != nil {
		t.Error("getCommit errored")
	}
	err = transactCommitSignature(context.Background(), req, commit, gitCommit)
	if err != nil {
		t.Error("transactCommitSignature errored")
	}
}
