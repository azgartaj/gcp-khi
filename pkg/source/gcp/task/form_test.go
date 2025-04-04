// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package task

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/GoogleCloudPlatform/khi/pkg/common"
	form_task_test "github.com/GoogleCloudPlatform/khi/pkg/inspection/form/test"
	inspection_task_interface "github.com/GoogleCloudPlatform/khi/pkg/inspection/interface"
	"github.com/GoogleCloudPlatform/khi/pkg/inspection/metadata/form"
	inspection_task "github.com/GoogleCloudPlatform/khi/pkg/inspection/task"
	inspection_task_test "github.com/GoogleCloudPlatform/khi/pkg/inspection/test"
	"github.com/GoogleCloudPlatform/khi/pkg/parameters"
	"github.com/GoogleCloudPlatform/khi/pkg/source/gcp/query/queryutil"
	"github.com/GoogleCloudPlatform/khi/pkg/task"
	task_test "github.com/GoogleCloudPlatform/khi/pkg/task/test"
	"github.com/google/go-cmp/cmp/cmpopts"

	_ "github.com/GoogleCloudPlatform/khi/internal/testflags"
)

func TestProjectIdInput(t *testing.T) {
	testClusterNamePrefix := task_test.StubTaskFromReferenceID(ClusterNamePrefixTaskID, "", nil)
	form_task_test.TestTextForms(t, "gcp-project-id", InputProjectIdTask, []*form_task_test.TextFormTestCase{
		{
			Name:          "With valid project ID",
			Input:         "foo-project",
			ExpectedValue: "foo-project",
			Dependencies: []task.UntypedTask{
				testClusterNamePrefix,
			},
			ExpectedFormField: form.FormField{
				Priority:    1,
				Id:          GCPPrefix + "input/project-id",
				Type:        "Text",
				Label:       "Project ID",
				Description: "The project ID containing logs of the cluster to query",
				HintType:    form.HintTypeInfo,
				AllowEdit:   true,
			},
		},
		{
			Name:          "With fixed project ID from environment variable",
			Input:         "foo-project",
			ExpectedValue: "bar-project",
			Dependencies: []task.UntypedTask{
				testClusterNamePrefix,
			},
			ExpectedFormField: form.FormField{
				Priority:    1,
				Id:          GCPPrefix + "input/project-id",
				Type:        "Text",
				Label:       "Project ID",
				Description: "The project ID containing logs of the cluster to query",
				AllowEdit:   false,
				HintType:    form.HintTypeInfo,
				Default:     "bar-project",
			},
			Before: func() {
				expectedFixedProjectId := "bar-project"
				parameters.Auth.FixedProjectID = &expectedFixedProjectId
			},
			After: func() {
				parameters.Auth.FixedProjectID = nil
			},
		},
		{
			Name:          "With invalid project ID",
			Input:         "A invalid project ID",
			ExpectedValue: "",
			Dependencies: []task.UntypedTask{
				testClusterNamePrefix,
			},
			ExpectedFormField: form.FormField{
				Priority:        1,
				Id:              GCPPrefix + "input/project-id",
				Type:            "Text",
				Label:           "Project ID",
				Description:     "The project ID containing logs of the cluster to query",
				AllowEdit:       true,
				HintType:        form.HintTypeInfo,
				ValidationError: "Project ID must match `^*[0-9a-z\\.:\\-]+$`",
			},
		},
		{
			Name:          "Spaces around project ID must be trimmed",
			Input:         "  project-foo   ",
			ExpectedValue: "project-foo",
			Dependencies: []task.UntypedTask{
				testClusterNamePrefix,
			},
			ExpectedFormField: form.FormField{
				Priority:    1,
				Id:          GCPPrefix + "input/project-id",
				Type:        "Text",
				Label:       "Project ID",
				Description: "The project ID containing logs of the cluster to query",
				HintType:    form.HintTypeInfo,
				AllowEdit:   true,
			},
		},
		{
			Name:          "With valid old style project ID",
			Input:         "  deprecated.com:but-still-usable-project-id   ",
			ExpectedValue: "deprecated.com:but-still-usable-project-id",
			Dependencies: []task.UntypedTask{
				testClusterNamePrefix,
			},
			ExpectedFormField: form.FormField{
				Priority:    1,
				Id:          GCPPrefix + "input/project-id",
				Type:        "Text",
				Label:       "Project ID",
				Description: "The project ID containing logs of the cluster to query",
				HintType:    form.HintTypeInfo,
				AllowEdit:   true,
			},
		},
	})
}

