package date

import (
	"context"
	"math"
	"time"

	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

var SpecialFns map[string]values.Function

func init() {
	SpecialFns = map[string]values.Function{
		"second": values.NewFunction(
			"second",
			runtime.MustLookupBuiltinType("date", "second"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				if v1 == nil {
					return nil, errors.New(codes.FailedPrecondition, "argument t was nil")
				}
				deps := execute.GetExecutionDependencies(ctx)
				t, err := deps.ResolveTimeable(v1)
				if err != nil {
					return nil, err
				}

				return values.NewInt(int64(t.Time().Second())), nil
			}, false,
		),
		"minute": values.NewFunction(
			"minute",
			runtime.MustLookupBuiltinType("date", "minute"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				if v1 == nil {
					return nil, errors.New(codes.FailedPrecondition, "argument t was nil")
				}

				deps := execute.GetExecutionDependencies(ctx)
				t, err := deps.ResolveTimeable(v1)
				if err != nil {
					return nil, err
				}

				return values.NewInt(int64(t.Time().Minute())), nil
			}, false,
		),
		"hour": values.NewFunction(
			"hour",
			runtime.MustLookupBuiltinType("date", "hour"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				if v1 == nil {
					return nil, errors.New(codes.FailedPrecondition, "argument t was nil")
				}

				deps := execute.GetExecutionDependencies(ctx)
				t, err := deps.ResolveTimeable(v1)
				if err != nil {
					return nil, err
				}
				return values.NewInt(int64(t.Time().Hour())), nil
			}, false,
		),
		"weekDay": values.NewFunction(
			"weekDay",
			runtime.MustLookupBuiltinType("date", "weekDay"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				if v1 == nil {
					return nil, errors.New(codes.FailedPrecondition, "argument t was nil")
				}

				deps := execute.GetExecutionDependencies(ctx)
				t, err := deps.ResolveTimeable(v1)
				if err != nil {
					return nil, err
				}
				return values.NewInt(int64(t.Time().Weekday())), nil
			}, false,
		),
		"monthDay": values.NewFunction(
			"monthDay",
			runtime.MustLookupBuiltinType("date", "monthDay"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				if v1 == nil {
					return nil, errors.New(codes.FailedPrecondition, "argument t was nil")
				}

				deps := execute.GetExecutionDependencies(ctx)
				t, err := deps.ResolveTimeable(v1)
				if err != nil {
					return nil, err
				}
				return values.NewInt(int64(t.Time().Day())), nil
			}, false,
		),
		"yearDay": values.NewFunction(
			"yearDay",
			runtime.MustLookupBuiltinType("date", "yearDay"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				if v1 == nil {
					return nil, errors.New(codes.FailedPrecondition, "argument t was nil")
				}

				deps := execute.GetExecutionDependencies(ctx)
				t, err := deps.ResolveTimeable(v1)
				if err != nil {
					return nil, err
				}
				return values.NewInt(int64(t.Time().YearDay())), nil
			}, false,
		),
		"month": values.NewFunction(
			"month",
			runtime.MustLookupBuiltinType("date", "month"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				if v1 == nil {
					return nil, errors.New(codes.FailedPrecondition, "argument t was nil")
				}

				deps := execute.GetExecutionDependencies(ctx)
				t, err := deps.ResolveTimeable(v1)
				if err != nil {
					return nil, err
				}
				return values.NewInt(int64(t.Time().Month())), nil
			}, false,
		),
		"year": values.NewFunction(
			"year",
			runtime.MustLookupBuiltinType("date", "year"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				if v1 == nil {
					return nil, errors.New(codes.FailedPrecondition, "argument t was nil")
				}

				deps := execute.GetExecutionDependencies(ctx)
				t, err := deps.ResolveTimeable(v1)
				if err != nil {
					return nil, err
				}
				return values.NewInt(int64(t.Time().Year())), nil
			}, false,
		),
		"week": values.NewFunction(
			"week",
			runtime.MustLookupBuiltinType("date", "week"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				if v1 == nil {
					return nil, errors.New(codes.FailedPrecondition, "argument t was nil")
				}

				deps := execute.GetExecutionDependencies(ctx)
				t, err := deps.ResolveTimeable(v1)
				if err != nil {
					return nil, err
				}
				_, week := t.Time().ISOWeek()
				return values.NewInt(int64(week)), nil
			}, false,
		),
		"quarter": values.NewFunction(
			"quarter",
			runtime.MustLookupBuiltinType("date", "quarter"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				if v1 == nil {
					return nil, errors.New(codes.FailedPrecondition, "argument t was nil")
				}

				deps := execute.GetExecutionDependencies(ctx)
				t, err := deps.ResolveTimeable(v1)
				if err != nil {
					return nil, err
				}
				month := t.Time().Month()
				return values.NewInt(int64(math.Ceil(float64(month) / 3.0))), nil

			}, false,
		),
		"millisecond": values.NewFunction(
			"millisecond",
			runtime.MustLookupBuiltinType("date", "millisecond"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				if v1 == nil {
					return nil, errors.New(codes.FailedPrecondition, "argument t was nil")
				}

				deps := execute.GetExecutionDependencies(ctx)
				t, err := deps.ResolveTimeable(v1)
				if err != nil {
					return nil, err
				}
				millisecond := int64(time.Nanosecond) * int64(t.Time().Nanosecond()) / int64(time.Millisecond)
				return values.NewInt(millisecond), nil
			}, false,
		),
		"microsecond": values.NewFunction(
			"microsecond",
			runtime.MustLookupBuiltinType("date", "microsecond"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				if v1 == nil {
					return nil, errors.New(codes.FailedPrecondition, "argument t was nil")
				}

				deps := execute.GetExecutionDependencies(ctx)
				t, err := deps.ResolveTimeable(v1)
				if err != nil {
					return nil, err
				}
				microsecond := int64(time.Nanosecond) * int64(t.Time().Nanosecond()) / int64(time.Microsecond)
				return values.NewInt(microsecond), nil
			}, false,
		),
		"nanosecond": values.NewFunction(
			"nanosecond",
			runtime.MustLookupBuiltinType("date", "nanosecond"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v1, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				if v1 == nil {
					return nil, errors.New(codes.FailedPrecondition, "argument t was nil")
				}

				deps := execute.GetExecutionDependencies(ctx)
				t, err := deps.ResolveTimeable(v1)
				if err != nil {
					return nil, err
				}
				return values.NewInt(int64(t.Time().Nanosecond())), nil
			}, false,
		),
		"truncate": values.NewFunction(
			"truncate",
			runtime.MustLookupBuiltinType("date", "truncate"),
			func(ctx context.Context, args values.Object) (values.Value, error) {
				v, ok := args.Get("t")
				if !ok {
					return nil, errors.New(codes.Invalid, "missing argument t")
				}

				if v == nil {
					return nil, errors.New(codes.FailedPrecondition, "argument t was nil")
				}

				u, unitOk := args.Get("unit")
				if !unitOk {
					return nil, errors.New(codes.Invalid, "missing argument unit")
				}

				deps := execute.GetExecutionDependencies(ctx)
				t, err := deps.ResolveTimeable(v)
				if err != nil {
					return nil, err
				}
				w, err := execute.NewWindow(u.Duration(), u.Duration(), execute.Duration{})
				if err != nil {
					return nil, err
				}
				b := w.GetEarliestBounds(t)
				return values.NewTime(b.Start), nil
			}, false,
		),
	}

	runtime.RegisterPackageValue("date", "second", SpecialFns["second"])
	runtime.RegisterPackageValue("date", "minute", SpecialFns["minute"])
	runtime.RegisterPackageValue("date", "hour", SpecialFns["hour"])
	runtime.RegisterPackageValue("date", "weekDay", SpecialFns["weekDay"])
	runtime.RegisterPackageValue("date", "monthDay", SpecialFns["monthDay"])
	runtime.RegisterPackageValue("date", "yearDay", SpecialFns["yearDay"])
	runtime.RegisterPackageValue("date", "month", SpecialFns["month"])
	runtime.RegisterPackageValue("date", "year", SpecialFns["year"])
	runtime.RegisterPackageValue("date", "week", SpecialFns["week"])
	runtime.RegisterPackageValue("date", "quarter", SpecialFns["quarter"])
	runtime.RegisterPackageValue("date", "millisecond", SpecialFns["millisecond"])
	runtime.RegisterPackageValue("date", "microsecond", SpecialFns["microsecond"])
	runtime.RegisterPackageValue("date", "nanosecond", SpecialFns["nanosecond"])
	runtime.RegisterPackageValue("date", "truncate", SpecialFns["truncate"])
}
