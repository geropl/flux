package universe_test

import (
	"context"
	"errors"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/internal/gen"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values/valuestest"
)

func TestSchemaMutions_NewQueries(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "test rename query",
			Raw:  `from(bucket:"mybucket") |> rename(columns:{old:"new"}) |> sum()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mybucket"},
						},
					},
					{
						ID: "rename1",
						Spec: &universe.RenameOpSpec{
							Columns: map[string]string{
								"old": "new",
							},
						},
					},
					{
						ID: "sum2",
						Spec: &universe.SumOpSpec{
							SimpleAggregateConfig: execute.DefaultSimpleAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "rename1"},
					{Parent: "rename1", Child: "sum2"},
				},
			},
		},
		{
			Name: "test drop query",
			Raw:  `from(bucket:"mybucket") |> drop(columns:["col1", "col2", "col3"]) |> sum()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mybucket"},
						},
					},
					{
						ID: "drop1",
						Spec: &universe.DropOpSpec{
							Columns: []string{"col1", "col2", "col3"},
						},
					},
					{
						ID: "sum2",
						Spec: &universe.SumOpSpec{
							SimpleAggregateConfig: execute.DefaultSimpleAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "drop1"},
					{Parent: "drop1", Child: "sum2"},
				},
			},
		},
		{
			Name: "test keep query",
			Raw:  `from(bucket:"mybucket") |> keep(columns:["col1", "col2", "col3"]) |> sum()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mybucket"},
						},
					},
					{
						ID: "keep1",
						Spec: &universe.KeepOpSpec{
							Columns: []string{"col1", "col2", "col3"},
						},
					},
					{
						ID: "sum2",
						Spec: &universe.SumOpSpec{
							SimpleAggregateConfig: execute.DefaultSimpleAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "keep1"},
					{Parent: "keep1", Child: "sum2"},
				},
			},
		},
		{
			Name: "test duplicate query",
			Raw:  `from(bucket:"mybucket") |> duplicate(column: "col1", as: "col1_new") |> sum()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mybucket"},
						},
					},
					{
						ID: "duplicate1",
						Spec: &universe.DuplicateOpSpec{
							Column: "col1",
							As:     "col1_new",
						},
					},
					{
						ID: "sum2",
						Spec: &universe.SumOpSpec{
							SimpleAggregateConfig: execute.DefaultSimpleAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "duplicate1"},
					{Parent: "duplicate1", Child: "sum2"},
				},
			},
		},
		{
			Name: "test drop query fn param",
			Raw:  `from(bucket:"mybucket") |> drop(fn: (column) => column =~ /reg*/) |> sum()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mybucket"},
						},
					},
					{
						ID: "drop1",
						Spec: &universe.DropOpSpec{
							Predicate: interpreter.ResolvedFunction{
								Fn:    executetest.FunctionExpression(t, "(column) => column =~ /reg*/"),
								Scope: valuestest.Scope(),
							},
						},
					},
					{
						ID: "sum2",
						Spec: &universe.SumOpSpec{
							SimpleAggregateConfig: execute.DefaultSimpleAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "drop1"},
					{Parent: "drop1", Child: "sum2"},
				},
			},
		},
		{
			Name: "test keep query fn param",
			Raw:  `from(bucket:"mybucket") |> keep(fn: (column) => column =~ /reg*/) |> sum()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mybucket"},
						},
					},
					{
						ID: "keep1",
						Spec: &universe.KeepOpSpec{
							Predicate: interpreter.ResolvedFunction{
								Fn:    executetest.FunctionExpression(t, "(column) => column =~ /reg*/"),
								Scope: valuestest.Scope(),
							},
						},
					},
					{
						ID: "sum2",
						Spec: &universe.SumOpSpec{
							SimpleAggregateConfig: execute.DefaultSimpleAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "keep1"},
					{Parent: "keep1", Child: "sum2"},
				},
			},
		},
		{
			Name: "test rename query fn param",
			Raw:  `from(bucket:"mybucket") |> rename(fn: (column) => "new_name") |> sum()`,
			Want: &flux.Spec{
				Operations: []*flux.Operation{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mybucket"},
						},
					},
					{
						ID: "rename1",
						Spec: &universe.RenameOpSpec{
							Fn: interpreter.ResolvedFunction{
								Fn:    executetest.FunctionExpression(t, `(column) => "new_name"`),
								Scope: valuestest.Scope(),
							},
						},
					},
					{
						ID: "sum2",
						Spec: &universe.SumOpSpec{
							SimpleAggregateConfig: execute.DefaultSimpleAggregateConfig,
						},
					},
				},
				Edges: []flux.Edge{
					{Parent: "from0", Child: "rename1"},
					{Parent: "rename1", Child: "sum2"},
				},
			},
		},
		{
			Name:    "test rename query invalid",
			Raw:     `from(bucket:"mybucket") |> rename(fn: (column) => "new_name", columns: {a:"b", c:"d"}) |> sum()`,
			Want:    nil,
			WantErr: true,
		},
		{
			Name:    "test drop query invalid",
			Raw:     `from(bucket:"mybucket") |> drop(fn: (column) => column == target, columns: ["a", "b"]) |> sum()`,
			Want:    nil,
			WantErr: true,
		},
		{
			Name:    "test keep query invalid",
			Raw:     `from(bucket:"mybucket") |> keep(fn: (column) => column == target, columns: ["a", "b"]) |> sum()`,
			Want:    nil,
			WantErr: true,
		},
		{
			Name:    "test duplicate query invalid",
			Raw:     `from(bucket:"mybucket") |> duplicate(columns: ["a", "b"], n: -1) |> sum()`,
			Want:    nil,
			WantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			querytest.NewQueryTestHelper(t, tc)
		})
	}
}