func TestClusterNameInput(t *testing.T) {
	testClusterNamePrefix := task_test.StubTaskFromReferenceID(ClusterNamePrefixTaskID, "", nil)
	mockClusterNamesTask1 := task_test.StubTaskFromReferenceID(AutocompleteClusterNamesTaskID, &AutocompleteClusterNameList{
		ClusterNames: []string{"foo-cluster", "bar-cluster"},
		Error:        "",
	}, nil)
	form_task_test.TestTextForms(t, "cluster name", InputClusterNameTask, []*form_task_test.TextFormTestCase{
		{
			Name:          "with valid cluster name",
			Input:         "foo-cluster",
			ExpectedValue: "foo-cluster",
			Dependencies:  []task.UntypedTask{mockClusterNamesTask1, testClusterNamePrefix},
			ExpectedFormField: form.FormField{
				Priority:    1,
				Id:          GCPPrefix + "input/cluster-name",
				Type:        "Text",
				Label:       "Cluster name",
				AllowEdit:   true,
				Suggestions: []string{"foo-cluster", "bar-cluster"},
				Default:     "foo-cluster",
				HintType:    form.HintTypeInfo,
			},
		},
		{
			Name:          "spaces around cluster name must be trimmed",
			Input:         "  foo-cluster   ",
			ExpectedValue: "foo-cluster",
			Dependencies:  []task.UntypedTask{mockClusterNamesTask1, testClusterNamePrefix},
			ExpectedFormField: form.FormField{
				Priority:    1,
				Id:          GCPPrefix + "input/cluster-name",
				Type:        "Text",
				Label:       "Cluster name",
				AllowEdit:   true,
				Suggestions: []string{"foo-cluster", "bar-cluster"},
				Default:     "foo-cluster",
				HintType:    form.HintTypeInfo,
			},
		},
		{
			Name:          "invalid cluster name",
			Input:         "An invalid cluster name",
			ExpectedValue: "foo-cluster",
			Dependencies:  []task.UntypedTask{mockClusterNamesTask1, testClusterNamePrefix},
			ExpectedFormField: form.FormField{
				Priority:        1,
				Id:              GCPPrefix + "input/cluster-name",
				Type:            "Text",
				Label:           "Cluster name",
				ValidationError: "Cluster name must match `^[0-9a-z:\\-]+$`",
				AllowEdit:       true,
				Suggestions:     common.SortForAutocomplete("An invalid cluster name", []string{"foo-cluster", "bar-cluster"}),
				Default:         "foo-cluster",
				HintType:        form.HintTypeInfo,
			},
		},
		{
			Name:          "non existing cluster should show a hint",
			Input:         "nonexisting-cluster",
			ExpectedValue: "nonexisting-cluster",
			Dependencies:  []task.UntypedTask{mockClusterNamesTask1, testClusterNamePrefix},
			ExpectedFormField: form.FormField{
				Priority:        1,
				Id:              GCPPrefix + "input/cluster-name",
				Type:            "Text",
				Label:           "Cluster name",
				ValidationError: "",
				AllowEdit:       true,
				Suggestions:     []string{"foo-cluster", "bar-cluster"},
				Default:         "foo-cluster",
				Hint:            "Cluster `nonexisting-cluster` was not found in the specified project at this time. It works for the clusters existed in the past but make sure the cluster name is right if you believe the cluster should be there.",
				HintType:        form.HintTypeWarning,
			},
		},
	})
}

func TestDurationInput(t *testing.T) {
	expectedDescription := ""
	expectedLabel := "Duration"
	expectedSuggestions := []string{"1m", "10m", "1h", "3h", "12h", "24h"}
	timezoneTaskUTC := task_test.StubTask(TimeZoneShiftInputTask, time.UTC, nil)
	timezoneTaskJST := task_test.StubTask(TimeZoneShiftInputTask, time.FixedZone("", 9*3600), nil)
	currentTimeTask1 := task_test.StubTask(inspection_task.InspectionTimeProducer, time.Date(2023, time.April, 5, 12, 0, 0, 0, time.UTC), nil)
	endTimeTask := task_test.StubTask(InputEndTimeTask, time.Date(2023, time.April, 1, 12, 0, 0, 0, time.UTC), nil)

	form_task_test.TestTextForms(t, "duration", InputDurationTask, []*form_task_test.TextFormTestCase{
		{
			Name:          "With valid time duration",
			Input:         "10m",
			ExpectedValue: time.Duration(time.Minute) * 10,
			Dependencies:  []task.UntypedTask{endTimeTask, currentTimeTask1, timezoneTaskUTC},
			ExpectedFormField: form.FormField{
				Label:       expectedLabel,
				Description: expectedDescription,
				Suggestions: expectedSuggestions,
				AllowEdit:   true,
				Default:     "1h",
				HintType:    form.HintTypeInfo,
				Hint: `Query range:
2023-04-01T11:50:00Z ~ 2023-04-01T12:00:00Z
(UTC: 2023-04-01T11:50:00 ~ 2023-04-01T12:00:00)
(PDT: 2023-04-01T04:50:00 ~ 2023-04-01T05:00:00)`,
			},
		},
		{
			Name:          "With invalid time duration",
			Input:         "foo",
			ExpectedValue: time.Hour,
			Dependencies:  []task.UntypedTask{endTimeTask, currentTimeTask1, timezoneTaskUTC},
			ExpectedFormField: form.FormField{
				Label:           expectedLabel,
				Description:     expectedDescription,
				AllowEdit:       true,
				Suggestions:     expectedSuggestions,
				ValidationError: "time: invalid duration \"foo\"",
				Default:         "1h",
				HintType:        form.HintTypeInfo,
			},
		},
		{
			Name:          "With invalid time duration(negative)",
			Input:         "-10m",
			ExpectedValue: time.Hour,
			Dependencies:  []task.UntypedTask{endTimeTask, currentTimeTask1, timezoneTaskUTC},
			ExpectedFormField: form.FormField{
				Label:           expectedLabel,
				Description:     expectedDescription,
				AllowEdit:       true,
				Suggestions:     expectedSuggestions,
				ValidationError: "duration must be positive",
				Default:         "1h",
				HintType:        form.HintTypeInfo,
			},
		},
		{
			Name:          "with longer duration starting before than 30 days",
			Input:         "672h", // starting time will be 30 days before the inspection time
			ExpectedValue: time.Hour * 672,
			Dependencies:  []task.UntypedTask{endTimeTask, currentTimeTask1, timezoneTaskUTC},
			ExpectedFormField: form.FormField{
				Type:        "Text",
				Label:       expectedLabel,
				Description: expectedDescription,
				AllowEdit:   true,
				Suggestions: expectedSuggestions,
				Default:     "1h",
				Hint: `Specified time range starts from over than 30 days ago, maybe some logs are missing and the generated result could be incomplete.
This duration can be too long for big clusters and lead OOM. Please retry with shorter duration when your machine crashed.
Query range:
2023-03-04T12:00:00Z ~ 2023-04-01T12:00:00Z
(UTC: 2023-03-04T12:00:00 ~ 2023-04-01T12:00:00)
(PDT: 2023-03-04T05:00:00 ~ 2023-04-01T05:00:00)`,
				HintType: form.HintTypeInfo,
			},
		},
		{
			Name:          "With non UTC timezone",
			Input:         "1h",
			ExpectedValue: time.Hour,
			Dependencies:  []task.UntypedTask{endTimeTask, currentTimeTask1, timezoneTaskJST},
			ExpectedFormField: form.FormField{
				Type:        "Text",
				Label:       expectedLabel,
				Description: expectedDescription,
				AllowEdit:   true,
				Suggestions: expectedSuggestions,
				Default:     "1h",
				Hint: `Query range:
2023-04-01T20:00:00+09:00 ~ 2023-04-01T21:00:00+09:00
(UTC: 2023-04-01T11:00:00 ~ 2023-04-01T12:00:00)
(PDT: 2023-04-01T04:00:00 ~ 2023-04-01T05:00:00)`,
				HintType: form.HintTypeInfo,
			},
		},
	})
}