func TestDropRenameKeep_Process(t *testing.T) {
	testCases := []struct {
		name    string
		spec    plan.ProcedureSpec
		data    []flux.Table
		want    []*executetest.Table
		wantErr error
	}{
		{
			name: "rename multiple cols",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.RenameOpSpec{
						Columns: map[string]string{
							"1a": "1b",
							"2a": "2b",
							"3a": "3b",
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "1b", Type: flux.TFloat},
					{Label: "2b", Type: flux.TFloat},
					{Label: "3b", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
		},

		{
			name: "drop multiple cols",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"a", "b"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
					{Label: "b", Type: flux.TFloat},
					{Label: "c", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "c", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{3.0},
					{13.0},
					{23.0},
				},
			}},
		},
		{
			name: "drop key col merge tables",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"b"},
					},
				},
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TFloat},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "two", 3.0},
						{"one", "two", 13.0},
					},
				},
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TFloat},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "three", 5.0},
						{"one", "three", 15.0},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TString},
					{Label: "c", Type: flux.TFloat},
				},
				KeyCols: []string{"a"},
				Data: [][]interface{}{
					{"one", 3.0},
					{"one", 13.0},
					{"one", 5.0},
					{"one", 15.0},
				},
			}},
		},
		{
			name: "drop key col merge error column count",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"b"},
					},
				},
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TFloat},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "two", 3.0},
						{"one", "two", 13.0},
					},
				},
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "three"},
						{"one", "three"},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TString},
					{Label: "c", Type: flux.TFloat},
				},
				KeyCols: []string{"a"},
				Data: [][]interface{}{
					{"one", 3.0},
					{"one", 13.0},
					{"one", nil},
					{"one", nil},
				},
			}},
		},
		{
			name: "drop key col merge error column type",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"b"},
					},
				},
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TFloat},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "two", 3.0},
						{"one", "two", 13.0},
					},
				},
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TString},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "three", "val"},
						{"one", "three", "val"},
					},
				},
			},
			wantErr: errors.New("schema collision detected: column \"c\" is both of type string and float"),
		},
		{
			name: "drop no exist",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"boo"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TString},
					{Label: "b", Type: flux.TString},
					{Label: "c", Type: flux.TFloat},
				},
				KeyCols: []string{"a", "b"},
				Data: [][]interface{}{
					{"one", "two", 3.0},
					{"one", "two", 13.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TString},
					{Label: "b", Type: flux.TString},
					{Label: "c", Type: flux.TFloat},
				},
				KeyCols: []string{"a", "b"},
				Data: [][]interface{}{
					{"one", "two", 3.0},
					{"one", "two", 13.0},
				},
			}},
		},
		{
			name: "keep multiple cols",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"a"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
					{Label: "b", Type: flux.TFloat},
					{Label: "c", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0},
					{11.0},
					{21.0},
				},
			}},
		},
		{
			name: "keep one key col merge tables",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"a", "c"},
					},
				},
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TFloat},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "two", 3.0},
						{"one", "two", 13.0},
					},
				},
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TFloat},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "three", 5.0},
						{"one", "three", 15.0},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TString},
					{Label: "c", Type: flux.TFloat},
				},
				KeyCols: []string{"a"},
				Data: [][]interface{}{
					{"one", 3.0},
					{"one", 13.0},
					{"one", 5.0},
					{"one", 15.0},
				},
			}},
		},
		{
			name: "keep one key col merge error column count",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"a", "c"},
					},
				},
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TFloat},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "two", 3.0},
						{"one", "two", 13.0},
					},
				},
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "three"},
						{"one", "three"},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TString},
					{Label: "c", Type: flux.TFloat},
				},
				KeyCols: []string{"a"},
				Data: [][]interface{}{
					{"one", 3.0},
					{"one", 13.0},
					{"one", nil},
					{"one", nil},
				},
			}},
		},
		{
			name: "keep one key col merge error column type",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"a", "c"},
					},
				},
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TFloat},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "two", 3.0},
						{"one", "two", 13.0},
					},
				},
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TString},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "three", "foo"},
						{"one", "three", "bar"},
					},
				},
			},
			wantErr: errors.New("schema collision detected: column \"c\" is both of type string and float"),
		},
		{
			name: "duplicate single col",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DuplicateOpSpec{
						Column: "a",
						As:     "a_1",
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
					{Label: "b", Type: flux.TFloat},
					{Label: "c", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
					{Label: "b", Type: flux.TFloat},
					{Label: "c", Type: flux.TFloat},
					{Label: "a_1", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0, 1.0},
					{11.0, 12.0, 13.0, 11.0},
					{21.0, 22.0, 23.0, 21.0},
				},
			}},
		},
		{
			name: "rename map fn (column) => name",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.RenameOpSpec{
						Fn: interpreter.ResolvedFunction{
							Fn:    executetest.FunctionExpression(t, `(column) => "new_name"`),
							Scope: valuestest.Scope(),
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			wantErr: errors.New("column 0 and 1 have the same name (\"new_name\") which is not allowed"),
		},
		{
			name: "drop predicate (column) => column ~= /reg/",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Predicate: interpreter.ResolvedFunction{
							Fn:    executetest.FunctionExpression(t, `(column) => column =~ /server*/`),
							Scope: valuestest.Scope(),
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "local", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{2.0},
					{12.0},
					{22.0},
				},
			}},
		},
		{
			name: "keep predicate (column) => column ~= /reg/",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Predicate: interpreter.ResolvedFunction{
							Fn:    executetest.FunctionExpression(t, `(column) => column =~ /server*/`),
							Scope: valuestest.Scope(),
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 3.0},
					{11.0, 13.0},
					{21.0, 23.0},
				},
			}},
		},
		{
			name: "drop and rename",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"server1", "server2"},
					},
					&universe.RenameOpSpec{
						Columns: map[string]string{
							"local": "localhost",
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "localhost", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{2.0},
					{12.0},
					{22.0},
				},
			}},
		},
		{
			name: "drop no exist",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"no_exist"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
		},
		{
			name: "rename no exist",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.RenameOpSpec{
						Columns: map[string]string{
							"no_exist": "noexist",
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want:    []*executetest.Table(nil),
			wantErr: errors.New(`rename error: column "no_exist" doesn't exist`),
		},
		{
			name: "keep no exist",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"no_exist"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{},
				Data:    [][]interface{}(nil),
			}},
		},
		{
			name: "keep no exist along with all other columns",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"no_exist", "server1", "local", "server2"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
		},
		{
			name: "duplicate no exist",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DuplicateOpSpec{
						Column: "no_exist",
						As:     "no_exist_2",
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want:    []*executetest.Table(nil),
			wantErr: errors.New(`duplicate error: column "no_exist" doesn't exist`),
		},
		{
			name: "rename group key",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.RenameOpSpec{
						Columns: map[string]string{
							"1a": "1b",
							"2a": "2b",
							"3a": "3b",
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"1a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{1.0, 12.0, 13.0},
					{1.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"1b"},
				ColMeta: []flux.ColMeta{
					{Label: "1b", Type: flux.TFloat},
					{Label: "2b", Type: flux.TFloat},
					{Label: "3b", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{1.0, 12.0, 13.0},
					{1.0, 22.0, 23.0},
				},
			}},
		},
		{
			name: "drop group key",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"2a"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"2a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 2.0, 13.0},
					{21.0, 2.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string(nil),
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 3.0},
					{11.0, 13.0},
					{21.0, 23.0},
				},
			}},
		},
		{
			name: "keep group key",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"1a"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"1a", "3a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{1.0, 12.0, 3.0},
					{1.0, 22.0, 3.0},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"1a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0},
					{1.0},
					{1.0},
				},
			}},
		},
		{
			name: "duplicate group key",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DuplicateOpSpec{
						Column: "1a",
						As:     "1a_1",
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"1a", "3a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{1.0, 12.0, 3.0},
					{1.0, 22.0, 3.0},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"1a", "3a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
					{Label: "1a_1", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0, 1.0},
					{1.0, 12.0, 3.0, 1.0},
					{1.0, 22.0, 3.0, 1.0},
				},
			}},
		},
		{
			name: "keep with changing schema",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"a"},
					},
				},
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"a"},
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TInt},
						{Label: "b", Type: flux.TFloat},
						{Label: "c", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{int64(1), 10.0, 3.0},
						{int64(1), 12.0, 4.0},
						{int64(1), 22.0, 5.0},
					},
				},
				&executetest.Table{
					KeyCols: []string{"a"},
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TInt},
						{Label: "b", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{int64(2), 11.0},
						{int64(2), 13.0},
						{int64(2), 23.0},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"a"},
					ColMeta: []flux.ColMeta{{Label: "a", Type: flux.TInt}},
					Data: [][]interface{}{
						{int64(1)},
						{int64(1)},
						{int64(1)},
					},
				},
				{
					KeyCols: []string{"a"},
					ColMeta: []flux.ColMeta{{Label: "a", Type: flux.TInt}},
					Data: [][]interface{}{
						{int64(2)},
						{int64(2)},
						{int64(2)},
					},
				},
			},
		},
		{
			name: "rename with nulls",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.RenameOpSpec{
						Columns: map[string]string{
							"1a": "1b",
							"2a": "2b",
							"3a": "3b",
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil, 2.0, 3.0},
					{11.0, 12.0, nil},
					{21.0, nil, nil},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "1b", Type: flux.TFloat},
					{Label: "2b", Type: flux.TFloat},
					{Label: "3b", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil, 2.0, 3.0},
					{11.0, 12.0, nil},
					{21.0, nil, nil},
				},
			}},
		},

		{
			name: "drop with nulls",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"a", "b"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
					{Label: "b", Type: flux.TFloat},
					{Label: "c", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil, 2.0, 3.0},
					{nil, nil, nil},
					{nil, 22.0, nil},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "c", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{3.0},
					{nil},
					{nil},
				},
			}},
		},
		{
			name: "keep with nulls",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"a"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
					{Label: "b", Type: flux.TFloat},
					{Label: "c", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, nil},
					{nil, 12.0, 13.0},
					{21.0, nil, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0},
					{nil},
					{21.0},
				},
			}},
		},
		{
			name: "duplicate with nulls",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DuplicateOpSpec{
						Column: "a",
						As:     "a_1",
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
					{Label: "b", Type: flux.TFloat},
					{Label: "c", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil, nil, 3.0},
					{nil, 12.0, nil},
					{21.0, nil, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
					{Label: "b", Type: flux.TFloat},
					{Label: "c", Type: flux.TFloat},
					{Label: "a_1", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil, nil, 3.0, nil},
					{nil, 12.0, nil, nil},
					{21.0, nil, 23.0, 21.0},
				},
			}},
		},
		{
			name: "rename group key with nulls",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.RenameOpSpec{
						Columns: map[string]string{
							"1a": "1b",
							"2a": "2b",
							"3a": "3b",
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"1a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil, 2.0, 3.0},
					{nil, 12.0, nil},
					{nil, nil, 23.0},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"1b"},
				ColMeta: []flux.ColMeta{
					{Label: "1b", Type: flux.TFloat},
					{Label: "2b", Type: flux.TFloat},
					{Label: "3b", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil, 2.0, 3.0},
					{nil, 12.0, nil},
					{nil, nil, 23.0},
				},
			}},
		},
		{
			name: "drop group key with nulls",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"2a"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"2a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, nil, 3.0},
					{nil, nil, 13.0},
					{21.0, nil, nil},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string(nil),
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 3.0},
					{nil, 13.0},
					{21.0, nil},
				},
			}},
		},
		{
			name: "keep group key with nulls",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"1a"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"1a", "3a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil, 2.0, nil},
					{nil, 12.0, nil},
					{nil, 22.0, nil},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"1a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil},
					{nil},
					{nil},
				},
			}},
		},
		{
			name: "duplicate group key with nulls",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DuplicateOpSpec{
						Column: "3a",
						As:     "3a_1",
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"1a", "3a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, nil},
					{1.0, 12.0, nil},
					{1.0, 22.0, nil},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"1a", "3a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
					{Label: "3a_1", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, nil, nil},
					{1.0, 12.0, nil, nil},
					{1.0, 22.0, nil, nil},
				},
			}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			executetest.ProcessTestHelper2(
				t,
				tc.data,
				tc.want,
				tc.wantErr,
				func(id execute.DatasetID, mem *memory.Allocator) (execute.Transformation, execute.Dataset) {
					spec := tc.spec.(*universe.SchemaMutationProcedureSpec)
					tr, d, err := universe.NewSchemaMutationTransformation(context.Background(), spec, id, mem)
					if err != nil {
						t.Fatal(err)
					}
					return tr, d
				},
			)
		})
	}
}