func TestInputEndtime(t *testing.T) {
	expectedDescription := "The endtime of query. Please input it in the format of RFC3339\n(example: 2006-01-02T15:04:05-07:00)"
	expectedLabel := "End time"
	expectedValue1, err := time.Parse(time.RFC3339, "2020-01-02T03:04:05Z")
	if err != nil {
		t.Errorf("unexpected error\n%s", err)
	}
	expectedValue2, err := time.Parse(time.RFC3339, "2020-01-02T00:00:00Z")
	timezoneTaskUTC := task_test.StubTask(TimeZoneShiftInputTask, time.UTC, nil)
	timezoneTaskJST := task_test.StubTask(TimeZoneShiftInputTask, time.FixedZone("", 9*3600), nil)

	if err != nil {
		t.Errorf("unexpected error\n%s", err)
	}
	form_task_test.TestTextForms(t, "endtime", InputEndTimeTask, []*form_task_test.TextFormTestCase{
		{
			Name:          "with empty",
			Input:         "",
			ExpectedValue: expectedValue1,
			Dependencies:  []task.UntypedTask{inspection_task.TestInspectionTimeTaskProducer("2020-01-02T03:04:05Z"), timezoneTaskUTC},
			ExpectedFormField: form.FormField{
				Label:           expectedLabel,
				Description:     expectedDescription,
				AllowEdit:       true,
				ValidationError: "invalid time format. Please specify in the format of `2006-01-02T15:04:05-07:00`(RFC3339)",
				Default:         "2020-01-02T03:04:05Z",
				HintType:        form.HintTypeInfo,
				Suggestions:     []string{},
			},
		},
		{
			Name:          "with valid timestamp and UTC timezone",
			Input:         "2020-01-02T00:00:00Z",
			ExpectedValue: expectedValue2,
			Dependencies:  []task.UntypedTask{inspection_task.TestInspectionTimeTaskProducer("2020-01-02T03:04:05Z"), timezoneTaskUTC},
			ExpectedFormField: form.FormField{
				Label:           expectedLabel,
				Description:     expectedDescription,
				AllowEdit:       true,
				ValidationError: "",
				Default:         "2020-01-02T03:04:05Z",
				HintType:        form.HintTypeInfo,
				Suggestions:     []string{},
			},
		},
		{
			Name:          "with valid timestamp and non UTC timezone",
			Input:         "2020-01-02T00:00:00Z",
			ExpectedValue: expectedValue2,
			Dependencies:  []task.UntypedTask{inspection_task.TestInspectionTimeTaskProducer("2020-01-02T03:04:05Z"), timezoneTaskJST},
			ExpectedFormField: form.FormField{
				Label:           expectedLabel,
				Description:     expectedDescription,
				AllowEdit:       true,
				ValidationError: "",
				Default:         "2020-01-02T12:04:05+09:00",
				HintType:        form.HintTypeInfo,
				Suggestions:     []string{},
			},
		},
	})
}

func TestInputStartTime(t *testing.T) {
	duration, err := time.ParseDuration("1h30m")
	if err != nil {
		t.Fatal(err)
	}
	endTime, err := time.Parse(time.RFC3339, "2023-01-02T15:45:00Z")
	if err != nil {
		t.Fatal(err)
	}

	ctx := inspection_task_test.WithDefaultTestInspectionTaskContext(context.Background())
	startTime, _, err := inspection_task_test.RunInspectionTask(ctx, InputStartTimeTask, inspection_task_interface.TaskModeDryRun, map[string]any{},
		task_test.NewTaskDependencyValuePair(InputDurationTaskID.GetTaskReference(), duration),
		task_test.NewTaskDependencyValuePair(InputEndTimeTaskID.GetTaskReference(), endTime),
		task_test.NewTaskDependencyValuePair(TimeZoneShiftInputTaskID.GetTaskReference(), time.UTC),
	)
	if err != nil {
		t.Errorf("unexpected error\n%v", err)
	}
	expectedTime, err := time.Parse(time.RFC3339, "2023-01-02T14:15:00Z")
	if err != nil {
		t.Errorf("unexpected error\n%v", err)
	}

	if startTime.String() != expectedTime.String() {
		t.Errorf("returned time is not matching with the expected value\n%s", startTime)
	}
}

func TestInputKindName(t *testing.T) {
	expectedDescription := ""
	expectedLabel := "Kind"
	form_task_test.TestTextForms(t, "kind", InputKindFilterTask, []*form_task_test.TextFormTestCase{
		{
			Input: "",
			ExpectedValue: &queryutil.SetFilterParseResult{
				Additives:       inputKindNameAliasMap["default"],
				Subtractives:    []string{},
				ValidationError: "",
				SubtractMode:    false,
			},
			ExpectedFormField: form.FormField{
				Label:           expectedLabel,
				Description:     expectedDescription,
				AllowEdit:       true,
				Default:         "@default",
				HintType:        form.HintTypeInfo,
				ValidationError: "kind filter can't be empty",
			},
		},
		{
			Input: "pods replicasets",
			ExpectedValue: &queryutil.SetFilterParseResult{
				Additives:       []string{"pods", "replicasets"},
				Subtractives:    []string{},
				ValidationError: "",
				SubtractMode:    false,
			},
			ExpectedFormField: form.FormField{
				Label:       expectedLabel,
				Description: expectedDescription,
				AllowEdit:   true,
				Default:     "@default",
				HintType:    form.HintTypeInfo,
			},
		},
		{
			Input: "@invalid_alias",
			ExpectedValue: &queryutil.SetFilterParseResult{
				Additives:       inputKindNameAliasMap["default"],
				Subtractives:    []string{},
				ValidationError: "",
				SubtractMode:    false,
			}, ExpectedFormField: form.FormField{
				Label:           expectedLabel,
				Description:     expectedDescription,
				AllowEdit:       true,
				ValidationError: "alias `invalid_alias` was not found",
				Default:         "@default",
				HintType:        form.HintTypeInfo,
			},
		},
	}, cmpopts.SortSlices(func(a string, b string) bool {
		return strings.Compare(a, b) > 0
	}))
}

func TestInputNamespaces(t *testing.T) {
	expectedDescription := ""
	expectedLabel := "Namespaces"
	form_task_test.TestTextForms(t, "namespaces", InputNamespaceFilterTask, []*form_task_test.TextFormTestCase{
		{
			Input: "",
			ExpectedValue: &queryutil.SetFilterParseResult{
				Additives: []string{
					"#namespaced",
					"#cluster-scoped",
				},
				Subtractives:    []string{},
				ValidationError: "",
				SubtractMode:    false,
			},
			ExpectedFormField: form.FormField{
				Label:           expectedLabel,
				Description:     expectedDescription,
				AllowEdit:       true,
				Default:         "@all_cluster_scoped @all_namespaced",
				HintType:        form.HintTypeInfo,
				ValidationError: "namespace filter can't be empty",
			},
		},
		{
			Input: "kube-system default",
			ExpectedValue: &queryutil.SetFilterParseResult{
				Additives:       []string{"kube-system", "default"},
				Subtractives:    []string{},
				ValidationError: "",
				SubtractMode:    false,
			},
			ExpectedFormField: form.FormField{
				Label:       expectedLabel,
				Description: expectedDescription,
				AllowEdit:   true,
				Default:     "@all_cluster_scoped @all_namespaced",
				HintType:    form.HintTypeInfo,
			},
		},
		{
			Input: "@all_cluster_scoped @all_namespaced",
			ExpectedValue: &queryutil.SetFilterParseResult{
				Additives:       []string{"#namespaced", "#cluster-scoped"},
				Subtractives:    []string{},
				ValidationError: "",
				SubtractMode:    false,
			}, ExpectedFormField: form.FormField{
				Label:           expectedLabel,
				Description:     expectedDescription,
				AllowEdit:       true,
				ValidationError: "",
				Default:         "@all_cluster_scoped @all_namespaced",
				HintType:        form.HintTypeInfo,
			},
		},
	}, cmpopts.SortSlices(func(a string, b string) bool {
		return strings.Compare(a, b) > 0
	}))
}