// TODO: determine SchemaMutationProcedureSpec pushdown/rewrite rules
/*
func TestRenameDrop_PushDown(t *testing.T) {
	m1, _ := functions.NewRenameMutator(&functions.RenameOpSpec{
		Cols: map[string]string{},
	})

	root := &plan.Procedure{
		Spec: &functions.SchemaMutationProcedureSpec{
			Mutations: []functions.SchemaMutator{m1},
		},
	}

	m2, _ := functions.NewDropKeepMutator(&functions.DropOpSpec{
		Cols: []string{},
	})

	m3, _ := functions.NewDropKeepMutator(&functions.KeepOpSpec{
		Cols: []string{},
	})

	spec := &functions.SchemaMutationProcedureSpec{
		Mutations: []functions.SchemaMutator{m2, m3},
	}

	want := &plan.Procedure{
		Spec: &functions.SchemaMutationProcedureSpec{
			Mutations: []functions.SchemaMutator{m1, m2, m3},
		},
	}
	plantest.PhysicalPlan_PushDown_TestHelper(t, spec, root, false, want)
}
*/

func BenchmarkKeep_Values(b *testing.B) {
	b.Run("1000", func(b *testing.B) {
		benchmarkSchemaMutator(b, 1000, &universe.KeepOpSpec{
			Columns: []string{"_measurement", "t0"},
		})
	})
}

func BenchmarkDrop_Values(b *testing.B) {
	b.Run("1000", func(b *testing.B) {
		benchmarkSchemaMutator(b, 1000, &universe.DropOpSpec{
			Columns: []string{"_measurement", "_field"},
		})
	})
}

func BenchmarkRename_Values(b *testing.B) {
	b.Run("1000", func(b *testing.B) {
		benchmarkSchemaMutator(b, 1000, &universe.RenameOpSpec{
			Columns: map[string]string{
				"_measurement": "m",
				"_field":       "f",
			},
		})
	})
}

func BenchmarkDuplicate_Values(b *testing.B) {
	b.Run("1000", func(b *testing.B) {
		benchmarkSchemaMutator(b, 1000, &universe.DuplicateOpSpec{
			Column: "_value",
			As:     "_prev_value",
		})
	})
}

func benchmarkSchemaMutator(b *testing.B, n int, m universe.SchemaMutation) {
	b.ReportAllocs()
	spec := &universe.SchemaMutationProcedureSpec{
		Mutations: []universe.SchemaMutation{m},
	}
	executetest.ProcessBenchmarkHelper(b,
		func(alloc *memory.Allocator) (flux.TableIterator, error) {
			schema := gen.Schema{
				NumPoints: n,
				Alloc:     alloc,
				Tags: []gen.Tag{
					{Name: "_measurement", Cardinality: 1},
					{Name: "_field", Cardinality: 6},
					{Name: "t0", Cardinality: 100},
					{Name: "t1", Cardinality: 50},
				},
				Nulls: 0.1,
			}
			return gen.Input(context.Background(), schema)
		},
		func(id execute.DatasetID, alloc *memory.Allocator) (execute.Transformation, execute.Dataset) {
			t, d, err := universe.NewSchemaMutationTransformation(context.Background(), spec, id, alloc)
			if err != nil {
				b.Fatal(err)
			}
			return t, d
		},
	)
}