func TestNodeNameFiltertask(t *testing.T) {
	wantLabelName := "Node names"
	wantDescription := "A space-separated list of node name substrings used to collect node-related logs. If left blank, KHI gathers logs from all nodes in the cluster."
	form_task_test.TestTextForms(t, "node-name", InputNodeNameFilterTask, []*form_task_test.TextFormTestCase{
		{
			Name:          "With an empty input",
			Input:         "",
			ExpectedValue: []string{},
			Dependencies:  []task.UntypedTask{},
			ExpectedFormField: form.FormField{
				Label:       wantLabelName,
				Description: wantDescription,
				AllowEdit:   true,
				HintType:    form.HintTypeInfo,
				Default:     "",
			},
		},
		{
			Name:          "With a single node name substring",
			Input:         "node-name-1",
			ExpectedValue: []string{"node-name-1"},
			Dependencies:  []task.UntypedTask{},
			ExpectedFormField: form.FormField{
				Label:       wantLabelName,
				Description: wantDescription,
				AllowEdit:   true,
				HintType:    form.HintTypeInfo,
				Default:     "",
			},
		},
		{
			Name:          "With multiple node name substrings",
			Input:         "node-name-1 node-name-2 node-name-3",
			ExpectedValue: []string{"node-name-1", "node-name-2", "node-name-3"},
			Dependencies:  []task.UntypedTask{},
			ExpectedFormField: form.FormField{
				Label:       wantLabelName,
				Description: wantDescription,
				AllowEdit:   true,
				HintType:    form.HintTypeInfo,
				Default:     "",
			},
		},
		{
			Name:          "With invalid node name substring",
			Input:         "node-name-1 invalid=node=name node-name-3",
			ExpectedValue: []string{},
			Dependencies:  []task.UntypedTask{},
			ExpectedFormField: form.FormField{
				Label:           wantLabelName,
				Description:     wantDescription,
				AllowEdit:       true,
				HintType:        form.HintTypeInfo,
				ValidationError: "substring `invalid=node=name` is not valid as a substring of node name",
				Default:         "",
			},
		},
		{
			Name:          "With spaces around node name substring",
			Input:         "  node-name-1  node-name-2  ",
			ExpectedValue: []string{"node-name-1", "node-name-2"},
			Dependencies:  []task.UntypedTask{},
			ExpectedFormField: form.FormField{
				Label:       wantLabelName,
				Description: wantDescription,
				AllowEdit:   true,
				HintType:    form.HintTypeInfo,
				Default:     "",
			},
		},
	})
}

func TestLocationInput(t *testing.T) {
	form_task_test.TestTextForms(t, "gcp-location", InputLocationsTask, []*form_task_test.TextFormTestCase{
		{
			Name:          "With valid location",
			Input:         "asia-northeast1",
			ExpectedValue: "asia-northeast1",
			Dependencies:  []task.UntypedTask{},
			ExpectedFormField: form.FormField{
				Priority:    1,
				Id:          GCPPrefix + "input/location",
				Type:        "Text",
				Label:       "Location",
				Description: "A location(regions) containing the environments to inspect",
				HintType:    form.HintTypeInfo,
				AllowEdit:   true,
			},
		},
	})
}
